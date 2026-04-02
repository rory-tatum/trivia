// TriviaDriver is the Layer 3 test driver.
//
// It speaks the server's driving ports (WebSocket events and HTTP endpoints)
// exclusively. Step methods in steps.go delegate here.
//
// All production behavior is exercised through:
//   - HTTP: GET /host, GET /play, GET /display, GET /media/*
//   - WebSocket: /ws (with ?token= for host room)
//
// No internal packages are imported. The driver is a pure black-box client.
package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// TriviaDriver drives the real trivia server through its public HTTP and WebSocket ports.
type TriviaDriver struct {
	server    *httptest.Server
	hostToken string
	world     *World

	// wsConns holds open WebSocket connections keyed by role+name.
	wsConns map[string]*websocket.Conn
}

// NewTriviaDriver creates a driver wired to the given test server.
func NewTriviaDriver(server *httptest.Server, hostToken string, world *World) *TriviaDriver {
	return &TriviaDriver{
		server:    server,
		hostToken: hostToken,
		world:     world,
		wsConns:   make(map[string]*websocket.Conn),
	}
}

// -- HTTP helpers ----------------------------------------------------------

// GetPlayerURL returns the HTTP URL for the /play interface.
func (d *TriviaDriver) GetPlayerURL() string {
	return d.server.URL + "/play"
}

// GetDisplayURL returns the HTTP URL for the /display interface.
func (d *TriviaDriver) GetDisplayURL() string {
	return d.server.URL + "/display"
}

// GetHostURL returns the HTTP URL for the /host interface (includes token).
func (d *TriviaDriver) GetHostURL() string {
	return fmt.Sprintf("%s/host?token=%s", d.server.URL, d.hostToken)
}

// GetHostURLWithToken returns the host URL with the given token (for negative tests).
func (d *TriviaDriver) GetHostURLWithToken(token string) string {
	return fmt.Sprintf("%s/host?token=%s", d.server.URL, token)
}

// GetHostURLNoToken returns the host URL without a token.
func (d *TriviaDriver) GetHostURLNoToken() string {
	return d.server.URL + "/host"
}

// HTTPGet performs a GET request and returns the status code and body.
func (d *TriviaDriver) HTTPGet(url string) (int, string, error) {
	resp, err := http.Get(url) //nolint:noctx
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), nil
}

// -- WebSocket helpers -------------------------------------------------------

// wsURL converts the server HTTP URL to a WebSocket URL.
func (d *TriviaDriver) wsURL(path string) string {
	return strings.Replace(d.server.URL, "http://", "ws://", 1) + path
}

// ConnectHost opens a WebSocket connection to the host room.
// Returns an error if the connection is refused (e.g. invalid token).
func (d *TriviaDriver) ConnectHost(ctx context.Context) error {
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

// ConnectDisplay opens a WebSocket connection to the display room.
func (d *TriviaDriver) ConnectDisplay(ctx context.Context) error {
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
func (d *TriviaDriver) ConnectPlay(ctx context.Context, teamName string) error {
	conn, _, err := websocket.Dial(ctx, d.wsURL("/ws?room=play"), nil)
	if err != nil {
		return fmt.Errorf("play ws connect for %q: %w", teamName, err)
	}
	key := connectionKey("play", teamName)
	d.wsConns[key] = conn
	go d.readLoop(ctx, key, conn)
	return nil
}

// CloseConnection closes the named WebSocket connection.
func (d *TriviaDriver) CloseConnection(role, name string) {
	key := connectionKey(role, name)
	if conn, ok := d.wsConns[key]; ok {
		conn.Close(websocket.StatusNormalClosure, "test teardown")
		delete(d.wsConns, key)
	}
}

// readLoop receives messages from a WebSocket connection and adds them to the world.
func (d *TriviaDriver) readLoop(ctx context.Context, key string, conn *websocket.Conn) {
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

// -- Driving port: host events -----------------------------------------------

// SendMessage sends a JSON message through the named WebSocket connection.
func (d *TriviaDriver) SendMessage(ctx context.Context, connKey string, msg map[string]interface{}) error {
	conn, ok := d.wsConns[connKey]
	if !ok {
		return fmt.Errorf("no connection for key %q", connKey)
	}
	return wsjson.Write(ctx, conn, msg)
}

// HostLoadQuiz sends the host_load_quiz event through the host connection.
func (d *TriviaDriver) HostLoadQuiz(ctx context.Context, filePath string) error {
	return d.SendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_load_quiz",
		"payload": map[string]interface{}{
			"file_path": filePath,
		},
	})
}

// HostStartRound sends the host_start_round event.
func (d *TriviaDriver) HostStartRound(ctx context.Context, roundIndex int) error {
	return d.SendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_start_round",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostRevealQuestion sends the host_reveal_question event.
func (d *TriviaDriver) HostRevealQuestion(ctx context.Context, roundIndex, questionIndex int) error {
	return d.SendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_reveal_question",
		"payload": map[string]interface{}{
			"round_index":    roundIndex,
			"question_index": questionIndex,
		},
	})
}

// HostEndRound sends the host_end_round event.
func (d *TriviaDriver) HostEndRound(ctx context.Context, roundIndex int) error {
	return d.SendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_end_round",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostMarkAnswer sends the host_mark_answer event.
func (d *TriviaDriver) HostMarkAnswer(ctx context.Context, teamID string, roundIndex, questionIndex int, verdict string) error {
	return d.SendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_mark_answer",
		"payload": map[string]interface{}{
			"team_id":        teamID,
			"round_index":    roundIndex,
			"question_index": questionIndex,
			"verdict":        verdict,
		},
	})
}

