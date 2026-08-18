package httpapi

import (
	"sync"
	"time"

	"github.com/towgenik/sharing-vision-backend/internal/post"
)

// stubRepo is an in-memory Repository for handler tests.
type stubRepo struct {
	mu     sync.Mutex
	nextID int64
	rows   []post.Article
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
	s.rows = append(s.rows, *a)
	return a, nil
}

func (s *stubRepo) List(limit, offset int, status string) ([]post.Article, int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	filtered := []post.Article{}
	for _, a := range s.rows {
		if status == "" || a.Status == status {
			filtered = append(filtered, a)
		}
	}
	total := int64(len(filtered))
	if limit < 1 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	if offset > len(filtered) {
		offset = len(filtered)
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], total, nil
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
