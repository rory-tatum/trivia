// HostUIDriver is the Layer 3 test driver for the host-ui acceptance tests.
//
// It speaks the server's WebSocket driving ports exclusively.
// All host commands are sent through /ws?token=HOST_TOKEN.
// No internal packages are imported — this is a pure black-box client.
package steps

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// HostUIDriver drives the trivia server through its public WebSocket driving port.
type HostUIDriver struct {
	server    *httptest.Server
	hostToken string
	world     *World

	// wsConns holds open WebSocket connections keyed by role+name.
	wsConns map[string]*websocket.Conn
}

// NewHostUIDriver creates a driver wired to the given test server.
func NewHostUIDriver(server *httptest.Server, hostToken string, world *World) *HostUIDriver {
	return &HostUIDriver{
		server:    server,
		hostToken: hostToken,
		world:     world,
		wsConns:   make(map[string]*websocket.Conn),
	}
}

// wsURL converts the server HTTP URL to a WebSocket URL.
func (d *HostUIDriver) wsURL(path string) string {
	return strings.Replace(d.server.URL, "http://", "ws://", 1) + path
}

// ConnectHost opens a WebSocket connection to the host room with the driver's token.
// Returns an error if the connection is refused (e.g. wrong token → HTTP 403).
func (d *HostUIDriver) ConnectHost(ctx context.Context) error {
	url := d.wsURL(fmt.Sprintf("/ws?token=%s", d.hostToken))
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("host ws connect: %w", err)
	}
	key := connectionKey("host", "")
	d.wsConns[key] = conn
	go d.readLoop(ctx, key, conn)
	return nil
}

// ConnectHostWithToken opens a WebSocket connection using an explicit token.
// For wrong-token scenarios the server returns a non-101 response and Dial fails;
// the caller is expected to handle the error as an auth rejection.
func (d *HostUIDriver) ConnectHostWithToken(ctx context.Context, token string) error {
	url := d.wsURL(fmt.Sprintf("/ws?token=%s", token))
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("host ws connect with token %q: %w", token, err)
	}
	key := connectionKey("host", "")
	d.wsConns[key] = conn
	go d.readLoop(ctx, key, conn)
	return nil
}

// ConnectDisplay opens a WebSocket connection to the display room.
func (d *HostUIDriver) ConnectDisplay(ctx context.Context) error {
	conn, _, err := websocket.Dial(ctx, d.wsURL("/ws?room=display"), nil)
	if err != nil {
		return fmt.Errorf("display ws connect: %w", err)
	}
	key := connectionKey("display", "")
	d.wsConns[key] = conn
	go d.readLoop(ctx, key, conn)
	return nil
}

// ConnectPlay opens a WebSocket connection for a player.
func (d *HostUIDriver) ConnectPlay(ctx context.Context, teamName string) error {
	conn, _, err := websocket.Dial(ctx, d.wsURL("/ws?room=play"), nil)
	if err != nil {
		return fmt.Errorf("play ws connect for %q: %w", teamName, err)
	}
	key := connectionKey("play", teamName)
	d.wsConns[key] = conn
	go d.readLoop(ctx, key, conn)
	return nil
}

// CloseConnection closes a named WebSocket connection.
func (d *HostUIDriver) CloseConnection(role, name string) {
	key := connectionKey(role, name)
	if conn, ok := d.wsConns[key]; ok {
		conn.Close(websocket.StatusNormalClosure, "test teardown")
		delete(d.wsConns, key)
	}
}

// DropHostConnection force-closes the host WebSocket with StatusGoingAway to simulate
// an unexpected mid-game network drop (as opposed to a clean shutdown).
func (d *HostUIDriver) DropHostConnection(ctx context.Context) {
	key := connectionKey("host", "")
	if conn, ok := d.wsConns[key]; ok {
		conn.Close(websocket.StatusGoingAway, "network drop simulation")
		delete(d.wsConns, key)
	}
}

// ReconnectHost re-dials the host WebSocket after a drop.
// Equivalent to ConnectHost but makes reconnect intent explicit in the test driver.
func (d *HostUIDriver) ReconnectHost(ctx context.Context) error {
	return d.ConnectHost(ctx)
}

// readLoop receives messages from a WebSocket connection and adds them to the world.
func (d *HostUIDriver) readLoop(ctx context.Context, key string, conn *websocket.Conn) {
	for {
		var raw map[string]interface{}
		err := wsjson.Read(ctx, conn, &raw)
		if err != nil {
			return
		}
		event, _ := raw["event"].(string)
		payload, _ := raw["payload"].(map[string]interface{})
		d.world.addMessage(key, WSMessage{
			Event:     event,
			Payload:   payload,
			Timestamp: time.Now(),
		})
	}
}

// sendMessage sends a JSON message through the named WebSocket connection.
func (d *HostUIDriver) sendMessage(ctx context.Context, connKey string, msg map[string]interface{}) error {
	conn, ok := d.wsConns[connKey]
	if !ok {
		return fmt.Errorf("no connection for key %q", connKey)
	}
	if event, ok := msg["event"].(string); ok {
		d.world.commandSentCount[event]++
		d.world.lastCommandSent = event
	}
	return wsjson.Write(ctx, conn, msg)
}

// -- Host commands (driving port: /ws?token=HOST_TOKEN) ----------------------

