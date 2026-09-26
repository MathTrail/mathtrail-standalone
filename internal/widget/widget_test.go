package widget_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/widget"
)

// Whether or not a build has run, the package has a page to serve.
func TestPageIsAnHTMLDocument(t *testing.T) {
	t.Parallel()

	page := widget.Page()
	if !strings.HasPrefix(strings.ToLower(page), "<!doctype html>") {
		t.Errorf("Page() starts %q, want an HTML document", page[:min(len(page), 40)])
	}
}
