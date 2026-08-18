package httpapi

import (
	"net/http"
	"strings"
	"testing"

	"github.com/towgenik/sharing-vision-backend/internal/post"
)

func TestCreateArticle(t *testing.T) {
	repo := newStub()
	h := NewRouter(repo)

	w := doReq(t, h, http.MethodPost, "/api/article/", `{"Title":"Hello","Content":"body","Category":"News","Status":"publish"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (%s)", w.Code, w.Body.String())
	}
	var a post.Article
	decode(t, w.Body.Bytes(), &a)
	if a.Id != 1 || a.Status != "publish" {
		t.Fatalf("unexpected article: %+v", a)
	}

	// invalid status
	w = doReq(t, h, http.MethodPost, "/api/article/", `{"Title":"x","Status":"deleted"}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid status: got %d, want 400", w.Code)
	}

	// missing title
	w = doReq(t, h, http.MethodPost, "/api/article/", `{"Title":"  "}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing title: got %d, want 400", w.Code)
	}
}

func TestListArticlesFilter(t *testing.T) {
	repo := newStub()
	repo.add(post.Article{Id: 1, Title: "a", Status: "publish"},
		post.Article{Id: 2, Title: "b", Status: "draft"})
	h := NewRouter(repo)

	w := doReq(t, h, http.MethodGet, "/api/article/?status=draft", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var list []post.Article
	decode(t, w.Body.Bytes(), &list)
	if len(list) != 1 || list[0].Title != "b" {
		t.Fatalf("unexpected list: %+v", list)
	}

	w = doReq(t, h, http.MethodGet, "/api/article/?status=bogus", "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad status filter: got %d, want 400", w.Code)
	}
}

func TestGetUpdateTrash(t *testing.T) {
	repo := newStub()
	repo.add(post.Article{Id: 1, Title: "a", Status: "draft"})
	h := NewRouter(repo)

	// get
	w := doReq(t, h, http.MethodGet, "/api/article/1", "")
	if w.Code != http.StatusOK {
		t.Fatalf("get: %d", w.Code)
	}
	// get missing
	if w := doReq(t, h, http.MethodGet, "/api/article/99", ""); w.Code != http.StatusNotFound {
		t.Fatalf("get missing: %d, want 404", w.Code)
	}

	// update
	w = doReq(t, h, http.MethodPut, "/api/article/1", `{"Title":"updated","Content":"","Category":"","Status":"publish"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("update: %d (%s)", w.Code, w.Body.String())
	}
	var a post.Article
	decode(t, w.Body.Bytes(), &a)
	if a.Title != "updated" {
		t.Fatalf("title = %q", a.Title)
	}

	// trash
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

func TestPreviewOrderAndRoute(t *testing.T) {
	repo := newStub()
	a := post.Article{Id: 1, Title: "a", Status: "publish", CreatedDate: tM(1)}
	b := post.Article{Id: 2, Title: "b", Status: "publish", CreatedDate: tM(2)}
	repo.add(a, b)
	h := NewRouter(repo)

	// /preview must NOT be swallowed by /{id}
	w := doReq(t, h, http.MethodGet, "/api/article/preview?page=1&per_page=10", "")
	if w.Code != http.StatusOK {
		t.Fatalf("preview: %d (%s)", w.Code, w.Body.String())
	}
	var pr struct {
		Data       []post.Article `json:"data"`
		Total      int64          `json:"total"`
		TotalPages int64          `json:"total_pages"`
	}
	decode(t, w.Body.Bytes(), &pr)
	if pr.Total != 2 || len(pr.Data) != 2 {
		t.Fatalf("preview wrong: %+v", pr)
	}
	if strings.Contains(w.Body.String(), `"error"`) {
		t.Fatal("preview returned an error payload")
	}
}
