// Package post holds the article model, validation, and repository contract.
package post

import (
	"errors"
	"strings"
	"time"
)

const (
	StatusPublish = "publish"
	StatusDraft   = "draft"
	StatusTrash   = "thrash"
)

var validStatuses = map[string]bool{
	StatusPublish: true,
	StatusDraft:   true,
	StatusTrash:   true,
}

// Article mirrors the `posts` table (exact spec column names).
type Article struct {
	Id          int64     `json:"Id"`
	Title       string    `json:"Title"`
	Content     string    `json:"Content"`
	Category    string    `json:"Category"`
	CreatedDate time.Time `json:"Created_date"`
	UpdatedDate time.Time `json:"Updated_date"`
	Status      string    `json:"Status"`
}

// Repository is the storage contract the HTTP layer depends on.
type Repository interface {
	Create(a *Article) (*Article, error)
	List(status string) ([]Article, error)
	Get(id int64) (*Article, error)
	Update(a *Article) (*Article, error)
	Trash(id int64) error
	Preview(page, perPage int) ([]Article, int64, error)
}

// IsValidStatus reports whether s is one of publish|draft|thrash.
func IsValidStatus(s string) bool { return validStatuses[s] }

// NormalizeStatus returns a valid status string, the draft default, or "" if invalid.
func NormalizeStatus(s string) string {
	if IsValidStatus(s) {
		return s
	}
	if s == "" {
		return StatusDraft
	}
	return ""
}

// ApplyDefaults trims fields and fills the status default for new articles.
func ApplyDefaults(a *Article) {
	a.Title = strings.TrimSpace(a.Title)
	a.Category = strings.TrimSpace(a.Category)
	if a.Status == "" {
		a.Status = StatusDraft
	}
}

// Validate enforces the spec constraints.
func Validate(a *Article) error {
	if strings.TrimSpace(a.Title) == "" {
		return errors.New("title is required")
	}
	if len([]rune(a.Title)) > 200 {
		return errors.New("title exceeds 200 characters")
	}
	if len([]rune(a.Category)) > 100 {
		return errors.New("category exceeds 100 characters")
	}
	if NormalizeStatus(a.Status) == "" {
		return errors.New("status must be one of publish, draft, thrash")
	}
	return nil
}
