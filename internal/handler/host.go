// Package handler contains HTTP handlers for the trivia server.
package handler

import (
	"context"
	"encoding/json"
	"fmt"
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
	Confirmation  string `json:"confirmation"`
	SessionID     string `json:"session_id"`
}

// HostHandler handles the quizmaster WebSocket connection.
// Auth guard is applied at the router level before reaching this handler.
type HostHandler struct {
	hub        *hub.Hub
	quizLoader QuizLoader
	baseURL    string
	session    *game.GameSession
}

// NewHostHandler creates a HostHandler wired to the given hub, quiz loader, base URL, and shared game session.
func NewHostHandler(h *hub.Hub, loader QuizLoader, baseURL string, session *game.GameSession) *HostHandler {
	return &HostHandler{
		hub:        h,
		quizLoader: loader,
		baseURL:    baseURL,
		session:    session,
	}
}

// SetBaseURL updates the base URL used to build player and display URLs.
// Call this after the test server starts to inject the real address.
func (hh *HostHandler) SetBaseURL(url string) {
	hh.baseURL = url
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

	hh.readLoop(r.Context(), conn, client, hh.session)
}

// readLoop reads incoming WebSocket messages from the host and dispatches them.
func (hh *HostHandler) readLoop(ctx context.Context, conn *websocket.Conn, client *hub.Client, session *game.GameSession) {
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
		case "host_start_round":
			hh.handleStartRound(ctx, client, session, raw["payload"])
		case "host_reveal_question":
			hh.handleRevealQuestion(ctx, client, session, raw["payload"])
		case "host_end_round":
			hh.handleEndRound(ctx, client, session, raw["payload"])
		case "host_begin_scoring":
			hh.handleBeginScoring(ctx, client, session, raw["payload"])
		case "host_mark_answer":
			hh.handleMarkAnswer(ctx, client, session, raw["payload"])
		case "host_ceremony_show_question":
			hh.handleCeremonyShowQuestion(ctx, client, session, raw["payload"])
		case "host_ceremony_reveal_answer":
			hh.handleCeremonyRevealAnswer(ctx, client, session, raw["payload"])
		case "host_publish_scores":
			hh.handlePublishScores(ctx, client, session, raw["payload"])
		case "host_end_game":
			hh.handleEndGame(ctx, client, session)
		default:
			errEvent := hub.NewErrorEvent("unknown_event", "unknown event: "+event)
			if err := hh.hub.Send(client, errEvent); err != nil {
				log.Printf("host: send error event: %v", err)
			}
		}
	}
}

// handleLoadQuiz processes a host_load_quiz event.
func (hh *HostHandler) handleLoadQuiz(ctx context.Context, conn *websocket.Conn, client *hub.Client, session *game.GameSession, payloadRaw json.RawMessage) {
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
	roundWord := "rounds"
	if meta.RoundCount == 1 {
		roundWord = "round"
	}
	meta.Confirmation = fmt.Sprintf("%s | %d %s | %d questions",
		meta.Title, meta.RoundCount, roundWord, meta.QuestionCount)
	meta.SessionID = session.GetSessionID()

	response := hub.ServerEvent{
		Event:   "quiz_loaded",
		Payload: meta,
	}
	if err := hh.hub.Send(client, response); err != nil {
		log.Printf("host: send quiz_loaded: %v", err)
	}
}

func (hh *HostHandler) handleStartRound(_ context.Context, client *hub.Client, session *game.GameSession, payloadRaw json.RawMessage) {
	var payload struct {
		RoundIndex int `json:"round_index"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("bad_request", "host_start_round requires round_index"))
		return
	}
	if err := session.StartRound(payload.RoundIndex); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("start_round_failed", err.Error()))
		return
	}
	evt := hub.NewRoundStartedEvent(payload.RoundIndex)
	_ = hh.hub.Broadcast(hub.RoomHost, evt)
	_ = hh.hub.Broadcast(hub.RoomPlay, evt)
	_ = hh.hub.Broadcast(hub.RoomDisplay, evt)
}

func (hh *HostHandler) handleRevealQuestion(_ context.Context, client *hub.Client, session *game.GameSession, payloadRaw json.RawMessage) {
	var payload struct {
		RoundIndex    int `json:"round_index"`
		QuestionIndex int `json:"question_index"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("bad_request", "host_reveal_question requires round_index and question_index"))
		return
	}
	if err := session.RevealQuestion(payload.RoundIndex, payload.QuestionIndex); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("reveal_failed", err.Error()))
		return
	}
	// Get the just-revealed question (last in the revealed list).
	revealed := session.RevealedQuestions()
	if len(revealed) == 0 {
		return
	}
	q := revealed[len(revealed)-1]
	totalQuestions := session.Quiz().QuestionCount
	evt := hub.NewQuestionRevealedEvent(q, len(revealed), totalQuestions)
	_ = hh.hub.Broadcast(hub.RoomHost, evt)
	_ = hh.hub.Broadcast(hub.RoomPlay, evt)
	_ = hh.hub.Broadcast(hub.RoomDisplay, evt)
}

