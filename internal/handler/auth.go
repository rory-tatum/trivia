package handler

import (
	"net/http"
)

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// NewAuthGuard returns a Middleware that validates the ?token query parameter
// against the provided hostToken. Requests with a missing or mismatched token
// receive HTTP 403 Forbidden and are not forwarded to the next handler.
func NewAuthGuard(hostToken string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.URL.Query().Get("token")
			if token != hostToken {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
