// Package steps contains the godog acceptance test suite for the trivia game.
//
// Entry point for `go test ./tests/acceptance/trivia/steps/...`
//
// All scenarios use a real in-process HTTP + WebSocket server started via
// net/http/httptest. No mocks at the acceptance level (D3: real services).
//
// One-at-a-time implementation sequence:
//   1. Enable the walking skeleton scenario (no @skip).
//   2. Implement production code until it passes.
//   3. Commit.
//   4. Remove @skip from the next scenario, repeat.
package steps

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestAcceptance(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "trivia-acceptance",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../"},
			Tags:     "~@infrastructure", // infrastructure scenarios require Docker; run separately
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("acceptance tests failed")
	}
}

// TestAcceptanceInfrastructure runs @infrastructure-tagged scenarios.
// These require Docker to be available. Run explicitly:
//
//	go test ./tests/acceptance/trivia/steps/... -run TestAcceptanceInfrastructure
func TestAcceptanceInfrastructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping infrastructure tests in short mode")
	}

	suite := godog.TestSuite{
		Name:                "trivia-acceptance-infrastructure",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../"},
			Tags:     "@infrastructure",
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("infrastructure acceptance tests failed")
	}
}
