package render

import (
	"strings"
	"testing"
)

func TestParseDocumentReadsTheHeaderAndKeepsTheBody(t *testing.T) {
	t.Parallel()

	doc, err := parseDocument([]byte("---\ntitle: Home\ndescription: What this is\n---\n\n# Home\n"))
	if err != nil {
		t.Fatalf("parseDocument() error = %v, want no error", err)
	}
	if doc.title != "Home" {
		t.Errorf("title = %q, want %q", doc.title, "Home")
	}
	if doc.description != "What this is" {
		t.Errorf("description = %q, want %q", doc.description, "What this is")
	}
	if strings.TrimSpace(string(doc.body)) != "# Home" {
		t.Errorf("body = %q, want %q", doc.body, "# Home")
	}
}

func TestParseDocumentRefusesAHeaderItCannotTrust(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  string
		wantErr string
	}{
		{
			name:    "no front matter at all",
			source:  "# Home\n",
			wantErr: "must open",
		},
		{
			name:    "front matter that is never closed",
			source:  "---\ntitle: Home\n",
			wantErr: "never closed",
		},
		{
			name:    "a line that is not a pair",
			source:  "---\ntitle Home\n---\n",
			wantErr: "not key: value",
		},
		{
			name:    "a key nobody renders",
			source:  "---\ntitle: Home\ndescription: D\nauthor: Someone\n---\n",
			wantErr: `key "author"`,
		},
		{
			name:    "no title",
			source:  "---\ndescription: D\n---\n",
			wantErr: "no title",
		},
		{
			name:    "no description",
			source:  "---\ntitle: Home\n---\n",
			wantErr: "no description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := parseDocument([]byte(tt.source))
			if err == nil {
				t.Fatalf("parseDocument() error = nil, want one containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("parseDocument() error = %q, want one containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestDirectionFollowsTheScriptNotTheRegion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		locale string
		want   string
	}{
		{locale: "en", want: "ltr"},
		{locale: "ru", want: "ltr"},
		{locale: "ar", want: "rtl"},
		{locale: "fa", want: "rtl"},
		{locale: "ur", want: "rtl"},
		{locale: "ar-EG", want: "rtl"},
		{locale: "pt-BR", want: "ltr"},
	}

	for _, tt := range tests {
		t.Run(tt.locale, func(t *testing.T) {
			t.Parallel()

			if got := direction(tt.locale); got != tt.want {
				t.Errorf("direction(%q) = %q, want %q", tt.locale, got, tt.want)
			}
		})
	}
}

func FuzzParseDocument(f *testing.F) {
	f.Add("---\ntitle: Home\ndescription: D\n---\n\n# Home\n")
	f.Add("---\n---\n")
	f.Add("")
	f.Add("---\ntitle: :::\n")

	f.Fuzz(func(t *testing.T, source string) {
		doc, err := parseDocument([]byte(source))
		if err != nil {
			return
		}
		if doc.title == "" || doc.description == "" {
			t.Errorf("parseDocument(%q) accepted a document with no title or no description", source)
		}
	})
}
