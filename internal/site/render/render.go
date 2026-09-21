// Package render turns the site's sources — Markdown texts, HTML templates and
// static assets — into the exact set of files a static host serves.
//
// Nothing here touches the filesystem: the sources arrive as an [fs.FS] and the
// result comes back as bytes, so a test renders the whole site in memory.
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/yuin/goldmark"
)

// Options are the choices a build cannot derive from the sources.
type Options struct {
	// BaseURL is the origin the site is published on: scheme and host, with no
	// path and no trailing slash. Every absolute address on the site is built
	// from it, and so is the CNAME the host is claimed with.
	BaseURL string
	// SiteName is the product's name as the masthead and the sharing previews
	// spell it.
	SiteName string
	// ReferenceLocale is the locale every other one is measured against: it
	// supplies the language of the apex page and the target of hreflang
	// x-default.
	ReferenceLocale string
}

// File is one rendered file: where it goes below the site root, and what is in
// it.
type File struct {
	// Path is relative to the site root and always uses forward slashes.
	Path string
	// Data is the file's whole content.
	Data []byte
}

const (
	contentDir   = "content"
	templatesDir = "templates"
	assetsDir    = "assets"

	// indexPage is the page name that becomes a locale's own front page rather
	// than a directory beneath it.
	indexPage = "index"
)

// Build renders the site below source into the files a host serves. The result
// is sorted by path, so two builds of the same sources are byte-identical.
func Build(source fs.FS, opt Options) ([]File, error) {
	origin, err := parseBaseURL(opt.BaseURL)
	if err != nil {
		return nil, err
	}
	if opt.SiteName == "" {
		return nil, fmt.Errorf("render: site name is empty")
	}
	if opt.ReferenceLocale == "" {
		return nil, fmt.Errorf("render: reference locale is empty")
	}

	templates, err := template.ParseFS(source, path.Join(templatesDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("render: parse templates: %w", err)
	}

	pages, err := readPages(source)
	if err != nil {
		return nil, err
	}
	if _, ok := pages.byLocale[opt.ReferenceLocale]; !ok {
		return nil, fmt.Errorf("render: reference locale %q has no content", opt.ReferenceLocale)
	}

	files, err := renderPages(templates, pages, opt)
	if err != nil {
		return nil, err
	}

	root, err := renderRoot(templates, pages, opt)
	if err != nil {
		return nil, err
	}
	files = append(files, root)

	assets, err := copyAssets(source)
	if err != nil {
		return nil, err
	}
	files = append(files, assets...)
	files = append(files, metadataFiles(pages, opt, origin)...)

	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	return files, nil
}

// parseBaseURL checks that a base URL is the bare origin every address on the
// site is built from, and returns it.
func parseBaseURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, fmt.Errorf("render: base URL is empty")
	}
	if strings.HasSuffix(raw, "/") {
		return nil, fmt.Errorf("render: base URL %q ends in a slash", raw)
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("render: parse base URL %q: %w", raw, err)
	}
	switch {
	case parsed.Scheme != "https" && parsed.Scheme != "http":
		return nil, fmt.Errorf("render: base URL %q is neither https nor http", raw)
	case parsed.Host == "":
		return nil, fmt.Errorf("render: base URL %q has no host", raw)
	case parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "":
		return nil, fmt.Errorf("render: base URL %q carries more than a scheme and a host", raw)
	}
	return parsed, nil
}

// pageSet is every source text the site has, indexed the two ways a build needs
// to walk it: by locale to render, and by page name to cross-link translations.
type pageSet struct {
	locales  []string
	names    []string
	byLocale map[string]map[string]document
}

// readPages loads every locale's texts from the content directory.
func readPages(source fs.FS) (pageSet, error) {
	entries, err := fs.ReadDir(source, contentDir)
	if err != nil {
		return pageSet{}, fmt.Errorf("render: read content: %w", err)
	}

	set := pageSet{byLocale: make(map[string]map[string]document)}
	names := make(map[string]bool)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		locale := entry.Name()
		documents, err := readLocale(source, locale)
		if err != nil {
			return pageSet{}, err
		}
		set.locales = append(set.locales, locale)
		set.byLocale[locale] = documents
		for name := range documents {
			names[name] = true
		}
	}

	if len(set.locales) == 0 {
		return pageSet{}, fmt.Errorf("render: content holds no locale")
	}
	for name := range names {
		set.names = append(set.names, name)
	}
	sort.Strings(set.locales)
	sort.Strings(set.names)
	return set, nil
}

