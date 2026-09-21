package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/site/render"
)

const sourceDir = "../../site"

func options() render.Options {
	return render.Options{
		BaseURL:         "https://example.test",
		SiteName:        "MathTrail",
		ReferenceLocale: "en",
	}
}

// A rendered site replaces whatever was in the output directory: a page that
// has been deleted from the sources must not survive into the next deployment.
func TestRunLeavesNothingOfThePreviousBuild(t *testing.T) {
	out := t.TempDir()
	stale := filepath.Join(out, "en", "gone", "index.html")
	if err := os.MkdirAll(filepath.Dir(stale), 0o755); err != nil {
		t.Fatalf("prepare a stale page: %v", err)
	}
	if err := os.WriteFile(stale, []byte("<html></html>"), 0o644); err != nil {
		t.Fatalf("prepare a stale page: %v", err)
	}

	if err := run(sourceDir, out, "", options()); err != nil {
		t.Fatalf("run() error = %v, want no error", err)
	}

	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("the stale page is still there, Stat() error = %v, want a not-exist error", err)
	}
	if _, err := os.Stat(filepath.Join(out, "en", "index.html")); err != nil {
		t.Errorf("the site was not written, Stat() error = %v, want no error", err)
	}
}

// A base URL the site cannot be built from stops the run before anything is
// written, so a half-correct site is never published.
func TestRunRefusesABaseURLItCannotBuildFrom(t *testing.T) {
	out := t.TempDir()
	opt := options()
	opt.BaseURL = "not a url at all"

	err := run(sourceDir, out, "", opt)
	if err == nil {
		t.Fatalf("run() error = nil, want one about the base URL")
	}
	if !strings.Contains(err.Error(), "base URL") {
		t.Errorf("run() error = %q, want one about the base URL", err)
	}

	entries, err := os.ReadDir(out)
	if err != nil {
		t.Fatalf("ReadDir() error = %v, want no error", err)
	}
	if len(entries) != 0 {
		t.Errorf("the output directory holds %d entries, want it untouched", len(entries))
	}
}
