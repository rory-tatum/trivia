//go:build !mutation

package steps

import (
	"os"
	"strings"
	"testing"
)

// Test Budget: 1 behavior x 2 = 2 max unit tests. Using 1.
// Behavior: .go-arch-lint.yml exists in project root and declares architecture boundary rules.

func TestGoArchLintConfigExists(t *testing.T) {
	// The project root is 4 levels up from this package directory.
	configPath := "../../../../.go-arch-lint.yml"

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf(".go-arch-lint.yml does not exist in project root: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "version") {
		t.Error(".go-arch-lint.yaml must declare a 'version' field")
	}
	if !strings.Contains(content, "components") {
		t.Error(".go-arch-lint.yaml must declare 'components' with package rules")
	}
	if !strings.Contains(content, "handler") {
		t.Error(".go-arch-lint.yaml must include the 'handler' component")
	}
	if !strings.Contains(content, "hub") {
		t.Error(".go-arch-lint.yaml must include the 'hub' component")
	}
}