// readLocale loads one locale's texts, keyed by page name.
func readLocale(source fs.FS, locale string) (map[string]document, error) {
	entries, err := fs.ReadDir(source, path.Join(contentDir, locale))
	if err != nil {
		return nil, fmt.Errorf("render: read locale %q: %w", locale, err)
	}

	documents := make(map[string]document)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		file := path.Join(contentDir, locale, entry.Name())
		raw, err := fs.ReadFile(source, file)
		if err != nil {
			return nil, fmt.Errorf("render: read %s: %w", file, err)
		}
		doc, err := parseDocument(raw)
		if err != nil {
			return nil, fmt.Errorf("render: %s: %w", file, err)
		}
		documents[strings.TrimSuffix(entry.Name(), ".md")] = doc
	}

	if len(documents) == 0 {
		return nil, fmt.Errorf("render: locale %q holds no text", locale)
	}
	return documents, nil
}

// alternate is one hreflang link: a translation of the page, or the x-default
// that a reader with no matching language is sent to.
type alternate struct {
	Hreflang string
	URL      string
}

// languageLink is one entry of the language switcher in the masthead.
type languageLink struct {
	Hreflang string
	Name     string
	URL      string
	Current  bool
}

// pageData is everything a template needs to lay out one page.
type pageData struct {
	Lang        string
	Dir         string
	Title       string
	Description string
	SiteName    string
	Canonical   string
	HomeURL     string
	Body        template.HTML
	Alternates  []alternate
	Languages   []languageLink
}

// renderPages lays out every locale's every page.
func renderPages(templates *template.Template, pages pageSet, opt Options) ([]File, error) {
	var files []File
	for _, locale := range pages.locales {
		for name, doc := range pages.byLocale[locale] {
			body, err := markdown(doc.body)
			if err != nil {
				return nil, fmt.Errorf("render: %s/%s: %w", locale, name, err)
			}

			data := pageData{
				Lang:        locale,
				Dir:         direction(locale),
				Title:       doc.title,
				Description: doc.description,
				SiteName:    opt.SiteName,
				Canonical:   opt.BaseURL + pageURL(locale, name),
				HomeURL:     pageURL(locale, indexPage),
				Body:        body,
				Alternates:  alternatesFor(pages, opt, name),
				Languages:   languagesFor(pages, name, locale),
			}

			rendered, err := execute(templates, "page.html", &data)
			if err != nil {
				return nil, err
			}
			files = append(files, File{Path: outputPath(pageURL(locale, name)), Data: rendered})
		}
	}
	return files, nil
}

// renderRoot lays out the page the apex serves. It is a doorway rather than a
// translation: it names the product, says what it is, and hands the reader the
// languages it exists in. A bare redirect would be cheaper, but the apex is the
// address the product is listed under, and a redirect is a poor thing to list.
func renderRoot(templates *template.Template, pages pageSet, opt Options) (File, error) {
	reference := pages.byLocale[opt.ReferenceLocale][indexPage]
	data := pageData{
		Lang:        opt.ReferenceLocale,
		Dir:         direction(opt.ReferenceLocale),
		Title:       reference.title,
		Description: reference.description,
		SiteName:    opt.SiteName,
		Canonical:   opt.BaseURL + "/",
		HomeURL:     "/",
		Alternates:  alternatesFor(pages, opt, indexPage),
		Languages:   languagesFor(pages, indexPage, ""),
	}

	rendered, err := execute(templates, "root.html", &data)
	if err != nil {
		return File{}, err
	}
	return File{Path: "index.html", Data: rendered}, nil
}

// execute runs one template and reports which one failed when it does.
func execute(templates *template.Template, name string, data *pageData) ([]byte, error) {
	var out bytes.Buffer
	if err := templates.ExecuteTemplate(&out, name, data); err != nil {
		return nil, fmt.Errorf("render: execute %s: %w", name, err)
	}
	return out.Bytes(), nil
}

