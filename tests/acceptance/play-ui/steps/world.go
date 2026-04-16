// Package steps contains the acceptance test step definitions for the play-ui feature.
//
// Three-layer abstraction:
//
//	Layer 1 (Gherkin)      — business language in play_ui.feature
//	Layer 2 (Step methods) — steps.go / step_impls.go, delegates to Layer 3
//	Layer 3 (Test driver)  — driver.go, speaks the server's WebSocket driving port
//
// All tests enter the system through the play room driving port:
//
//	/ws?room=play         — play room (primary driving port for play-ui)
//	/ws?token=HOST_TOKEN  — host room (used only in Given steps to arrange state)
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

	// quizFilePaths maps filename to the absolute filesystem path where it was written.
	// Used to load the quiz into the server via the host port.
	quizFilePaths map[string]string

	// connections holds active WebSocket test connections keyed by role+name.
	// Key format: "host", "display", "play:Team Awesome", etc.
	connections map[string]*WSConnection

	// receivedMessages collects all WebSocket messages received per connection key.
	// Protected by mu.
	mu               sync.Mutex
	receivedMessages map[string][]WSMessage

	// teamIDs maps team name to the server-assigned team_id.
	teamIDs map[string]string

	// deviceTokens maps team name to the server-assigned device_token.
	deviceTokens map[string]string

	// lastError holds the most recent error event received from the server,
	// keyed by connection role+name.
	lastErrors map[string]string

	// submissionAcks tracks whether each team (by name) has received a submission_ack.
	submissionAcks map[string]bool

	// currentRoundIndex is the 0-based index of the active round (set when a round is started).
	currentRoundIndex int

	// revealedCount tracks the number of questions revealed in the current round.
	revealedCount int

	// totalQuestions is the total number of questions in the current round.
	totalQuestions int

	// lastCommandSent is the most recently sent command event name.
	lastCommandSent string

	// commandSentCount tracks how many commands of each event type were sent.
	commandSentCount map[string]int

	// ctx is the base context for this scenario (cancelled in teardown).
	ctx    context.Context
	cancel context.CancelFunc
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
	driver *PlayUIDriver
}

// newWorld creates a fresh World for a scenario.
func newWorld() *World {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	return &World{
		hostToken:         "test-secret-token",
		quizFixtures:      make(map[string]string),
		quizFilePaths:     make(map[string]string),
		connections:       make(map[string]*WSConnection),
		receivedMessages:  make(map[string][]WSMessage),
		teamIDs:           make(map[string]string),
		deviceTokens:      make(map[string]string),
		lastErrors:        make(map[string]string),
		submissionAcks:    make(map[string]bool),
		commandSentCount:  make(map[string]int),
		currentRoundIndex: -1,
		revealedCount:     0,
		totalQuestions:    0,
		ctx:               ctx,
		cancel:            cancel,
	}
}

// teardown shuts down the test server and closes all connections.
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
// It also extracts observable state from well-known server events.
func (w *World) addMessage(key string, msg WSMessage) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.receivedMessages[key] = append(w.receivedMessages[key], msg)
	if msg.Payload == nil {
		return
	}
	switch msg.Event {
	case eventTeamRegistered:
		w.captureTeamRegistered(key, msg.Payload)
	case eventRoundStarted:
		w.captureRoundStarted(msg.Payload)
	case eventQuestionRevealed:
		w.captureQuestionRevealed(msg.Payload)
	case eventError:
		if errMsg, ok := msg.Payload["message"].(string); ok {
			w.lastErrors[key] = errMsg
		}
	case eventSubmissionAck:
		w.captureSubmissionAck(key, msg.Payload)
	}
}

// captureTeamRegistered records the server-assigned team_id and device_token.
// Called under w.mu — must not lock.
func (w *World) captureTeamRegistered(key string, payload map[string]interface{}) {
	teamID, _ := payload["team_id"].(string)
	deviceToken, _ := payload["device_token"].(string)
	const playPrefix = rolePlay + ":"
	if len(key) > len(playPrefix) && key[:len(playPrefix)] == playPrefix {
		teamName := key[len(playPrefix):]
		if teamID != "" {
			w.teamIDs[teamName] = teamID
		}
		if deviceToken != "" {
			w.deviceTokens[teamName] = deviceToken
		}
	}
}

// captureRoundStarted resets round counters from a round_started payload.
// Called under w.mu — must not lock.
func (w *World) captureRoundStarted(payload map[string]interface{}) {
	if ri, ok := payload["round_index"].(float64); ok {
		w.currentRoundIndex = int(ri)
	}
	if qc, ok := payload["question_count"].(float64); ok {
		w.totalQuestions = int(qc)
	}
	w.revealedCount = 0
}

// captureQuestionRevealed increments the revealed counter.
// Called under w.mu — must not lock.
func (w *World) captureQuestionRevealed(_ map[string]interface{}) {
	w.revealedCount++
}

