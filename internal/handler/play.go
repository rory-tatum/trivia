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

// PlayHandler handles WebSocket connections from player clients.
type PlayHandler struct {
	h       *hub.Hub
	session game.GamePort
	reader  game.StateReader
}

// NewPlayHandler creates a PlayHandler wired to the given hub and game session.
func NewPlayHandler(h *hub.Hub, session game.GamePort, reader game.StateReader) *PlayHandler {
	return &PlayHandler{h: h, session: session, reader: reader}
}

// ServeHTTP upgrades the connection, registers the client in RoomPlay,
// sends a state snapshot, and dispatches incoming player events.
func (ph *PlayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := AcceptWebSocket(w, r)
	if err != nil {
		log.Printf("play: websocket upgrade failed: %v", err)
		return
	}

	client := hub.NewClient(conn, hub.RoomPlay, "")
	ph.h.Register(client)
	defer ph.h.Deregister(client)

	// Send state snapshot to the newly connected client.
	snapshot := hub.NewStateSnapshotEvent(hub.StateSnapshotPayload{
		State:             ph.reader.CurrentState(),
		Quiz:              ph.reader.Quiz(),
		Teams:             ph.reader.TeamRegistry(),
		CurrentRound:      ph.reader.CurrentRoundIndex(),
		RevealedQuestions: ph.reader.RevealedQuestions(),
	})
	if err := ph.h.Send(client, snapshot); err != nil {
		log.Printf("play: send state_snapshot: %v", err)
	}

	ph.readLoop(r.Context(), conn, client)
}

// teamNameByID looks up the display name for the given team ID in the registry.
// Returns id itself as a fallback when the team is not found.
func (ph *PlayHandler) teamNameByID(id string) string {
	for _, t := range ph.reader.TeamRegistry() {
		if t.ID == id {
			return t.Name
		}
	}
	return id
}

// teamIDByName looks up the team ID for the given display name in the registry.
// Returns empty string when no team with that name is found.
func (ph *PlayHandler) teamIDByName(name string) string {
	for _, t := range ph.reader.TeamRegistry() {
		if t.Name == name {
			return t.ID
		}
	}
	return ""
}

// readLoop reads incoming WebSocket messages from the player and dispatches them.
func (ph *PlayHandler) readLoop(ctx context.Context, conn *websocket.Conn, client *hub.Client) {
	for {
		var raw map[string]json.RawMessage
		if err := wsjson.Read(ctx, conn, &raw); err != nil {
			return
		}

		var event string
		if eventRaw, ok := raw["event"]; ok {
			_ = json.Unmarshal(eventRaw, &event)
		}

		switch event {
		case "team_register":
			ph.handleTeamRegister(ctx, conn, client, raw["payload"])
		case "team_rejoin":
			ph.handleTeamRejoin(ctx, client, raw["payload"])
		case "draft_answer":
			// Fire-and-forget: store the draft, no response required.
			ph.handleDraftAnswer(ctx, raw["payload"])
		case "submit_answers":
			ph.handleSubmitAnswers(ctx, conn, client, raw["payload"])
		default:
			errEvent := hub.NewErrorEvent("unknown_event", "unknown play event: "+event)
			if err := ph.h.Send(client, errEvent); err != nil {
				log.Printf("play: send error event: %v", err)
			}
		}
	}
}

