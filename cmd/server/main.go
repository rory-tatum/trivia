// Package main is the server entry point.
// It reads configuration from the environment, wires dependencies,
// and starts the HTTP server.
package main

import (
	"log"
	"net/http"

	"trivia/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	mux := http.NewServeMux()
	_ = mux
	_ = cfg

	log.Println("starting trivia server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
