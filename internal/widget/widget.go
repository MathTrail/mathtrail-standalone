// Package widget holds the page every card in a chat is drawn by: the widget
// as the web build left it. Where no build has run — a fresh checkout, the
// tests — a placeholder page stands in, and says so to whoever opens it.
//
// It holds the design tokens too, the one stylesheet of colours, spacing and
// type everything the service shows is drawn with.
package widget

import (
	"embed"
	"io/fs"
)

// The web build writes the widget beside this file. The placeholder is
// committed, so the package compiles and its tests run whether or not a build
// has run.
const (
	built       = "widget.html"
	placeholder = "placeholder.html"
)

//go:embed *.html
var pages embed.FS

//go:embed tokens.css
var tokens string

// Page is the widget: the page the build left when there is one, and the
// placeholder when there is not.
func Page() string {
	return page(pages)
}

// Tokens is the stylesheet of the design tokens: custom properties alone,
// for the light and the dark theme, which a page's own styles read.
func Tokens() string {
	return tokens
}

// page reads the built page from files when it is there, and the placeholder
// when it is not. It is empty only when neither is, which a build of this
// package cannot produce.
func page(files fs.FS) string {
	if html, err := fs.ReadFile(files, built); err == nil {
		return string(html)
	}
	html, err := fs.ReadFile(files, placeholder)
	if err != nil {
		return ""
	}
	return string(html)
}