// markdown converts a source body to HTML.
func markdown(source []byte) (template.HTML, error) {
	var out bytes.Buffer
	if err := goldmark.New().Convert(source, &out); err != nil {
		return "", fmt.Errorf("convert markdown: %w", err)
	}
	return template.HTML(out.String()), nil //nolint:gosec // the source is the site's own text, and goldmark escapes what it does not own
}

// alternatesFor lists every translation of a page, plus the x-default a reader
// with no matching language is sent to. For a locale's front page that default
// is the apex; for anything else it is the reference locale's version, which is
// the only copy guaranteed to exist.
func alternatesFor(pages pageSet, opt Options, name string) []alternate {
	alternates := make([]alternate, 0, len(pages.locales)+1)
	for _, locale := range pages.locales {
		if _, ok := pages.byLocale[locale][name]; !ok {
			continue
		}
		alternates = append(alternates, alternate{
			Hreflang: locale,
			URL:      opt.BaseURL + pageURL(locale, name),
		})
	}

	fallback := opt.BaseURL + pageURL(opt.ReferenceLocale, name)
	if name == indexPage {
		fallback = opt.BaseURL + "/"
	}
	return append(alternates, alternate{Hreflang: "x-default", URL: fallback})
}

// languagesFor builds the language switcher for one page. A locale that lacks
// this page is left out: a switcher that leads to a missing file is worse than
// one language fewer.
func languagesFor(pages pageSet, name, current string) []languageLink {
	links := make([]languageLink, 0, len(pages.locales))
	for _, locale := range pages.locales {
		if _, ok := pages.byLocale[locale][name]; !ok {
			continue
		}
		links = append(links, languageLink{
			Hreflang: locale,
			Name:     languageName(locale),
			URL:      pageURL(locale, name),
			Current:  locale == current,
		})
	}
	return links
}

// pageURL is the address a page is served at. Every page is a directory, so an
// address never carries a file extension and never has to change when the
// renderer behind it does.
func pageURL(locale, name string) string {
	if name == indexPage {
		return "/" + locale + "/"
	}
	return "/" + locale + "/" + name + "/"
}

// outputPath turns an address into the file that serves it.
func outputPath(address string) string {
	return strings.TrimPrefix(address, "/") + "index.html"
}

// copyAssets carries the static files through untouched.
func copyAssets(source fs.FS) ([]File, error) {
	var files []File
	err := fs.WalkDir(source, assetsDir, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := fs.ReadFile(source, name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		files = append(files, File{Path: name, Data: data})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("render: copy assets: %w", err)
	}
	return files, nil
}

// metadataFiles are the files no reader opens: the ones that tell a crawler and
// the host what this site is.
func metadataFiles(pages pageSet, opt Options, origin *url.URL) []File {
	return []File{
		// The host serves whatever is handed to it, so the domain is claimed by
		// a file rather than by a setting somebody has to remember.
		{Path: "CNAME", Data: []byte(origin.Host + "\n")},
		{Path: "robots.txt", Data: []byte("User-agent: *\nAllow: /\n\nSitemap: " + opt.BaseURL + "/sitemap.xml\n")},
		{Path: "sitemap.xml", Data: sitemap(pages, opt)},
		// Static hosting with a build of its own does not run Jekyll, but this
		// costs one empty file and removes the question.
		{Path: ".nojekyll", Data: nil},
	}
}

// sitemap lists every address the site serves, each with its translations, so a
// crawler learns the whole language matrix from one file.
func sitemap(pages pageSet, opt Options) []byte {
	var out bytes.Buffer
	out.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	out.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"` +
		` xmlns:xhtml="http://www.w3.org/1999/xhtml">` + "\n")

	write := func(address string, alternates []alternate) {
		out.WriteString("  <url>\n    <loc>" + address + "</loc>\n")
		for _, alt := range alternates {
			out.WriteString(`    <xhtml:link rel="alternate" hreflang="` + alt.Hreflang +
				`" href="` + alt.URL + `"/>` + "\n")
		}
		out.WriteString("  </url>\n")
	}

	write(opt.BaseURL+"/", alternatesFor(pages, opt, indexPage))
	for _, locale := range pages.locales {
		for _, name := range pages.names {
			if _, ok := pages.byLocale[locale][name]; !ok {
				continue
			}
			write(opt.BaseURL+pageURL(locale, name), alternatesFor(pages, opt, name))
		}
	}

	out.WriteString("</urlset>\n")
	return out.Bytes()
}
