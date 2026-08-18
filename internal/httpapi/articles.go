package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/towgenik/sharing-vision-backend/internal/post"
)

type articleAPI struct {
	repo post.Repository
}

// createArticle handles POST /api/article/.
func (h *articleAPI) createArticle(w http.ResponseWriter, r *http.Request) {
	var a post.Article
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	post.ApplyDefaults(&a)
	if err := post.Validate(&a); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.repo.Create(&a)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// listArticles handles GET /api/article/ (optional ?status=).
func (h *articleAPI) listArticles(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status != "" && !post.IsValidStatus(status) {
		writeErr(w, http.StatusBadRequest, "status must be one of publish, draft, thrash")
		return
	}
	list, err := h.repo.List(status)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// previewArticles handles GET /api/article/preview (published + pagination).
func (h *articleAPI) previewArticles(w http.ResponseWriter, r *http.Request) {
	page := atoiDefault(r.URL.Query().Get("page"), 1)
	perPage := atoiDefault(r.URL.Query().Get("per_page"), 10)
	list, total, err := h.repo.Preview(page, perPage)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"data":        list,
		"page":        page,
		"per_page":    perPage,
		"total":       total,
		"total_pages": post.TotalPages(total, perPage),
	})
}
