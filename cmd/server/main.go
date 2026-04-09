// Package main is the server entry point.
// It reads configuration from the environment, wires dependencies,
// and starts the HTTP server.
package main

import (
	"log"
	"net/http"
	"os"

	"trivia/config"
	"trivia/internal/game"
	"trivia/internal/handler"
	"trivia/internal/hub"
	"trivia/internal/quiz"
	"trivia/internal/static"
)

// quizLoaderAdapter adapts quiz.Loader to the handler.QuizLoader port,
// keeping QuizFull out of the handler package.
type quizLoaderAdapter struct{ loader *quiz.Loader }

func (a *quizLoaderAdapter) LoadIntoSession(path string, session game.GamePort) (handler.QuizLoadedMeta, error) {
	q, err := a.loader.LoadFromPath(path)
	if err != nil {
		return handler.QuizLoadedMeta{}, err
	}
	if err := session.Load(q); err != nil {
		return handler.QuizLoadedMeta{}, err
	}
	total := 0
	for _, r := range q.Rounds {
		total += len(r.Questions)
	}
	return handler.QuizLoadedMeta{Title: q.Title, RoundCount: len(q.Rounds), QuestionCount: total}, nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	h := hub.NewHub()
	session := game.NewGameSession()
	loader := &quizLoaderAdapter{loader: quiz.NewLoader()}
	hostHandler := handler.NewHostHandler(h, loader, baseURL, session)
	playHandler := handler.NewPlayHandler(h, session, session)
	displayHandler := handler.NewDisplayHandler(h, session)
	authGuard := handler.NewAuthGuard(cfg.HostToken)

	mux := http.NewServeMux()
	mux.Handle("/", static.NewStaticHandler())
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		room := r.URL.Query().Get("room")
		if token != "" {
			authGuard(hostHandler).ServeHTTP(w, r)
		} else if room == "play" {
			playHandler.ServeHTTP(w, r)
		} else if room == "display" {
			displayHandler.ServeHTTP(w, r)
		} else {
			http.Error(w, "invalid room", http.StatusBadRequest)
		}
	})

	log.Println("starting trivia server on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
