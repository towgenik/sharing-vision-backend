package post

import (
	"strings"
	"testing"
)

// validArticle returns an article that satisfies every spec rule (§4).
func validArticle() *Article {
	return &Article{
		Title:    strings.Repeat("A", 20),
		Content:  strings.Repeat("B", 200),
		Category: "Tech",
		Status:   StatusPublish,
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*Article)
		want string // expected error fragment, "" = valid
	}{
		{"valid publish", func(a *Article) {}, ""},
		{"missing title", func(a *Article) { a.Title = "  " }, "title is required"},
		{"title too short", func(a *Article) { a.Title = strings.Repeat("A", 19) }, "at least 20"},
		{"title too long", func(a *Article) { a.Title = strings.Repeat("A", 201) }, "exceeds 200"},
		{"missing content", func(a *Article) { a.Content = " " }, "content is required"},
		{"content too short", func(a *Article) { a.Content = strings.Repeat("B", 199) }, "at least 200"},
		{"missing category", func(a *Article) { a.Category = "" }, "category is required"},
		{"category too short", func(a *Article) { a.Category = "AB" }, "at least 3"},
		{"category too long", func(a *Article) { a.Category = strings.Repeat("C", 101) }, "exceeds 100"},
		{"missing status", func(a *Article) { a.Status = "" }, "status is required"},
		{"invalid status", func(a *Article) { a.Status = "deleted" }, "status is required"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := validArticle()
			tc.mut(a)
			ApplyDefaults(a)
			err := Validate(a)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("expected valid, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %q", tc.want, err.Error())
			}
		})
	}
}

func TestNormalizeStatus(t *testing.T) {
	cases := map[string]string{
		"publish": StatusPublish,
		"draft":   StatusDraft,
		"thrash":  StatusTrash,
		"":        "",
		"junk":    "",
	}
	for in, want := range cases {
		if got := NormalizeStatus(in); got != want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", in, got, want)
		}
	}
}
