// Package steps contains the acceptance test step definitions for the host-ui feature.
//
// Three-layer abstraction:
//
//	Layer 1 (Gherkin)      — business language in host_ui.feature
//	Layer 2 (Step methods) — steps.go / step_impls.go, delegates to Layer 3
//	Layer 3 (Test driver)  — driver.go, speaks the server's WebSocket driving port
//
// All tests enter the system through one of two driving ports:
//
//	/ws?token=HOST_TOKEN  — host room (quizmaster commands)
//	/ws?room=play         — play room (team connections)
//	/ws?room=display      — display room (ceremony/answer display)
//
// No internal packages are imported in the driver. Black-box boundary enforced.
package steps

import (
	"context"
	"fmt"
	"net/http/httptest"
	"sync"
	"time"
)

// World holds all state for a single scenario execution.
// It is created fresh for every scenario; no state bleeds between scenarios.
type World struct {
	// server is the in-process test server (started lazily on first use).
	server *httptest.Server

	// hostToken is the authentication token used for the quizmaster connection.
	hostToken string

	// quizFixtures maps filename to YAML content for quiz files used in this scenario.
	quizFixtures map[string]string

	// connections holds active WebSocket test connections keyed by role+name.
	// Key format: "host", "display", "play:Team Awesome", etc.
	connections map[string]*WSConnection

	// receivedMessages collects all WebSocket messages received per connection key.
	// Protected by mu.
	mu               sync.Mutex
	receivedMessages map[string][]WSMessage

	// lastQuizMeta holds the metadata returned after a successful quiz load.
	lastQuizMeta QuizMeta

	// lastError holds the most recent error event received from the server.
	lastError string

	// teamIDs maps team name to the server-assigned team_id.
	teamIDs map[string]string

	// quizLoaded tracks whether the quiz has been loaded into the session.
	quizLoaded bool

	// lastCommandSent is the most recently sent host command event name.
	// Used to verify no spurious commands were sent (e.g. empty-path guard).
	lastCommandSent string

	// commandSentCount tracks how many commands of each event type were sent.
	commandSentCount map[string]int

	// lastConnectError holds the error returned by a failed ConnectHostWithToken call.
	lastConnectError error

	// authFailed is true when a connection attempt was refused due to an invalid token.
	authFailed bool

	// authErrorMessage is the human-readable message for the auth failure.
	authErrorMessage string

	// reconnectAttemptCount tracks how many reconnect attempts were made after auth failure.
	reconnectAttemptCount int

	// connectionStatus tracks the observable protocol state: "connecting", "connected",
	// "reconnecting", or "disconnected". Set by When steps that drive connection lifecycle.
	connectionStatus string

	// connectionDropped is true when the host connection was force-closed by the driver.
	connectionDropped bool

	// reconnectExhausted is true when the WsClient RECONNECT_FAILED event is modelled:
	// 10 consecutive close events without a successful handshake have been simulated.
	reconnectExhausted bool

	// reconnectFailureCount tracks the number of consecutive reconnect failures simulated.
	reconnectFailureCount int

	// currentRoundIndex is the 0-based index of the active round (set when a round is started).
	// A value >= 0 indicates a round is in progress; -1 means no round has started.
	currentRoundIndex int

	// ctx is the base context for this scenario (cancelled in teardown).
	ctx    context.Context
	cancel context.CancelFunc
}

// QuizMeta holds observable metadata about a loaded quiz as seen on the host panel.
type QuizMeta struct {
	Title         string
	RoundCount    int
	QuestionCount int
	PlayerURL     string
	DisplayURL    string
	Confirmation  string
}

// WSMessage represents a single WebSocket message received from the server.
type WSMessage struct {
	Event     string
	Payload   map[string]interface{}
	Timestamp time.Time
}

// WSConnection wraps a WebSocket test connection.
type WSConnection struct {
	// Role is "host", "play", or "display".
	Role string
	// Name is the team name for play connections; empty otherwise.
	Name string
	// Connected is true while the connection is active.
	Connected bool
	// driver is the test driver used to send messages on this connection.
	driver *HostUIDriver
}

