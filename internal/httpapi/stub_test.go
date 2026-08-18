package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/towgenik/sharing-vision-backend/internal/post"
)

// stubRepo is an in-memory Repository for handler tests.
type stubRepo struct {
	mu      sync.Mutex
	nextID  int64
	rows    []post.Article
	preview func(page, per int) ([]post.Article, int64, error)
}

func newStub() *stubRepo {
	return &stubRepo{nextID: 1, rows: []post.Article{}}
}

func (s *stubRepo) add(rows ...post.Article) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows = append(s.rows, rows...)
}

func (s *stubRepo) Create(a *post.Article) (*post.Article, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a.Id = s.nextID
	a.CreatedDate = time.Now()
	a.UpdatedDate = time.Now()
	s.nextID++
	s.rows = append([]post.Article{*a}, s.rows...)
	return a, nil
}

func (s *stubRepo) List(status string) ([]post.Article, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := []post.Article{}
	for _, a := range s.rows {
		if status == "" || a.Status == status {
			out = append(out, a)
		}
	}
	return out, nil
}

func (s *stubRepo) Get(id int64) (*post.Article, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.rows {
		if s.rows[i].Id == id {
			return &s.rows[i], nil
		}
	}
	return nil, post.ErrNotFound
}

func (s *stubRepo) Update(a *post.Article) (*post.Article, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.rows {
		if s.rows[i].Id == a.Id {
			s.rows[i] = *a
			s.rows[i].UpdatedDate = time.Now()
			return &s.rows[i], nil
		}
	}
	return nil, post.ErrNotFound
}

func (s *stubRepo) Trash(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.rows {
		if s.rows[i].Id == id {
			s.rows[i].Status = post.StatusTrash
			return nil
		}
	}
	return post.ErrNotFound
}

func (s *stubRepo) Preview(page, per int) ([]post.Article, int64, error) {
	if s.preview != nil {
		return s.preview(page, per)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	published := []post.Article{}
	for _, a := range s.rows {
		if a.Status == post.StatusPublish {
			published = append(published, a)
		}
	}
	return published, int64(len(published)), nil
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

// tM builds a time.Time offset by n days for ordering assertions.
func tM(n int64) time.Time {
	return time.Unix(1700000000+n*86400, 0).UTC()
}