// HostLoadQuiz sends the host_load_quiz command.
func (d *HostUIDriver) HostLoadQuiz(ctx context.Context, filePath string) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_load_quiz",
		"payload": map[string]interface{}{
			"file_path": filePath,
		},
	})
}

// HostStartRound sends the host_start_round command.
func (d *HostUIDriver) HostStartRound(ctx context.Context, roundIndex int) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_start_round",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostRevealQuestion sends the host_reveal_question command.
func (d *HostUIDriver) HostRevealQuestion(ctx context.Context, roundIndex, questionIndex int) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_reveal_question",
		"payload": map[string]interface{}{
			"round_index":    roundIndex,
			"question_index": questionIndex,
		},
	})
}

// HostEndRound sends the host_end_round command.
func (d *HostUIDriver) HostEndRound(ctx context.Context, roundIndex int) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_end_round",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostBeginScoring sends the host_begin_scoring command.
func (d *HostUIDriver) HostBeginScoring(ctx context.Context, roundIndex int) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_begin_scoring",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostMarkAnswer sends the host_mark_answer command.
func (d *HostUIDriver) HostMarkAnswer(ctx context.Context, teamID string, roundIndex, questionIndex int, verdict string) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_mark_answer",
		"payload": map[string]interface{}{
			"team_id":        teamID,
			"round_index":    roundIndex,
			"question_index": questionIndex,
			"verdict":        verdict,
		},
	})
}

// HostPublishScores sends the host_publish_scores command.
func (d *HostUIDriver) HostPublishScores(ctx context.Context, roundIndex int) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_publish_scores",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostCeremonyShowQuestion sends the host_ceremony_show_question command.
func (d *HostUIDriver) HostCeremonyShowQuestion(ctx context.Context, questionIndex int) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_ceremony_show_question",
		"payload": map[string]interface{}{
			"question_index": questionIndex,
		},
	})
}

// HostCeremonyRevealAnswer sends the host_ceremony_reveal_answer command.
func (d *HostUIDriver) HostCeremonyRevealAnswer(ctx context.Context, questionIndex int) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_ceremony_reveal_answer",
		"payload": map[string]interface{}{
			"question_index": questionIndex,
		},
	})
}

// HostEndGame sends the host_end_game command.
func (d *HostUIDriver) HostEndGame(ctx context.Context) error {
	return d.sendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event":   "host_end_game",
		"payload": map[string]interface{}{},
	})
}

// -- Play commands -----------------------------------------------------------

// PlayRegisterTeam sends the team_register command for a player.
func (d *HostUIDriver) PlayRegisterTeam(ctx context.Context, teamName string) error {
	return d.sendMessage(ctx, connectionKey("play", teamName), map[string]interface{}{
		"event": "team_register",
		"payload": map[string]interface{}{
			"team_name": teamName,
		},
	})
}

// PlayDraftAnswer sends a draft_answer command for a player.
func (d *HostUIDriver) PlayDraftAnswer(ctx context.Context, teamName string, roundIndex, questionIndex int, answer string) error {
	return d.sendMessage(ctx, connectionKey("play", teamName), map[string]interface{}{
		"event": "draft_answer",
		"payload": map[string]interface{}{
			"round_index":    roundIndex,
			"question_index": questionIndex,
			"answer":         answer,
		},
	})
}

// PlaySubmitAnswers sends the submit_answers command for a player.
func (d *HostUIDriver) PlaySubmitAnswers(ctx context.Context, teamName string, roundIndex int, answers []map[string]interface{}) error {
	return d.sendMessage(ctx, connectionKey("play", teamName), map[string]interface{}{
		"event": "submit_answers",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
			"answers":     answers,
		},
	})
}

// -- Quiz fixture helpers ----------------------------------------------------

// WriteQuizFixture writes a YAML quiz fixture to a temp directory and returns the path.
func (d *HostUIDriver) WriteQuizFixture(filename, content string) (string, error) {
	dir, err := os.MkdirTemp("", "host-ui-quiz-*")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", err
	}
	return path, nil
}

// SimpleQuizYAML generates a minimal single-round quiz YAML fixture.
func SimpleQuizYAML(title string, questions []QuizQuestion) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "title: %q\n", title)
	b.WriteString("rounds:\n")
	b.WriteString("  - name: \"Round 1\"\n")
	b.WriteString("    questions:\n")
	for _, q := range questions {
		fmt.Fprintf(b, "      - text: %q\n", q.Text)
		fmt.Fprintf(b, "        answer: %q\n", q.Answer)
	}
	return b.String()
}

// MultiRoundQuizYAML generates a multi-round quiz YAML fixture.
func MultiRoundQuizYAML(title string, rounds, questionsPerRound int) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "title: %q\n", title)
	b.WriteString("rounds:\n")
	for r := 0; r < rounds; r++ {
		fmt.Fprintf(b, "  - name: %q\n", fmt.Sprintf("Round %d", r+1))
		b.WriteString("    questions:\n")
		for i := 0; i < questionsPerRound; i++ {
			qNum := r*questionsPerRound + i + 1
			fmt.Fprintf(b, "      - text: %q\n", fmt.Sprintf("Question %d?", qNum))
			fmt.Fprintf(b, "        answer: %q\n", fmt.Sprintf("Answer %d", qNum))
		}
	}
	return b.String()
}

// QuizQuestion is a simple question fixture for building test quiz content.
type QuizQuestion struct {
	Text   string
	Answer string
}