// newWorld creates a fresh World for a scenario.
func newWorld() *World {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	return &World{
		hostToken:         "test-secret-token",
		quizFixtures:     make(map[string]string),
		connections:      make(map[string]*WSConnection),
		receivedMessages: make(map[string][]WSMessage),
		teamIDs:          make(map[string]string),
		commandSentCount: make(map[string]int),
		connectionStatus: "connecting",
		currentRoundIndex: -1,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// teardown shuts down the test server and closes all connections.
func (w *World) teardown() {
	w.cancel()
	for role, conn := range w.connections {
		if conn.Connected && conn.driver != nil {
			_ = role
			conn.driver.CloseConnection(conn.Role, conn.Name)
		}
	}
	if w.server != nil {
		w.server.Close()
	}
}

// connectionKey returns the map key for a connection by role and name.
func connectionKey(role, name string) string {
	if name == "" {
		return role
	}
	return fmt.Sprintf("%s:%s", role, name)
}

// addMessage appends a received message for a connection key (thread-safe).
func (w *World) addMessage(key string, msg WSMessage) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.receivedMessages[key] = append(w.receivedMessages[key], msg)
	// Capture team_id from team_registered events.
	if msg.Event == "team_registered" && msg.Payload != nil {
		if teamID, ok := msg.Payload["team_id"].(string); ok && teamID != "" {
			if len(key) > 5 && key[:5] == "play:" {
				teamName := key[5:]
				w.teamIDs[teamName] = teamID
			}
		}
	}
	// Capture quiz metadata from quiz_loaded events.
	if msg.Event == "quiz_loaded" && msg.Payload != nil {
		if conf, ok := msg.Payload["confirmation"].(string); ok {
			w.lastQuizMeta.Confirmation = conf
		}
		if title, ok := msg.Payload["title"].(string); ok {
			w.lastQuizMeta.Title = title
		}
		if playerURL, ok := msg.Payload["player_url"].(string); ok {
			w.lastQuizMeta.PlayerURL = playerURL
		}
		if displayURL, ok := msg.Payload["display_url"].(string); ok {
			w.lastQuizMeta.DisplayURL = displayURL
		}
		if rc, ok := msg.Payload["round_count"].(float64); ok {
			w.lastQuizMeta.RoundCount = int(rc)
		}
		if qc, ok := msg.Payload["question_count"].(float64); ok {
			w.lastQuizMeta.QuestionCount = int(qc)
		}
	}
	// Capture the last error message.
	if msg.Event == "error" && msg.Payload != nil {
		if errMsg, ok := msg.Payload["message"].(string); ok {
			w.lastError = errMsg
		}
	}
}

// teamID returns the server-assigned team_id for a team name, or the name itself as fallback.
func (w *World) teamID(teamName string) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if id, ok := w.teamIDs[teamName]; ok {
		return id
	}
	return teamName
}

// messagesFor returns a snapshot of received messages for a connection key.
func (w *World) messagesFor(key string) []WSMessage {
	w.mu.Lock()
	defer w.mu.Unlock()
	msgs := w.receivedMessages[key]
	result := make([]WSMessage, len(msgs))
	copy(result, msgs)
	return result
}

// waitForEvent blocks until a message with the given event type is received
// on the named connection, or the deadline elapses.
func (w *World) waitForEvent(key, eventType string, deadline time.Duration) (WSMessage, bool) {
	timeout := time.After(deadline)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return WSMessage{}, false
		case <-ticker.C:
			for _, msg := range w.messagesFor(key) {
				if msg.Event == eventType {
					return msg, true
				}
			}
		}
	}
}

// hasReceivedEvent returns true if the connection has received any message with the given event type.
func (w *World) hasReceivedEvent(key, eventType string) bool {
	for _, msg := range w.messagesFor(key) {
		if msg.Event == eventType {
			return true
		}
	}
	return false
}

// hostDriver returns the TriviaDriver for the host connection.
// Panics if no host connection has been established — precondition violated.
func (w *World) hostDriver() *HostUIDriver {
	conn, ok := w.connections["host"]
	if !ok || conn.driver == nil {
		panic("host connection not established — check Given steps")
	}
	return conn.driver
}
