// Package hub manages WebSocket connections, rooms, and event broadcasting.
package hub

import (
	"context"
	"fmt"
	"sync"

	"nhooyr.io/websocket/wsjson"
)

// Hub manages WebSocket clients across the three named rooms.
// It is goroutine-safe: all mutations are protected by mu.
type Hub struct {
	mu      sync.RWMutex
	rooms   map[Room]map[*Client]struct{}
}

// NewHub creates an initialised Hub with three empty rooms.
func NewHub() *Hub {
	h := &Hub{
		rooms: make(map[Room]map[*Client]struct{}),
	}
	h.rooms[RoomHost] = make(map[*Client]struct{})
	h.rooms[RoomPlay] = make(map[*Client]struct{})
	h.rooms[RoomDisplay] = make(map[*Client]struct{})
	return h
}

// Register adds c to its assigned room.
// Safe to call concurrently.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.rooms[c.room]; !ok {
		h.rooms[c.room] = make(map[*Client]struct{})
	}
	h.rooms[c.room][c] = struct{}{}
}

// Deregister removes c from its assigned room.
// Safe to call concurrently.
func (h *Hub) Deregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if members, ok := h.rooms[c.room]; ok {
		delete(members, c)
	}
}

// Broadcast sends msg as JSON to all clients currently in room.
// Errors from individual clients are collected but do not abort delivery to others.
// Returns a combined error if any sends failed.
func (h *Hub) Broadcast(room Room, msg interface{}) error {
	h.mu.RLock()
	members := h.rooms[room]
	clients := make([]*Client, 0, len(members))
	for c := range members {
		clients = append(clients, c)
	}
	h.mu.RUnlock()

	var errs []error
	for _, c := range clients {
		ctx := context.Background()
		if err := wsjson.Write(ctx, c.conn, msg); err != nil {
			errs = append(errs, fmt.Errorf("client %p: %w", c, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("broadcast to %q: %d error(s): %v", room, len(errs), errs)
	}
	return nil
}

// Send delivers msg as JSON to a single client only.
// Used for error events and submission_ack.
func (h *Hub) Send(c *Client, msg interface{}) error {
	ctx := context.Background()
	return wsjson.Write(ctx, c.conn, msg)
}

// ClientCount returns the number of clients currently in room.
// Safe to call concurrently.
func (h *Hub) ClientCount(room Room) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.rooms[room])
}
