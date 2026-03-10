// internal/ui/ui.go
package ui

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var dist embed.FS

// Handler returns an http.Handler that serves the embedded React app.
// Unknown paths fall back to index.html for client-side routing.
func Handler() http.Handler {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return &spaHandler{fs: http.FS(sub)}
}

type spaHandler struct {
	fs http.FileSystem
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, err := h.fs.Open(r.URL.Path)
	if err != nil {
		// Fall back to index.html for SPA routing
		r2 := *r
		r2.URL.Path = "/"
		http.FileServer(h.fs).ServeHTTP(w, &r2)
		return
	}
	f.Close()
	http.FileServer(h.fs).ServeHTTP(w, r)
}
