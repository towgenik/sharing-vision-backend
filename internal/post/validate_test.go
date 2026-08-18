package post

import "testing"

func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		art  Article
		want string // expected error fragment, "" = valid
	}{
		{"valid publish", Article{Title: "Hi", Content: "c", Category: "News", Status: "publish"}, ""},
		{"missing title", Article{Title: "", Status: "draft"}, "title is required"},
		{"title too long", Article{Title: string(make([]rune, 201)), Status: "draft"}, "exceeds 200"},
		{"category too long", Article{Title: "ok", Category: string(make([]rune, 101)), Status: "draft"}, "exceeds 100"},
		{"invalid status", Article{Title: "ok", Status: "deleted"}, "status must be one of"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(&tc.art)
			if tc.want == "" {
				if err != nil {
					t.Fatalf("expected valid, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.want)
			}
			if !contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %q", tc.want, err.Error())
			}
		})
	}
}

func TestNormalizeStatus(t *testing.T) {
	cases := map[string]string{
		"":        StatusDraft,
		"publish": StatusPublish,
		"draft":   StatusDraft,
		"thrash":  StatusTrash,
		"junk":    "",
	}
	for in, want := range cases {
		if got := NormalizeStatus(in); got != want {
			t.Errorf("NormalizeStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