// HostCeremonyShowQuestion sends the host_ceremony_show_question event.
func (d *TriviaDriver) HostCeremonyShowQuestion(ctx context.Context, questionIndex int) error {
	return d.SendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_ceremony_show_question",
		"payload": map[string]interface{}{
			"question_index": questionIndex,
		},
	})
}

// HostCeremonyRevealAnswer sends the host_ceremony_reveal_answer event.
func (d *TriviaDriver) HostCeremonyRevealAnswer(ctx context.Context, questionIndex int) error {
	return d.SendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_ceremony_reveal_answer",
		"payload": map[string]interface{}{
			"question_index": questionIndex,
		},
	})
}

// HostPublishScores sends the host_publish_scores event.
func (d *TriviaDriver) HostPublishScores(ctx context.Context, roundIndex int) error {
	return d.SendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event": "host_publish_scores",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostEndGame sends the host_end_game event.
func (d *TriviaDriver) HostEndGame(ctx context.Context) error {
	return d.SendMessage(ctx, connectionKey("host", ""), map[string]interface{}{
		"event":   "host_end_game",
		"payload": map[string]interface{}{},
	})
}

// -- Driving port: play events -----------------------------------------------

// PlayRegisterTeam sends the team_register event for a player connection.
func (d *TriviaDriver) PlayRegisterTeam(ctx context.Context, teamName string) error {
	return d.SendMessage(ctx, connectionKey("play", teamName), map[string]interface{}{
		"event": "team_register",
		"payload": map[string]interface{}{
			"team_name": teamName,
		},
	})
}

// PlayRejoinTeam sends the team_rejoin event.
func (d *TriviaDriver) PlayRejoinTeam(ctx context.Context, teamName, teamID, deviceToken string) error {
	return d.SendMessage(ctx, connectionKey("play", teamName), map[string]interface{}{
		"event": "team_rejoin",
		"payload": map[string]interface{}{
			"team_id":      teamID,
			"device_token": deviceToken,
		},
	})
}

// PlayDraftAnswer sends the draft_answer event.
func (d *TriviaDriver) PlayDraftAnswer(ctx context.Context, teamName string, roundIndex, questionIndex int, answer string) error {
	return d.SendMessage(ctx, connectionKey("play", teamName), map[string]interface{}{
		"event": "draft_answer",
		"payload": map[string]interface{}{
			"team_name":      teamName,
			"round_index":    roundIndex,
			"question_index": questionIndex,
			"answer":         answer,
		},
	})
}

// PlaySubmitAnswers sends the submit_answers event for a team.
func (d *TriviaDriver) PlaySubmitAnswers(ctx context.Context, teamName string, roundIndex int, answers []map[string]interface{}) error {
	return d.SendMessage(ctx, connectionKey("play", teamName), map[string]interface{}{
		"event": "submit_answers",
		"payload": map[string]interface{}{
			"team_name":   teamName,
			"round_index": roundIndex,
			"answers":     answers,
		},
	})
}

// -- Quiz fixture helpers -----------------------------------------------------

// WriteQuizFixture writes a YAML quiz fixture to a temp directory and returns the path.
func (d *TriviaDriver) WriteQuizFixture(filename, content string) (string, error) {
	dir, err := os.MkdirTemp("", "trivia-quiz-*")
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return "", err
	}
	return path, nil
}

// SimpleQuizYAML generates a minimal valid YAML quiz with the given title, questions per round.
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

// MultiRoundQuizYAML generates a YAML quiz with the given number of rounds,
// distributing totalQuestions evenly across rounds (adding extras to last round).
func MultiRoundQuizYAML(title string, rounds, totalQuestions int) string {
	b := &strings.Builder{}
	fmt.Fprintf(b, "title: %q\n", title)
	b.WriteString("rounds:\n")
	perRound := totalQuestions / rounds
	if perRound < 1 {
		perRound = 1
	}
	q := 0
	for r := 0; r < rounds; r++ {
		fmt.Fprintf(b, "  - name: %q\n", fmt.Sprintf("Round %d", r+1))
		b.WriteString("    questions:\n")
		count := perRound
		if r == rounds-1 {
			// Last round gets remaining questions.
			count = totalQuestions - q
			if count < 1 {
				count = 1
			}
		}
		for i := 0; i < count; i++ {
			q++
			fmt.Fprintf(b, "      - text: %q\n", fmt.Sprintf("Question %d?", q))
			fmt.Fprintf(b, "        answer: %q\n", fmt.Sprintf("Answer %d", q))
		}
	}
	return b.String()
}

// QuizQuestion is a simple question fixture.
type QuizQuestion struct {
	Text   string
	Answer string
}

// ExtractStringField safely extracts a string field from a map.
func ExtractStringField(m map[string]interface{}, key string) (string, bool) {
	v, ok := m[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// PayloadContainsAnswerField returns true if the payload contains
// a field named "answer" or "answers" at any nesting level.
// Used to verify the answer-boundary invariant.
func PayloadContainsAnswerField(payload map[string]interface{}) bool {
	return containsAnswerFieldRecursive(payload)
}

func containsAnswerFieldRecursive(m map[string]interface{}) bool {
	for k, v := range m {
		if k == "answer" || k == "answers" {
			return true
		}
		if nested, ok := v.(map[string]interface{}); ok {
			if containsAnswerFieldRecursive(nested) {
				return true
			}
		}
	}
	return false
}

// MarshalJSON is a convenience helper for debug output.
func MarshalJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
