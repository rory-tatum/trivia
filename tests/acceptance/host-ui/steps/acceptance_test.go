// acceptance_test.go is the test entry point for the host-ui acceptance suite.
//
// Run all non-skipped scenarios:
//
//	go test ./tests/acceptance/host-ui/steps/...
//
// Run only the walking skeleton:
//
//	go test ./tests/acceptance/host-ui/steps/... -run TestAcceptanceWalkingSkeleton
//
// Run real-io adapter integration tests:
//
//	go test ./tests/acceptance/host-ui/steps/... -run TestAcceptanceAdapterIntegration
package steps

import (
	"testing"

	"github.com/cucumber/godog"
)

// TestAcceptance runs all non-skipped host-ui acceptance scenarios.
// Walking skeleton scenario is always enabled; focused scenarios are added one at a time.
func TestAcceptance(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "host-ui",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../"},
			Tags:     "~@skip",
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("host-ui acceptance tests failed")
	}
}

// TestAcceptanceWalkingSkeleton runs only the walking skeleton scenario.
// Use this to verify the minimum observable user value end-to-end.
func TestAcceptanceWalkingSkeleton(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "host-ui-walking-skeleton",
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
// These exercise actual filesystem reads and real WebSocket dials.
func TestAcceptanceAdapterIntegration(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "host-ui-adapter-integration",
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
