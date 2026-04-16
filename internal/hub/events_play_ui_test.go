package hub_test

// Tests for hub event payload changes required by play-ui step 01-01:
//   - CeremonyAnswerRevealedPayload.Verdicts field is included in JSON serialization
//   - RoundScoresPayload.Scores is a typed slice ([]ScoreEntry) not map[string]int

import (
	"encoding/json"
	"testing"

	"trivia/internal/hub"
)

func TestCeremonyAnswerRevealedPayload_IncludesVerdicts(t *testing.T) {
	verdicts := []hub.TeamVerdict{
		{TeamID: "team-1", TeamName: "Alpha", Verdict: "correct"},
		{TeamID: "team-2", TeamName: "Beta", Verdict: "incorrect"},
	}
	evt := hub.NewCeremonyAnswerRevealedEvent(0, "Paris", verdicts)

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)
	payload, _ := result["payload"].(map[string]interface{})
	if payload == nil {
		t.Fatal("expected payload in marshaled event")
	}
	v, ok := payload["verdicts"]
	if !ok {
		t.Fatal("expected 'verdicts' key in ceremony_answer_revealed payload")
	}
	list, ok := v.([]interface{})
	if !ok || len(list) != 2 {
		t.Fatalf("expected 2 verdicts in list, got %v", v)
	}
}

func TestRoundScoresPayload_ScoresIsTypedSlice(t *testing.T) {
	entries := []hub.ScoreEntry{
		{TeamID: "team-1", TeamName: "Alpha", RoundScore: 2, RunningTotal: 2},
		{TeamID: "team-2", TeamName: "Beta", RoundScore: 0, RunningTotal: 0},
	}
	evt := hub.NewRoundScoresPublishedEvent(0, entries)

	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result map[string]interface{}
	_ = json.Unmarshal(data, &result)
	payload, _ := result["payload"].(map[string]interface{})
	scores, ok := payload["scores"].([]interface{})
	if !ok || len(scores) != 2 {
		t.Fatalf("expected scores as array of 2 entries, got %v", payload["scores"])
	}
	first, _ := scores[0].(map[string]interface{})
	if first["team_name"] != "Alpha" {
		t.Errorf("expected first entry team_name %q, got %v", "Alpha", first["team_name"])
	}
}
