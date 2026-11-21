package public

import "testing"

func TestInlineImageType(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  string
	}{
		{"image/png", "image/png"},
		{"image/jpeg; charset=binary", "image/jpeg"},
		{"image/gif", "image/gif"},
		{"image/webp", "image/webp"},
		{"image/svg+xml", ""},
		{"text/html", ""},
		{"not a media type", ""},
	} {
		if got := inlineImageType(tc.input); got != tc.want {
			t.Errorf("inlineImageType(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
