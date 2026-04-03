package steps

import (
	"net/http"
	"net/http/httptest"

	"trivia/internal/game"
	"trivia/internal/handler"
	"trivia/internal/hub"
	"trivia/internal/quiz"
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

// NewTestServer wires an in-process test server for acceptance tests.
// It creates a shared game session and registers host, play, and display handlers.
func NewTestServer(hostToken string) *httptest.Server {
	h := hub.NewHub()
	session := game.NewGameSession()
	loader := &quizLoaderAdapter{loader: quiz.NewLoader()}
	hostHandler := handler.NewHostHandler(h, loader, "", session)
	playHandler := handler.NewPlayHandler(h, session, session)
	displayHandler := handler.NewDisplayHandler(h, session)
	authGuard := handler.NewAuthGuard(hostToken)

	mux := http.NewServeMux()
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

	server := httptest.NewServer(mux)
	hostHandler.SetBaseURL(server.URL)
	return server
}
