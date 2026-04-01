package config_test

import (
	"testing"

	"trivia/config"
)

func TestLoad_MissingHostToken_ReturnsError(t *testing.T) {
	t.Setenv("HOST_TOKEN", "")
	t.Setenv("QUIZ_DIR", "/some/dir")

	_, err := config.Load()

	if err == nil {
		t.Fatal("expected error when HOST_TOKEN is missing, got nil")
	}
}

func TestLoad_MissingQuizDir_ReturnsError(t *testing.T) {
	t.Setenv("HOST_TOKEN", "secret-token")
	t.Setenv("QUIZ_DIR", "")

	_, err := config.Load()

	if err == nil {
		t.Fatal("expected error when QUIZ_DIR is missing, got nil")
	}
}

func TestLoad_BothPresent_ReturnsValidConfig(t *testing.T) {
	t.Setenv("HOST_TOKEN", "secret-token")
	t.Setenv("QUIZ_DIR", "/quizzes")

	cfg, err := config.Load()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.HostToken != "secret-token" {
		t.Errorf("expected HostToken %q, got %q", "secret-token", cfg.HostToken)
	}
	if cfg.QuizDir != "/quizzes" {
		t.Errorf("expected QuizDir %q, got %q", "/quizzes", cfg.QuizDir)
	}
}
