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

// The tokens are custom properties for both themes: the light ones at the
// root, and the dark ones for a viewer who prefers them.
func TestTokensDefineBothThemes(t *testing.T) {
	t.Parallel()

	tokens := widget.Tokens()
	for _, want := range []string{":root", "--surface:", "--text:", "--accent:", "--font-sans:", "prefers-color-scheme: dark", `[data-theme="dark"]`} {
		if !strings.Contains(tokens, want) {
			t.Errorf("Tokens() has no %q, want the design tokens of both themes", want)
		}
	}
}
