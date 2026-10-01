package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/site/check"
)

const origin = "https://example.test"

func options() check.Options {
	return check.Options{BaseURL: origin, ReferenceLocale: "en", MaxPageBytes: 300 * 1024}
}

// page is a front page at address, in lang, that breaks none of the rules: it
// names its language and its direction, its title and its description, its
// own address, and every translation of a front page.
func page(address, lang string) string {
	return `<!DOCTYPE html><html lang="` + lang + `" dir="ltr"><head>` +
		`<title>Title</title><meta name="description" content="Description">` +
		`<link rel="canonical" href="` + origin + address + `">` +
		`<link rel="alternate" hreflang="en" href="` + origin + `/en/">` +
		`<link rel="alternate" hreflang="ru" href="` + origin + `/ru/">` +
		`<link rel="alternate" hreflang="x-default" href="` + origin + `/">` +
		`<link rel="stylesheet" href="/assets/style.css">` +
		`</head><body><a href="/ru/">Русский</a></body></html>`
}

// publishable writes a small site with nothing wrong in it — an apex, two
// locales and a stylesheet — into a directory, which is the only kind of input
// this command is ever pointed at.
func publishable(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	files := map[string]string{
		"index.html":       page("/", "en"),
		"en/index.html":    page("/en/", "en"),
		"ru/index.html":    page("/ru/", "ru"),
		"assets/style.css": "body{color:#000}",
	}
	for name, data := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create the directory of %s: %v", name, err)
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return dir
}

func TestRunPassesASiteWithNothingWrong(t *testing.T) {
	t.Parallel()

	var out strings.Builder
	if err := run(&out, publishable(t), options()); err != nil {
		t.Fatalf("run() error = %v, want no error", err)
	}
	if !strings.Contains(out.String(), "publishable") {
		t.Errorf("run() printed %q, want it to say the site is publishable", out.String())
	}
}

// A page that was hand-edited into the built directory is caught, and the
// failure names how many problems there are rather than only the first.
func TestRunFailsOnASiteThatShouldNotBePublished(t *testing.T) {
	t.Parallel()

	dir := publishable(t)
	if err := os.WriteFile(filepath.Join(dir, "en", "index.html"),
		[]byte(`<html><body><a href="/en/missing/">gone</a></body></html>`), 0o644); err != nil {
		t.Fatalf("overwrite a page: %v", err)
	}

	var out strings.Builder
	err := run(&out, dir, options())
	if err == nil {
		t.Fatalf("run() error = nil, want one counting the problems")
	}
	if !strings.Contains(err.Error(), "problems in") {
		t.Errorf("run() error = %q, want one counting the problems", err)
	}
	if !strings.Contains(out.String(), "which the site does not have") {
		t.Errorf("run() printed %q, want the dead link named", out.String())
	}
}

func TestRunRefusesADirectoryThatIsNotASite(t *testing.T) {
	t.Parallel()

	err := run(io.Discard, t.TempDir(), options())
	if err == nil {
		t.Fatalf("run() error = nil, want one about an empty site")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Errorf("run() error = %q, want one about an empty site", err)
	}
}
