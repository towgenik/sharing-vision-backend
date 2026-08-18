package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/towgenik/sharing-vision-backend/internal/post"
)

// getArticle handles GET /api/article/{id}.
func (h *articleAPI) getArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	a, err := h.repo.Get(id)
	if errors.Is(err, post.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "article not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, a)
}

// updateArticle handles PUT /api/article/{id} (full update incl. status).
func (h *articleAPI) updateArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
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
	a.Id = id
	updated, err := h.repo.Update(&a)
	if errors.Is(err, post.ErrNotFound) {
		writeErr(w, http.StatusNotFound, "article not found")
		return
	}
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// trashArticle handles DELETE /api/article/{id} (soft delete -> thrash).
func (h *articleAPI) trashArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := idParam(r)
	if !ok {
		writeErr(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.Trash(id); err != nil {
		if errors.Is(err, post.ErrNotFound) {
			writeErr(w, http.StatusNotFound, "article not found")
			return
		}
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// idParam parses the {id} route parameter.
func idParam(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id < 1 {
		return 0, false
	}
	return id, true
}
