// driver.go is the Layer 3 test driver for the play-ui acceptance tests.
//
// It speaks the server's WebSocket driving ports exclusively.
// Play-room commands are sent through /ws?room=play.
// Host commands that set up game state are sent through /ws?token=HOST_TOKEN.
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

// PlayUIDriver drives the trivia server through its public WebSocket driving ports.
// The primary driving port for play-ui tests is /ws?room=play.
// The host driving port /ws?token=HOST_TOKEN is used only to arrange game state
// preconditions (Given steps) — it is not the subject under test.
type PlayUIDriver struct {
	server    *httptest.Server
	hostToken string
	world     *World

	// wsConns holds open WebSocket connections keyed by role+name.
	wsConns map[string]*websocket.Conn
}

// NewPlayUIDriver creates a driver wired to the given test server.
func NewPlayUIDriver(server *httptest.Server, hostToken string, world *World) *PlayUIDriver {
	return &PlayUIDriver{
		server:    server,
		hostToken: hostToken,
		world:     world,
		wsConns:   make(map[string]*websocket.Conn),
	}
}

// wsURL converts the server HTTP URL to a WebSocket URL.
func (d *PlayUIDriver) wsURL(path string) string {
	return strings.Replace(d.server.URL, "http://", "ws://", 1) + path
}

// -- Connection management ---------------------------------------------------

// ConnectPlay opens a WebSocket connection for a named team to the play room (/ws?room=play).
// This is the primary driving port for play-ui acceptance tests.
func (d *PlayUIDriver) ConnectPlay(ctx context.Context, teamName string) error {
	conn, _, err := websocket.Dial(ctx, d.wsURL("/ws?room=play"), nil)
	if err != nil {
		return fmt.Errorf("play ws connect for %q: %w", teamName, err)
	}
	key := connectionKey(rolePlay, teamName)
	d.wsConns[key] = conn
	go d.readLoop(ctx, key, conn)
	return nil
}

// ConnectHost opens a WebSocket connection to the host room.
// Used only to arrange game state preconditions in Given steps.
func (d *PlayUIDriver) ConnectHost(ctx context.Context) error {
	url := d.wsURL(fmt.Sprintf("/ws?token=%s", d.hostToken))
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("host ws connect: %w", err)
	}
	key := connectionKey(roleHost, "")
	d.wsConns[key] = conn
	go d.readLoop(ctx, key, conn)
	return nil
}

// ConnectDisplay opens a WebSocket connection to the display room.
// Used where scenarios require a display client for ceremony verification.
func (d *PlayUIDriver) ConnectDisplay(ctx context.Context) error {
	conn, _, err := websocket.Dial(ctx, d.wsURL("/ws?room=display"), nil)
	if err != nil {
		return fmt.Errorf("display ws connect: %w", err)
	}
	key := connectionKey(roleDisplay, "")
	d.wsConns[key] = conn
	go d.readLoop(ctx, key, conn)
	return nil
}

// ConnectPlayWithNewConn opens a fresh WebSocket connection for a team to the play
// room, simulating a device reconnect. The old connection (if any) is left as-is
// (simulating a stale connection that the server will eventually close).
func (d *PlayUIDriver) ConnectPlayWithNewConn(ctx context.Context, teamName string) error {
	conn, _, err := websocket.Dial(ctx, d.wsURL("/ws?room=play"), nil)
	if err != nil {
		return fmt.Errorf("play ws reconnect for %q: %w", teamName, err)
	}
	key := connectionKey(rolePlay, teamName) + ":reconnect"
	d.wsConns[key] = conn
	go d.readLoop(ctx, key, conn)
	return nil
}

// CloseConnection closes a named WebSocket connection.
func (d *PlayUIDriver) CloseConnection(role, name string) {
	key := connectionKey(role, name)
	if conn, ok := d.wsConns[key]; ok {
		conn.Close(websocket.StatusNormalClosure, "test teardown")
		delete(d.wsConns, key)
	}
}

