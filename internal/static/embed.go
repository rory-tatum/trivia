// Package static embeds the compiled frontend assets.
package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var Assets embed.FS

// NewStaticHandler returns an http.Handler that serves the embedded frontend
// assets. GET / serves dist/index.html; all other paths are served from dist/.
func NewStaticHandler() http.Handler {
	distFS, err := fs.Sub(Assets, "dist")
	if err != nil {
		panic("static: failed to create sub-filesystem: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(distFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			http.ServeFileFS(w, r, distFS, "index.html")
			return
		}
		// Try to serve the static asset; if it doesn't exist, fall back to
		// index.html so client-side routing (BrowserRouter) can take over.
		f, err := distFS.Open(r.URL.Path[1:])
		if err != nil {
			http.ServeFileFS(w, r, distFS, "index.html")
			return
		}
		f.Close()
		fileServer.ServeHTTP(w, r)
	})
}
