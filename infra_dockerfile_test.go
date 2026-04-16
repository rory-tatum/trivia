//go:build !mutation

// Package trivia tests infrastructure: verifies the Dockerfile exists and
// follows the required multi-stage build structure.
package main

import (
	"os"
	"strings"
	"testing"
)

func TestDockerfileExistsAndIsMultiStage(t *testing.T) {
	content, err := os.ReadFile("Dockerfile")
	if err != nil {
		t.Fatalf("Dockerfile not found at project root: %v", err)
	}

	fromCount := 0
	for _, line := range strings.Split(string(content), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToUpper(trimmed), "FROM ") {
			fromCount++
		}
	}

	if fromCount < 3 {
		t.Errorf("Dockerfile must have at least 3 FROM statements (multi-stage build), found %d", fromCount)
	}
}