// readLoop receives messages from a WebSocket connection and adds them to the world.
func (d *PlayUIDriver) readLoop(ctx context.Context, key string, conn *websocket.Conn) {
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
func (d *PlayUIDriver) sendMessage(ctx context.Context, connKey string, msg map[string]interface{}) error {
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

// -- Play room commands (driving port: /ws?room=play) ------------------------

// PlayRegisterTeam sends the team_register command for a player.
func (d *PlayUIDriver) PlayRegisterTeam(ctx context.Context, teamName string) error {
	return d.sendMessage(ctx, connectionKey(rolePlay, teamName), map[string]interface{}{
		"event": "team_register",
		"payload": map[string]interface{}{
			"team_name": teamName,
		},
	})
}

// PlayRejoinTeam sends the team_rejoin command using stored team identity.
func (d *PlayUIDriver) PlayRejoinTeam(ctx context.Context, teamName, teamID, deviceToken string) error {
	return d.sendMessage(ctx, connectionKey(rolePlay, teamName), map[string]interface{}{
		"event": "team_rejoin",
		"payload": map[string]interface{}{
			"team_id":      teamID,
			"device_token": deviceToken,
		},
	})
}

// PlayRejoinTeamWithBadToken sends the team_rejoin command with an unrecognised token.
func (d *PlayUIDriver) PlayRejoinTeamWithBadToken(ctx context.Context, teamName string) error {
	return d.sendMessage(ctx, connectionKey(rolePlay, teamName), map[string]interface{}{
		"event": "team_rejoin",
		"payload": map[string]interface{}{
			"team_id":      "00000000-0000-0000-0000-000000000000",
			"device_token": "00000000-0000-0000-0000-000000000000",
		},
	})
}

// PlayDraftAnswer sends a draft_answer command for a player.
func (d *PlayUIDriver) PlayDraftAnswer(ctx context.Context, teamName string, roundIndex, questionIndex int, answer string) error {
	return d.sendMessage(ctx, connectionKey(rolePlay, teamName), map[string]interface{}{
		"event": "draft_answer",
		"payload": map[string]interface{}{
			"team_name":      teamName,
			"round_index":    roundIndex,
			"question_index": questionIndex,
			"answer":         answer,
		},
	})
}

// PlaySubmitAnswers sends the submit_answers command for a player.
func (d *PlayUIDriver) PlaySubmitAnswers(ctx context.Context, teamName string, roundIndex int, answers []map[string]interface{}) error {
	teamID := d.world.teamID(teamName)
	return d.sendMessage(ctx, connectionKey(rolePlay, teamName), map[string]interface{}{
		"event": "submit_answers",
		"payload": map[string]interface{}{
			"team_id":     teamID,
			"round_index": roundIndex,
			"answers":     answers,
		},
	})
}

// PlaySubmitAnswersWithID sends submit_answers using an explicit team_id (for error scenarios).
func (d *PlayUIDriver) PlaySubmitAnswersWithID(ctx context.Context, teamName, teamID string, roundIndex int, answers []map[string]interface{}) error {
	return d.sendMessage(ctx, connectionKey(rolePlay, teamName), map[string]interface{}{
		"event": "submit_answers",
		"payload": map[string]interface{}{
			"team_id":     teamID,
			"round_index": roundIndex,
			"answers":     answers,
		},
	})
}

// -- Host commands (driving port: /ws?token=HOST_TOKEN) — for Given steps only

// HostLoadQuiz sends the host_load_quiz command.
func (d *PlayUIDriver) HostLoadQuiz(ctx context.Context, filePath string) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
		"event": "host_load_quiz",
		"payload": map[string]interface{}{
			"file_path": filePath,
		},
	})
}

// HostStartRound sends the host_start_round command.
func (d *PlayUIDriver) HostStartRound(ctx context.Context, roundIndex int) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
		"event": "host_start_round",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostRevealQuestion sends the host_reveal_question command.
