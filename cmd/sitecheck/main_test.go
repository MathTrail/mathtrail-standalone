package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/site/check"
	"github.com/MathTrail/mathtrail-standalone/internal/site/render"
)

const origin = "https://example.test"

func options() check.Options {
	return check.Options{BaseURL: origin, ReferenceLocale: "en", MaxPageBytes: 300 * 1024}
}

// built renders the real sources into a directory, which is the only kind of
// input this command is ever pointed at.
func built(t *testing.T) string {
	t.Helper()

	files, err := render.Build(os.DirFS("../../site"), render.Options{
		BaseURL:         origin,
		SiteName:        "MathTrail",
		ReferenceLocale: "en",
	})
	if err != nil {
		t.Fatalf("Build() error = %v, want no error", err)
	}

	dir := t.TempDir()
	if err := render.Write(dir, files); err != nil {
		t.Fatalf("Write() error = %v, want no error", err)
	}
	return dir
}

func TestRunPassesTheSiteThisRepositoryBuilds(t *testing.T) {
	t.Parallel()

	var out strings.Builder
	if err := run(&out, built(t), options()); err != nil {
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

	dir := built(t)
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
