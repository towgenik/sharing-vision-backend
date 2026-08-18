package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/towgenik/sharing-vision-backend/internal/post"
)

type articleAPI struct {
	repo post.Repository
}

// createArticle handles POST /article/ and POST /api/article/ (spec §3 row 1).
func (h *articleAPI) createArticle(w http.ResponseWriter, r *http.Request) {
	var a post.Article
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	post.ApplyDefaults(&a)
	if err := post.Validate(&a); err != nil {
		writeErr(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	created, err := h.repo.Create(&a)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// listArticles handles GET /article/{limit}/{offset} (spec §3 row 2).
// Pagination uses limit + offset path parameters; response body is a JSON
// array, with the total row count carried in the X-Total-Count header.
func (h *articleAPI) listArticles(w http.ResponseWriter, r *http.Request) {
	limit, ok1 := posParam(r, "limit")
	offset, ok2 := nonnegParam(r, "offset")
	if !ok1 || !ok2 {
		writeErr(w, http.StatusBadRequest, "limit must be a positive integer and offset a non-negative integer")
		return
	}
	status := r.URL.Query().Get("status")
	if status != "" && !post.IsValidStatus(status) {
		writeErr(w, http.StatusBadRequest, "status must be one of publish, draft, thrash")
		return
	}
	list, total, err := h.repo.List(limit, offset, status)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	hdr := w.Header()
	hdr.Set("X-Total-Count", itoa(total))
	hdr.Set("Access-Control-Expose-Headers", "X-Total-Count")
	writeJSON(w, http.StatusOK, list)
}