func (hh *HostHandler) handleEndRound(_ context.Context, client *hub.Client, session *game.GameSession, payloadRaw json.RawMessage) {
	var payload struct {
		RoundIndex int `json:"round_index"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("bad_request", "host_end_round requires round_index"))
		return
	}
	if err := session.ForceEndRound(payload.RoundIndex); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("end_round_failed", err.Error()))
	}
}

func (hh *HostHandler) handleBeginScoring(_ context.Context, client *hub.Client, session *game.GameSession, payloadRaw json.RawMessage) {
	var payload struct {
		RoundIndex int `json:"round_index"`
	}
	_ = json.Unmarshal(payloadRaw, &payload)

	if err := session.BeginScoring(); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("begin_scoring_failed", err.Error()))
		return
	}
	evt := hub.NewScoringOpenedEvent(payload.RoundIndex)
	_ = hh.hub.Broadcast(hub.RoomHost, evt)
	_ = hh.hub.Broadcast(hub.RoomPlay, evt)
	_ = hh.hub.Broadcast(hub.RoomDisplay, evt)
}

func (hh *HostHandler) handleMarkAnswer(_ context.Context, client *hub.Client, session *game.GameSession, payloadRaw json.RawMessage) {
	var payload struct {
		TeamID        string `json:"team_id"`
		RoundIndex    int    `json:"round_index"`
		QuestionIndex int    `json:"question_index"`
		Verdict       string `json:"verdict"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil || payload.TeamID == "" {
		_ = hh.hub.Send(client, hub.NewErrorEvent("bad_request", "host_mark_answer requires team_id and verdict"))
		return
	}
	verdict := game.Verdict(payload.Verdict)
	if err := session.MarkAnswerVerdict(payload.TeamID, payload.RoundIndex, payload.QuestionIndex, verdict); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("mark_answer_failed", err.Error()))
	}
}

func (hh *HostHandler) handleCeremonyShowQuestion(_ context.Context, client *hub.Client, session *game.GameSession, payloadRaw json.RawMessage) {
	var payload struct {
		QuestionIndex int `json:"question_index"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("bad_request", "host_ceremony_show_question requires question_index"))
		return
	}

	roundIndex := session.CurrentRoundIndex()

	// Transition to ceremony on first call (questionIndex == 0), or advance.
	if payload.QuestionIndex == 0 {
		if err := session.StartCeremony(); err != nil {
			_ = hh.hub.Send(client, hub.NewErrorEvent("ceremony_failed", err.Error()))
			return
		}
	} else {
		if err := session.AdvanceCeremony(payload.QuestionIndex); err != nil {
			_ = hh.hub.Send(client, hub.NewErrorEvent("ceremony_failed", err.Error()))
			return
		}
	}

	q := session.CeremonyQuestion(roundIndex, payload.QuestionIndex)
	evt := hub.NewCeremonyQuestionShownEvent(payload.QuestionIndex, q)
	_ = hh.hub.Broadcast(hub.RoomDisplay, evt)
	_ = hh.hub.Broadcast(hub.RoomPlay, evt)
}

func (hh *HostHandler) handleCeremonyRevealAnswer(_ context.Context, client *hub.Client, session *game.GameSession, payloadRaw json.RawMessage) {
	var payload struct {
		QuestionIndex int `json:"question_index"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("bad_request", "host_ceremony_reveal_answer requires question_index"))
		return
	}

	roundIndex := session.CurrentRoundIndex()
	answer := session.CeremonyAnswer(roundIndex, payload.QuestionIndex)
	evt := hub.NewCeremonyAnswerRevealedEvent(payload.QuestionIndex, answer)
	// Answer revealed only to display room (not play room — boundary rule).
	_ = hh.hub.Broadcast(hub.RoomDisplay, evt)
}

func (hh *HostHandler) handlePublishScores(_ context.Context, client *hub.Client, session *game.GameSession, payloadRaw json.RawMessage) {
	var payload struct {
		RoundIndex int `json:"round_index"`
	}
	_ = json.Unmarshal(payloadRaw, &payload)

	if err := session.PublishRoundScores(payload.RoundIndex); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("publish_scores_failed", err.Error()))
		return
	}
	scores := session.RoundScores(payload.RoundIndex)
	evt := hub.NewRoundScoresPublishedEvent(payload.RoundIndex, scores)
	_ = hh.hub.Broadcast(hub.RoomHost, evt)
	_ = hh.hub.Broadcast(hub.RoomPlay, evt)
	_ = hh.hub.Broadcast(hub.RoomDisplay, evt)
}

func (hh *HostHandler) handleEndGame(_ context.Context, client *hub.Client, session *game.GameSession) {
	if err := session.EndGame(); err != nil {
		_ = hh.hub.Send(client, hub.NewErrorEvent("end_game_failed", err.Error()))
		return
	}
	finalScores := session.FinalScores()
	evt := hub.NewGameOverEvent(finalScores)
	_ = hh.hub.Broadcast(hub.RoomHost, evt)
	_ = hh.hub.Broadcast(hub.RoomPlay, evt)
	_ = hh.hub.Broadcast(hub.RoomDisplay, evt)
}
