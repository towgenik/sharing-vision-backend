package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/towgenik/sharing-vision-backend/internal/post"
)

// Fixtures satisfying the spec validation rules (§4).
const (
	validTitle = "A valid article title that is long enough"
	validCat   = "Tech"
)

var validContent = strings.Repeat("Content ", 40) // 320 chars, > 200 required

func jsonBody(title, content, category, status string) string {
	b, _ := json.Marshal(map[string]string{
		"Title":    title,
		"Content":  content,
		"Category": category,
		"Status":   status,
	})
	return string(b)
}

func doReq(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, path, nil)
	} else {
		r = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	}
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func decode(t *testing.T, payload []byte, into any) {
	t.Helper()
	if err := json.Unmarshal(payload, into); err != nil {
		t.Fatalf("decode %s: %v", payload, err)
	}
}

func TestCreateArticle(t *testing.T) {
	repo := newStub()
	h := NewRouter(repo)

	w := doReq(t, h, http.MethodPost, "/article/",
		jsonBody(validTitle, validContent, validCat, "publish"))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (%s)", w.Code, w.Body.String())
	}
	var a post.Article
	decode(t, w.Body.Bytes(), &a)
	if a.Id != 1 || a.Status != "publish" {
		t.Fatalf("unexpected article: %+v", a)
	}

	// Too-short title → 422 (spec §4: min 20 chars).
	w = doReq(t, h, http.MethodPost, "/article/",
		jsonBody("too short", validContent, validCat, "publish"))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("short title: got %d, want 422", w.Code)
	}

	// Missing/invalid status → 422 (no silent default anymore).
	w = doReq(t, h, http.MethodPost, "/article/",
		jsonBody(validTitle, validContent, validCat, "deleted"))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid status: got %d, want 422", w.Code)
	}
	w = doReq(t, h, http.MethodPost, "/article/",
		jsonBody(validTitle, validContent, validCat, ""))
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("missing status: got %d, want 422", w.Code)
	}
}

func TestListArticlesPagination(t *testing.T) {
	repo := newStub()
	repo.add(
		post.Article{Id: 1, Title: "one", Status: "publish"},
		post.Article{Id: 2, Title: "two", Status: "draft"},
		post.Article{Id: 3, Title: "three", Status: "publish"},
		post.Article{Id: 4, Title: "four", Status: "publish"},
	)
	h := NewRouter(repo)

	// spec: /article/{limit}/{offset} — page 2 of published (size 2).
	w := doReq(t, h, http.MethodGet, "/article/2/2?status=publish", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d (%s)", w.Code, w.Body.String())
	}
	var list []post.Article
	decode(t, w.Body.Bytes(), &list)
	if len(list) != 1 {
		t.Fatalf("page 2: expected 1 row, got %d", len(list))
	}
	if totalHdr := w.Header().Get("X-Total-Count"); totalHdr != "3" {
		t.Fatalf("X-Total-Count = %q, want 3", totalHdr)
	}

	// limit=0 must be rejected.
	if w := doReq(t, h, http.MethodGet, "/article/0/0", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("limit=0 should be rejected, got %d", w.Code)
	}

	// Bad status filter.
	if w := doReq(t, h, http.MethodGet, "/article/10/0?status=bogus", ""); w.Code != http.StatusBadRequest {
		t.Fatalf("bad status filter: got %d, want 400", w.Code)
	}

	// Alias mount /api/article serves the same surface.
	if w := doReq(t, h, http.MethodGet, "/api/article/10/0", ""); w.Code != http.StatusOK {
		t.Fatalf("alias mount: got %d", w.Code)
	}
}

func TestGetUpdateTrash(t *testing.T) {
	repo := newStub()
	repo.add(post.Article{Id: 1, Title: validTitle, Content: validContent, Category: validCat, Status: "draft"})
	h := NewRouter(repo)

	// get → 200; missing → 404
	if w := doReq(t, h, http.MethodGet, "/article/1", ""); w.Code != http.StatusOK {
		t.Fatalf("get: %d", w.Code)
	}
	if w := doReq(t, h, http.MethodGet, "/article/99", ""); w.Code != http.StatusNotFound {
		t.Fatalf("get missing: %d, want 404", w.Code)
	}

	// update (valid body)
	w := doReq(t, h, http.MethodPut, "/article/1",
		jsonBody(validTitle, validContent, validCat, "publish"))
	if w.Code != http.StatusOK {
		t.Fatalf("update: %d (%s)", w.Code, w.Body.String())
	}
	var a post.Article
	decode(t, w.Body.Bytes(), &a)
	if a.Title != validTitle || a.Status != "publish" {
		t.Fatalf("after update: %+v", a)
	}

	// update with too-short content → 422 (spec §4 min 200).
	if w := doReq(t, h, http.MethodPut, "/article/1",
		jsonBody(validTitle, "tiny", validCat, "publish")); w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("short content update: got %d, want 422", w.Code)
	}

	// PATCH accepted on the alias mount too.
	w = doReq(t, h, http.MethodPatch, "/api/article/1",
		jsonBody(validTitle, validContent, validCat, "draft"))
	if w.Code != http.StatusOK {
		t.Fatalf("patch: %d (%s)", w.Code, w.Body.String())
	}

	// trash (soft delete) → still fetchable, status flips to thrash.
	if w := doReq(t, h, http.MethodDelete, "/api/article/1", ""); w.Code != http.StatusNoContent {
		t.Fatalf("trash: %d", w.Code)
	}
	if w := doReq(t, h, http.MethodGet, "/api/article/1", ""); w.Code != http.StatusOK {
		t.Fatalf("after trash should remain fetchable: %d", w.Code)
	}
	got, _ := repo.Get(1)
	if got.Status != post.StatusTrash {
		t.Fatalf("status = %q, want thrash", got.Status)
	}
}
