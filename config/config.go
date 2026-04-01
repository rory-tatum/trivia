// Package config reads server configuration from environment variables.
package config

import (
	"fmt"
	"os"
)

// Config holds the server's runtime configuration.
type Config struct {
	HostToken string
	QuizDir   string
}

// Load reads HOST_TOKEN and QUIZ_DIR from the environment.
// Returns an error if either variable is missing or empty.
func Load() (Config, error) {
	hostToken := os.Getenv("HOST_TOKEN")
	if hostToken == "" {
		return Config{}, fmt.Errorf("HOST_TOKEN environment variable is required")
	}
	quizDir := os.Getenv("QUIZ_DIR")
	if quizDir == "" {
		return Config{}, fmt.Errorf("QUIZ_DIR environment variable is required")
	}
	return Config{
		HostToken: hostToken,
		QuizDir:   quizDir,
	}, nil
}
