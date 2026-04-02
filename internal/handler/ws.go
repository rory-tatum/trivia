package handler

import (
	"net/http"

	"nhooyr.io/websocket"
)

// AcceptWebSocket upgrades the HTTP connection to a WebSocket connection.
// Returns the accepted connection or writes an error response and returns nil on failure.
func AcceptWebSocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // allow cross-origin for test clients
	})
	if err != nil {
		return nil, err
	}
	return conn, nil
}
