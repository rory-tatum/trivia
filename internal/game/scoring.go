package game

// RoundScores tracks the score for each team within a single round.
type RoundScores struct {
	scores map[string]int
}

// NewRoundScores creates an empty RoundScores ready for verdict application.
func NewRoundScores() *RoundScores {
	return &RoundScores{scores: make(map[string]int)}
}

// ApplyVerdict updates a team's score based on the verdict.
// A correct verdict adds one point; incorrect verdicts are no-ops.
func (r *RoundScores) ApplyVerdict(teamID string, verdict Verdict) {
	if verdict == VerdictCorrect {
		r.scores[teamID]++
	}
}

// TeamScore returns the current score for the given team.
func (r *RoundScores) TeamScore(teamID string) int {
	return r.scores[teamID]
}

// AllScores returns a copy of all team scores.
func (r *RoundScores) AllScores() map[string]int {
	result := make(map[string]int, len(r.scores))
	for k, v := range r.scores {
		result[k] = v
	}
	return result
}

// TotalScores accumulates scores across multiple rounds.
type TotalScores struct {
	totals map[string]int
}

// NewTotalScores creates an empty TotalScores tracker.
func NewTotalScores() *TotalScores {
	return &TotalScores{totals: make(map[string]int)}
}

// AddRound merges a RoundScores into the running totals.
func (t *TotalScores) AddRound(rs *RoundScores) {
	for teamID, pts := range rs.AllScores() {
		t.totals[teamID] += pts
	}
}

// TeamTotal returns the cumulative score for the given team.
func (t *TotalScores) TeamTotal(teamID string) int {
	return t.totals[teamID]
}

// AllTotals returns a copy of all cumulative team scores.
func (t *TotalScores) AllTotals() map[string]int {
	result := make(map[string]int, len(t.totals))
	for k, v := range t.totals {
		result[k] = v
	}
	return result
}
