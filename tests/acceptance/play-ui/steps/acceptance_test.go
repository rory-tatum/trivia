// acceptance_test.go is the test entry point for the play-ui acceptance suite.
//
// Run all non-skipped scenarios:
//
//	go test ./tests/acceptance/play-ui/steps/...
//
// Run only the walking skeleton:
//
//	go test ./tests/acceptance/play-ui/steps/... -run TestAcceptanceWalkingSkeleton
//
// Run real-io adapter integration tests:
//
//	go test ./tests/acceptance/play-ui/steps/... -run TestAcceptanceAdapterIntegration
package steps

import (
	"testing"

	"github.com/cucumber/godog"
)

// TestAcceptance runs all non-skipped play-ui acceptance scenarios.
// Walking skeleton scenario is always enabled; focused scenarios are added one at a time.
func TestAcceptance(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "play-ui",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../"},
			Tags:     "~@skip",
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("play-ui acceptance tests failed")
	}
}

// TestAcceptanceWalkingSkeleton runs only the walking skeleton scenario.
// Use this to verify the minimum observable user value end-to-end.
func TestAcceptanceWalkingSkeleton(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "play-ui-walking-skeleton",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../"},
			Tags:     "@walking_skeleton",
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("walking skeleton test failed")
	}
}

// TestAcceptanceAdapterIntegration runs real-io adapter integration scenarios.
// These exercise actual WebSocket connections and real protocol exchanges.
func TestAcceptanceAdapterIntegration(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "play-ui-adapter-integration",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../"},
			Tags:     "@adapter-integration",
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("adapter integration tests failed")
	}
}
