package handler

import (
	"context"
	"log"
	"net/http"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"trivia/internal/game"
	"trivia/internal/hub"
)

// DisplayHandler handles WebSocket connections from display screen clients.
type DisplayHandler struct {
	h      *hub.Hub
	reader game.StateReader
}

// NewDisplayHandler creates a DisplayHandler wired to the given hub and state reader.
func NewDisplayHandler(h *hub.Hub, reader game.StateReader) *DisplayHandler {
	return &DisplayHandler{h: h, reader: reader}
}

// ServeHTTP upgrades the connection, registers the client in RoomDisplay,
// sends a state snapshot, and holds the connection open to receive broadcasts.
func (dh *DisplayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := AcceptWebSocket(w, r)
	if err != nil {
		log.Printf("display: websocket upgrade failed: %v", err)
		return
	}

	client := hub.NewClient(conn, hub.RoomDisplay, "")
	dh.h.Register(client)
	defer dh.h.Deregister(client)

	// Send state snapshot to the newly connected display.
	snapshot := hub.NewStateSnapshotEvent(hub.StateSnapshotPayload{
		State:             dh.reader.CurrentState(),
		Quiz:              dh.reader.Quiz(),
		Teams:             dh.reader.TeamRegistry(),
		CurrentRound:      dh.reader.CurrentRoundIndex(),
		RevealedQuestions: dh.reader.RevealedQuestions(),
	})
	if err := dh.h.Send(client, snapshot); err != nil {
		log.Printf("display: send state_snapshot: %v", err)
	}

	receiveLoop(r.Context(), conn)
}

// receiveLoop reads and discards incoming messages, keeping the connection
// alive until the client disconnects or the context is cancelled.
func receiveLoop(ctx context.Context, conn *websocket.Conn) {
	for {
		var msg interface{}
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			return
		}
		// Display clients are receive-only; incoming messages are silently discarded.
	}
}
