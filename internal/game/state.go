package game

import "fmt"

// GameState represents the current phase of a game session.
type GameState string

const (
	StateLobby       GameState = "LOBBY"
	StateRoundActive GameState = "ROUND_ACTIVE"
	StateRoundEnded  GameState = "ROUND_ENDED"
	StateScoring     GameState = "SCORING"
	StateCeremony    GameState = "CEREMONY"
	StateRoundScores GameState = "ROUND_SCORES"
	StateGameOver    GameState = "GAME_OVER"
)

// validTransitions defines all permitted state machine transitions.
var validTransitions = map[GameState][]GameState{
	StateLobby:       {StateRoundActive},
	StateRoundActive: {StateRoundEnded},
	StateRoundEnded:  {StateScoring},
	StateScoring:     {StateCeremony},
	StateCeremony:    {StateRoundScores},
	StateRoundScores: {StateRoundActive, StateGameOver},
	StateGameOver:    {},
}

// ValidateTransition returns nil if transitioning from -> to is permitted,
// or an error describing the invalid transition.
func ValidateTransition(from, to GameState) error {
	allowed, ok := validTransitions[from]
	if !ok {
		return fmt.Errorf("unknown state %q", from)
	}
	for _, s := range allowed {
		if s == to {
			return nil
		}
	}
	return fmt.Errorf("invalid transition: %q -> %q", from, to)
}
