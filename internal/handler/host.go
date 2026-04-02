// Package handler contains HTTP handlers for the trivia server.
package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"trivia/internal/game"
	"trivia/internal/hub"
)

// QuizLoader is the driven port for loading a quiz from a file path into a session.
// It abstracts quiz.Loader and game.GamePort to keep QuizFull out of the handler package.
type QuizLoader interface {
	// LoadIntoSession loads the quiz at path into the given game session.
	// Returns observable quiz metadata on success.
	LoadIntoSession(path string, session game.GamePort) (QuizLoadedMeta, error)
}

// QuizLoadedMeta carries the metadata sent back to the host after a successful quiz load.
type QuizLoadedMeta struct {
	Title         string `json:"title"`
	RoundCount    int    `json:"round_count"`
	QuestionCount int    `json:"question_count"`
	PlayerURL     string `json:"player_url"`
	DisplayURL    string `json:"display_url"`
}

// HostHandler handles the quizmaster WebSocket connection.
// Auth guard is applied at the router level before reaching this handler.
type HostHandler struct {
	hub       *hub.Hub
	quizLoader QuizLoader
	baseURL   string
}

// NewHostHandler creates a HostHandler wired to the given hub, quiz loader, and base URL.
func NewHostHandler(h *hub.Hub, loader QuizLoader, baseURL string) *HostHandler {
	return &HostHandler{
		hub:        h,
		quizLoader: loader,
		baseURL:    baseURL,
	}
}

// ServeHTTP upgrades the connection to WebSocket, registers the client in RoomHost,
// and dispatches incoming host events.
func (hh *HostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := AcceptWebSocket(w, r)
	if err != nil {
		// AcceptWebSocket already wrote the HTTP error response.
		log.Printf("host: websocket upgrade failed: %v", err)
		return
	}

	client := hub.NewClient(conn, hub.RoomHost, "")
	hh.hub.Register(client)
	defer hh.hub.Deregister(client)

	session := game.NewGameSession()
	hh.readLoop(r.Context(), conn, client, session)
}

// readLoop reads incoming WebSocket messages from the host and dispatches them.
func (hh *HostHandler) readLoop(ctx context.Context, conn *websocket.Conn, client *hub.Client, session game.GamePort) {
	for {
		var raw map[string]json.RawMessage
		if err := wsjson.Read(ctx, conn, &raw); err != nil {
			// Connection closed or context cancelled — normal exit.
			return
		}

		var event string
		if eventRaw, ok := raw["event"]; ok {
			_ = json.Unmarshal(eventRaw, &event)
		}

		switch event {
		case "host_load_quiz":
			hh.handleLoadQuiz(ctx, conn, client, session, raw["payload"])
		default:
			errEvent := hub.NewErrorEvent("unknown_event", "unknown event: "+event)
			if err := hh.hub.Send(client, errEvent); err != nil {
				log.Printf("host: send error event: %v", err)
			}
		}
	}
}

// handleLoadQuiz processes a host_load_quiz event.
func (hh *HostHandler) handleLoadQuiz(ctx context.Context, conn *websocket.Conn, client *hub.Client, session game.GamePort, payloadRaw json.RawMessage) {
	var payload struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil || payload.FilePath == "" {
		errEvent := hub.NewErrorEvent("bad_request", "host_load_quiz requires file_path")
		_ = hh.hub.Send(client, errEvent)
		return
	}

	meta, err := hh.quizLoader.LoadIntoSession(payload.FilePath, session)
	if err != nil {
		errEvent := hub.NewErrorEvent("quiz_load_failed", err.Error())
		_ = hh.hub.Send(client, errEvent)
		return
	}

	meta.PlayerURL = hh.baseURL + "/play"
	meta.DisplayURL = hh.baseURL + "/display"

	response := hub.ServerEvent{
		Event:   "quiz_loaded",
		Payload: meta,
	}
	if err := hh.hub.Send(client, response); err != nil {
		log.Printf("host: send quiz_loaded: %v", err)
	}
}
