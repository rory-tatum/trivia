// Package steps contains step definitions for the trivia acceptance test suite.
//
// Three-layer abstraction:
//   Layer 1 (Gherkin)       -- business language in .feature files
//   Layer 2 (Step methods)  -- this file, delegates to Layer 3, no assertions here
//   Layer 3 (Test drivers)  -- trivia_driver.go, speaks the server's driving ports
//
// The World struct is the per-scenario shared context.
// It is created fresh for every scenario; state never bleeds between scenarios.
package steps

import (
	"context"
	"fmt"
	"net/http/httptest"
	"sync"
	"time"
)

// World holds all state for a single scenario execution.
// It is created in InitializeScenario and torn down after each scenario.
type World struct {
	// server is the in-process test server (started lazily on first use).
	server *httptest.Server

	// hostToken is the quizmaster authentication token for this scenario.
	hostToken string

	// quizFixtures maps filename to content for YAML files used in this scenario.
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

	// lastError holds the most recent error message received from the server.
	lastError string

	// gameSessionID is set after a successful quiz load.
	gameSessionID string

	// ctx is the base context for this scenario (cancelled in teardown).
	ctx    context.Context
	cancel context.CancelFunc
}

// QuizMeta holds observable metadata about a loaded quiz as seen on the host interface.
type QuizMeta struct {
	Title         string
	RoundCount    int
	QuestionCount int
	PlayerURL     string
	DisplayURL    string
}

// WSMessage represents a single WebSocket message (event + raw payload).
type WSMessage struct {
	Event     string
	Payload   map[string]interface{}
	Timestamp time.Time
}

// WSConnection wraps a WebSocket test connection with message collection.
type WSConnection struct {
	// Role is "host", "play", or "display".
	Role string
	// Name is the team name (play connections) or empty (host/display).
	Name string
	// Token is the persistence token stored for this player connection.
	Token string
	// Connected is true while the connection is active.
	Connected bool
	// driver reference for sending messages.
	driver *TriviaDriver
}

// newWorld creates a fresh World for a scenario.
func newWorld() *World {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	return &World{
		hostToken:        "test-secret-token",
		quizFixtures:     make(map[string]string),
		connections:      make(map[string]*WSConnection),
		receivedMessages: make(map[string][]WSMessage),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// teardown shuts down the test server and cancels the context.
func (w *World) teardown() {
	w.cancel()
	for _, conn := range w.connections {
		if conn.Connected && conn.driver != nil {
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
