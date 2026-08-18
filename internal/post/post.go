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

// Spec validation bounds (Backend_Test-Sharing_Vision.md §4).
const (
	MinTitleLen    = 20
	MaxTitleLen    = 200
	MinContentLen  = 200
	MinCategoryLen = 3
	MaxCategoryLen = 100
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
	// List returns one page (limit/offset) ordered newest first, optionally
	// filtered by status, plus the total row count for that filter.
	List(limit, offset int, status string) ([]Article, int64, error)
	Get(id int64) (*Article, error)
	Update(a *Article) (*Article, error)
	Trash(id int64) error
}

// IsValidStatus reports whether s is one of publish|draft|thrash.
func IsValidStatus(s string) bool { return validStatuses[s] }

// NormalizeStatus returns s when it is valid, otherwise "".
func NormalizeStatus(s string) string {
	if IsValidStatus(s) {
		return s
	}
	return ""
}

// ApplyDefaults trims text fields. Status is intentionally NOT defaulted:
// the spec marks it required on both create and update.
func ApplyDefaults(a *Article) {
	a.Title = strings.TrimSpace(a.Title)
	a.Content = strings.TrimSpace(a.Content)
	a.Category = strings.TrimSpace(a.Category)
}

// Validate enforces the spec constraints (run before create and update).
func Validate(a *Article) error {
	if a.Title == "" {
		return errors.New("title is required")
	}
	if n := len([]rune(a.Title)); n < MinTitleLen {
		return errors.New("title must be at least 20 characters")
	} else if n > MaxTitleLen {
		return errors.New("title exceeds 200 characters")
	}
	if a.Content == "" {
		return errors.New("content is required")
	}
	if len([]rune(a.Content)) < MinContentLen {
		return errors.New("content must be at least 200 characters")
	}
	if a.Category == "" {
		return errors.New("category is required")
	}
	if n := len([]rune(a.Category)); n < MinCategoryLen {
		return errors.New("category must be at least 3 characters")
	} else if n > MaxCategoryLen {
		return errors.New("category exceeds 100 characters")
	}
	if !IsValidStatus(a.Status) {
		return errors.New("status is required and must be one of publish, draft, thrash")
	}
	return nil
}