// captureSubmissionAck marks the team as having received a submission acknowledgement.
// Called under w.mu — must not lock.
func (w *World) captureSubmissionAck(key string, _ map[string]interface{}) {
	const playPrefix = rolePlay + ":"
	if len(key) > len(playPrefix) && key[:len(playPrefix)] == playPrefix {
		teamName := key[len(playPrefix):]
		w.submissionAcks[teamName] = true
	}
}

// teamID returns the server-assigned team_id for a team name, or empty string if unknown.
func (w *World) teamID(teamName string) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.teamIDs[teamName]
}

// deviceToken returns the server-assigned device_token for a team name.
func (w *World) deviceToken(teamName string) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.deviceTokens[teamName]
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

// pollUntil repeatedly calls check until it returns (true, nil) or the deadline elapses.
func pollUntil(deadline time.Duration, tick time.Duration, check func() (done bool, timeoutErr error)) error {
	timer := time.After(deadline)
	ticker := time.NewTicker(tick)
	defer ticker.Stop()
	for {
		select {
		case <-timer:
			_, err := check()
			if err != nil {
				return err
			}
			return fmt.Errorf("pollUntil: deadline elapsed but check returned no error")
		case <-ticker.C:
			if done, _ := check(); done {
				return nil
			}
		}
	}
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

// waitForEventCount blocks until at least count messages with the given event type
// are received on the named connection, or the deadline elapses.
func (w *World) waitForEventCount(key, eventType string, count int, deadline time.Duration) bool {
	timeout := time.After(deadline)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return false
		case <-ticker.C:
			found := 0
			for _, msg := range w.messagesFor(key) {
				if msg.Event == eventType {
					found++
				}
			}
			if found >= count {
				return true
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

// latestEvent returns the most recently received message with the given event type,
// or (WSMessage{}, false) if none has been received.
func (w *World) latestEvent(key, eventType string) (WSMessage, bool) {
	var latest WSMessage
	found := false
	for _, msg := range w.messagesFor(key) {
		if msg.Event == eventType {
			latest = msg
			found = true
		}
	}
	return latest, found
}

// countEvents counts how many messages with the given event type have been received.
func (w *World) countEvents(key, eventType string) int {
	count := 0
	for _, msg := range w.messagesFor(key) {
		if msg.Event == eventType {
			count++
		}
	}
	return count
}

// playDriver returns the PlayUIDriver for the named team's play connection.
// Panics if no play connection has been established for the team.
func (w *World) playDriver(teamName string) *PlayUIDriver {
	key := connectionKey(rolePlay, teamName)
	conn, ok := w.connections[key]
	if !ok || conn.driver == nil {
		panic(fmt.Sprintf("play connection not established for team %q — check Given steps", teamName))
	}
	return conn.driver
}

// hostDriver returns the PlayUIDriver for the host connection.
// Panics if no host connection has been established.
func (w *World) hostDriver() *PlayUIDriver {
	conn, ok := w.connections[roleHost]
	if !ok || conn.driver == nil {
		panic("host connection not established — check Given steps")
	}
	return conn.driver
}

// ensureHostDriver returns the host PlayUIDriver, creating and registering one if needed.
// Does not open a WebSocket connection — call driver.ConnectHost separately.
func (w *World) ensureHostDriver() *PlayUIDriver {
	if conn, ok := w.connections[roleHost]; ok && conn.driver != nil {
		return conn.driver
	}
	d := NewPlayUIDriver(w.server, w.hostToken, w)
	w.connections[roleHost] = &WSConnection{
		Role:      roleHost,
		Connected: false,
		driver:    d,
	}
	return d
}

// ensurePlayDriver returns the play PlayUIDriver for the given team,
// creating and registering one if needed. Does not open a WebSocket connection.
func (w *World) ensurePlayDriver(teamName string) *PlayUIDriver {
	key := connectionKey(rolePlay, teamName)
	if conn, ok := w.connections[key]; ok && conn.driver != nil {
		return conn.driver
	}
	d := NewPlayUIDriver(w.server, w.hostToken, w)
	w.connections[key] = &WSConnection{
		Role:      rolePlay,
		Name:      teamName,
		Connected: false,
		driver:    d,
	}
	return d
}

// ensureHostConnected opens the host WebSocket connection if not already open.
func (w *World) ensureHostConnected(d *PlayUIDriver) error {
	key := connectionKey(roleHost, "")
	if _, ok := d.wsConns[key]; ok {
		return nil // already connected
	}
	if err := d.ConnectHost(w.ctx); err != nil {
		return err
	}
	if conn, ok := w.connections[roleHost]; ok {
		conn.Connected = true
	}
	return nil
}

// ensurePlayConnected opens the play WebSocket connection for a team if not already open.
func (w *World) ensurePlayConnected(d *PlayUIDriver, teamName string) error {
	key := connectionKey(rolePlay, teamName)
	if _, ok := d.wsConns[key]; ok {
		return nil // already connected
	}
	if err := d.ConnectPlay(w.ctx, teamName); err != nil {
		return err
	}
	if conn, ok := w.connections[key]; ok {
		conn.Connected = true
	}
	return nil
}

// Connection role keys used as map keys.
const (
	roleHost    = "host"
	roleDisplay = "display"
	rolePlay    = "play"
)
