package hub

import (
	"nhooyr.io/websocket"
)

// Room identifies which of the three named rooms a client belongs to.
type Room string

const (
	// RoomHost is for the quizmaster connection.
	RoomHost Room = "host"
	// RoomPlay is for player connections.
	RoomPlay Room = "play"
	// RoomDisplay is for the display screen connection.
	RoomDisplay Room = "display"
)

// Client wraps a WebSocket connection and its room assignment.
// It holds the nhooyr.io/websocket Conn, a room, and an optional team_id.
type Client struct {
	conn   *websocket.Conn
	room   Room
	teamID string
}

// NewClient creates a Client assigned to the given room.
// teamID is empty for host and display connections.
func NewClient(conn *websocket.Conn, room Room, teamID string) *Client {
	return &Client{conn: conn, room: room, teamID: teamID}
}

// Room returns the room this client is assigned to.
func (c *Client) Room() Room { return c.room }

// TeamID returns the team ID for play-room clients (empty for host/display).
func (c *Client) TeamID() string { return c.teamID }

// Conn returns the underlying WebSocket connection.
func (c *Client) Conn() *websocket.Conn { return c.conn }