func (ph *PlayHandler) handleTeamRegister(ctx context.Context, conn *websocket.Conn, client *hub.Client, payloadRaw json.RawMessage) {
	var payload struct {
		TeamName string `json:"team_name"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil || payload.TeamName == "" {
		_ = ph.h.Send(client, hub.NewErrorEvent("bad_request", "team_register requires team_name"))
		return
	}

	team, err := ph.session.RegisterTeam(payload.TeamName)
	if err != nil {
		_ = ph.h.Send(client, hub.NewErrorEvent("team_name_taken", err.Error()))
		return
	}

	// Send team_registered directly to the registering client.
	resp := hub.ServerEvent{
		Event: "team_registered",
		Payload: map[string]interface{}{
			"team_id":      team.ID,
			"device_token": team.DeviceToken,
		},
	}
	if err := ph.h.Send(client, resp); err != nil {
		log.Printf("play: send team_registered: %v", err)
	}

	// Broadcast team_joined to the host room.
	_ = ph.h.Broadcast(hub.RoomHost, hub.NewTeamJoinedEvent(team.ID, team.Name))
}

func (ph *PlayHandler) handleTeamRejoin(_ context.Context, client *hub.Client, payloadRaw json.RawMessage) {
	var payload struct {
		TeamID      string `json:"team_id"`
		DeviceToken string `json:"device_token"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil || payload.TeamID == "" || payload.DeviceToken == "" {
		_ = ph.h.Send(client, hub.NewErrorEvent("bad_request", "team_rejoin requires team_id and device_token"))
		return
	}

	if !ph.session.ValidateTeamToken(payload.TeamID, payload.DeviceToken) {
		_ = ph.h.Send(client, hub.NewErrorEvent("invalid_token", "device token does not match"))
		return
	}

	snapshot := hub.NewStateSnapshotEvent(hub.StateSnapshotPayload{
		State:             ph.reader.CurrentState(),
		Quiz:              ph.reader.Quiz(),
		Teams:             ph.reader.TeamRegistry(),
		CurrentRound:      ph.reader.CurrentRoundIndex(),
		RevealedQuestions: ph.reader.RevealedQuestions(),
		DraftAnswers:      ph.reader.GetAllDrafts(payload.TeamID),
	})
	if err := ph.h.Send(client, snapshot); err != nil {
		log.Printf("play: send state_snapshot on rejoin: %v", err)
	}
}

func (ph *PlayHandler) handleDraftAnswer(_ context.Context, payloadRaw json.RawMessage) {
	var payload struct {
		TeamID        string `json:"team_id"`
		TeamName      string `json:"team_name"`
		RoundIndex    int    `json:"round_index"`
		QuestionIndex int    `json:"question_index"`
		Answer        string `json:"answer"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return // malformed payload — fire-and-forget, ignore silently
	}
	// Accept team_name as the team identifier when team_id is absent
	// (the driver sends team_name; the session uses team-N IDs).
	teamID := payload.TeamID
	if teamID == "" {
		teamID = ph.teamIDByName(payload.TeamName)
	}
	if teamID == "" {
		return // unknown team — silently ignore
	}
	_ = ph.session.SaveDraft(teamID, payload.RoundIndex, payload.QuestionIndex, payload.Answer)
}

func (ph *PlayHandler) handleSubmitAnswers(_ context.Context, conn *websocket.Conn, client *hub.Client, payloadRaw json.RawMessage) {
	var payload struct {
		TeamID     string `json:"team_id"`
		RoundIndex int    `json:"round_index"`
		Answers    []struct {
			QuestionIndex int    `json:"question_index"`
			Answer        string `json:"answer"`
		} `json:"answers"`
	}
	if err := json.Unmarshal(payloadRaw, &payload); err != nil || payload.TeamID == "" {
		_ = ph.h.Send(client, hub.NewErrorEvent("bad_request", "submit_answers requires team_id and answers"))
		return
	}

	submissions := make([]game.Submission, len(payload.Answers))
	for i, a := range payload.Answers {
		submissions[i] = game.Submission{
			TeamID:        payload.TeamID,
			RoundIndex:    payload.RoundIndex,
			QuestionIndex: a.QuestionIndex,
			Answer:        a.Answer,
		}
	}

	if err := ph.session.SubmitAnswers(payload.TeamID, payload.RoundIndex, submissions); err != nil {
		_ = ph.h.Send(client, hub.NewErrorEvent("submit_failed", err.Error()))
		return
	}

	// Acknowledge submission to the submitting client.
	ack := hub.NewSubmissionAckEvent(payload.TeamID, payload.RoundIndex, true)
	if err := ph.h.Send(client, ack); err != nil {
		log.Printf("play: send submission_ack: %v", err)
	}

	// Notify host of received submission.
	teamName := ph.teamNameByID(payload.TeamID)
	_ = ph.h.Broadcast(hub.RoomHost, hub.NewSubmissionReceivedEvent(payload.TeamID, teamName, payload.RoundIndex))
}
