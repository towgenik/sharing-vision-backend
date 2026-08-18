package httpapi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/towgenik/sharing-vision-backend/internal/post"
)

// NewRouter builds the HTTP handler tree.
func NewRouter(repo post.Repository) http.Handler {
	h := &articleAPI{repo: repo}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(25 * time.Second))
	r.Use(corsMiddleware)

	r.Get("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Route("/api/article", func(r chi.Router) {
		r.Post("/", h.createArticle)       // create
		r.Get("/", h.listArticles)         // list (all / ?status=)
		r.Get("/preview", h.previewArticles) // published + pagination (BEFORE /{id})
		r.Get("/{id}", h.getArticle)       // read one
		r.Put("/{id}", h.updateArticle)    // full update incl. status
		r.Delete("/{id}", h.trashArticle)  // trash (soft delete)
	})

	return r
}

// corsMiddleware allows the public origin and local dev ports.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if isAllowedOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isAllowedOrigin(o string) bool {
	switch o {
	case "https://sv.farrel.moe",
		"http://localhost:5173",
		"http://localhost:4173":
		return true
	}
	return o == ""
}
