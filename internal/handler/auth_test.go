package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"trivia/internal/handler"
)

// Test budget: 3 behaviors x 2 = 6 max unit tests. Using 3.

func TestAuthGuard_NoToken_Returns403(t *testing.T) {
	guard := handler.NewAuthGuard("secret")
	h := guard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called without token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/host", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestAuthGuard_WrongToken_Returns403(t *testing.T) {
	guard := handler.NewAuthGuard("secret")
	h := guard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("next handler should not be called with wrong token")
	}))

	req := httptest.NewRequest(http.MethodGet, "/host?token=wrong", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rr.Code)
	}
}

func TestAuthGuard_CorrectToken_PassesThrough(t *testing.T) {
	const token = "secret"
	guard := handler.NewAuthGuard(token)

	nextCalled := false
	h := guard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/host?token="+token, nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if !nextCalled {
		t.Error("expected next handler to be called with correct token")
	}
}
