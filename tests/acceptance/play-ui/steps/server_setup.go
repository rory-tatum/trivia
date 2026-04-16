// server_setup.go wires the in-process test server for play-ui acceptance tests.
//
// The server is identical in structure to the production server: same hub, same
// handlers, same auth guard. The play-ui tests drive primarily through the play
// room WebSocket (/ws?room=play), with the host room (/ws?token=HOST_TOKEN) used
// to drive game state transitions that create preconditions for play-room scenarios.
//
// NewPlayUITestServer is the Layer 3 server factory.
// It wires: hub → session → quizLoaderAdapter → hostHandler + playHandler + displayHandler
package steps

import (
	"net/http"
	"net/http/httptest"

	"trivia/internal/game"
	"trivia/internal/handler"
	"trivia/internal/hub"
	"trivia/internal/quiz"
	"trivia/internal/static"
)

// quizLoaderAdapter adapts quiz.Loader to the handler.QuizLoader port.
// This keeps QuizFull out of the handler package (architectural invariant).
type quizLoaderAdapter struct {
	loader *quiz.Loader
}

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
	return handler.QuizLoadedMeta{
		Title:         q.Title,
		RoundCount:    len(q.Rounds),
		QuestionCount: total,
	}, nil
}

// NewPlayUITestServer wires an in-process test server for play-ui acceptance tests.
// The server is production-equivalent: real hub, real session, real handlers.
func NewPlayUITestServer(hostToken string) *httptest.Server {
	h := hub.NewHub()
	session := game.NewGameSession()
	loader := &quizLoaderAdapter{loader: quiz.NewLoader()}
	hostHandler := handler.NewHostHandler(h, loader, "", session)
	playHandler := handler.NewPlayHandler(h, session, session)
	displayHandler := handler.NewDisplayHandler(h, session)
	authGuard := handler.NewAuthGuard(hostToken)

	mux := http.NewServeMux()
	mux.Handle("/", static.NewStaticHandler())
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		room := r.URL.Query().Get("room")
		switch {
		case token != "":
			authGuard(hostHandler).ServeHTTP(w, r)
		case room == "play":
			playHandler.ServeHTTP(w, r)
		case room == "display":
			displayHandler.ServeHTTP(w, r)
		default:
			http.Error(w, "invalid room parameter", http.StatusBadRequest)
		}
	})

	server := httptest.NewServer(mux)
	hostHandler.SetBaseURL(server.URL)
	return server
}