func (d *PlayUIDriver) HostRevealQuestion(ctx context.Context, roundIndex, questionIndex int) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
		"event": "host_reveal_question",
		"payload": map[string]interface{}{
			"round_index":    roundIndex,
			"question_index": questionIndex,
		},
	})
}

// HostEndRound sends the host_end_round command.
func (d *PlayUIDriver) HostEndRound(ctx context.Context, roundIndex int) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
		"event": "host_end_round",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostBeginScoring sends the host_begin_scoring command.
func (d *PlayUIDriver) HostBeginScoring(ctx context.Context, roundIndex int) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
		"event": "host_begin_scoring",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostMarkAnswer sends the host_mark_answer command.
func (d *PlayUIDriver) HostMarkAnswer(ctx context.Context, teamID string, roundIndex, questionIndex int, verdict string) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
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
func (d *PlayUIDriver) HostPublishScores(ctx context.Context, roundIndex int) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
		"event": "host_publish_scores",
		"payload": map[string]interface{}{
			"round_index": roundIndex,
		},
	})
}

// HostCeremonyShowQuestion sends the host_ceremony_show_question command.
func (d *PlayUIDriver) HostCeremonyShowQuestion(ctx context.Context, questionIndex int) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
		"event": "host_ceremony_show_question",
		"payload": map[string]interface{}{
			"question_index": questionIndex,
		},
	})
}

// HostCeremonyRevealAnswer sends the host_ceremony_reveal_answer command.
func (d *PlayUIDriver) HostCeremonyRevealAnswer(ctx context.Context, questionIndex int) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
		"event": "host_ceremony_reveal_answer",
		"payload": map[string]interface{}{
			"question_index": questionIndex,
		},
	})
}

// HostEndGame sends the host_end_game command.
func (d *PlayUIDriver) HostEndGame(ctx context.Context) error {
	return d.sendMessage(ctx, connectionKey(roleHost, ""), map[string]interface{}{
		"event":   "host_end_game",
		"payload": map[string]interface{}{},
	})
}

// -- Quiz fixture helpers ----------------------------------------------------

// WriteQuizFixture writes a YAML quiz fixture to a temp directory and returns the path.
func (d *PlayUIDriver) WriteQuizFixture(filename, content string) (string, error) {
	dir, err := os.MkdirTemp("", "play-ui-quiz-*")
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

// MultipleChoiceQuizYAML generates a quiz with one multiple choice question.
func MultipleChoiceQuizYAML(title string) string {
	return fmt.Sprintf(`title: %q
rounds:
  - name: "Round 1"
    questions:
      - text: "Which planet is closest to the sun?"
        answer: "Mercury"
        choices:
          - "Venus"
          - "Mercury"
          - "Mars"
          - "Earth"
`, title)
}

// MultiPartQuizYAML generates a quiz with one multi-part question.
func MultiPartQuizYAML(title string) string {
	return fmt.Sprintf(`title: %q
rounds:
  - name: "Round 1"
    questions:
      - text: "Name the three primary colors."
        answers:
          - "Red"
          - "Blue"
          - "Yellow"
`, title)
}

// MediaQuizYAML generates a quiz with one image question.
func MediaQuizYAML(title string) string {
	return fmt.Sprintf(`title: %q
rounds:
  - name: "Round 1"
    questions:
      - text: "Name this landmark."
        answer: "Eiffel Tower"
        media:
          type: "image"
          url: "/media/eiffel.jpg"
`, title)
}

// QuizQuestion is a simple question fixture for building test quiz content.
type QuizQuestion struct {
	Text   string
	Answer string
}

// TitleFromFilename derives a human-readable quiz title from a filename by:
// stripping the ".yaml" extension, splitting on "-", title-casing each part.
func TitleFromFilename(filename string) string {
	name := strings.TrimSuffix(filename, ".yaml")
	parts := strings.Split(name, "-")
	titleParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		titleParts = append(titleParts, strings.ToUpper(part[:1])+part[1:])
	}
	return strings.Join(titleParts, " ")
}
