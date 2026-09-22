// Package check inspects a built site and reports what a browser, a crawler or
// a reviewer would otherwise each find broken separately: a link that leads
// nowhere, a page that quietly fetches from somebody else's domain, a head that
// is missing what a search result is built from, a locale that lost a page, a
// page that grew past its budget.
//
// It reads the finished directory rather than the renderer that produced it, so
// it keeps working when the renderer is replaced.
package check

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

// Options are the expectations a built site cannot state about itself.
type Options struct {
	// BaseURL is the origin the site is published on, with no trailing slash.
	// Canonical and alternate addresses are measured against it.
	BaseURL string
	// ReferenceLocale is the locale whose set of pages every other locale must
	// match.
	ReferenceLocale string
	// MaxPageBytes caps a page together with the local files it pulls in. Zero
	// means the size is not checked.
	MaxPageBytes int64
}

// Finding is one thing that is wrong, named where a person can go and fix it.
type Finding struct {
	// Path is the file the finding is about, relative to the site root.
	Path string
	// Rule is the short name of the rule that was broken.
	Rule string
	// Message says what is wrong, in the words the fixer needs.
	Message string
}

// String renders a finding as one line.
func (f Finding) String() string {
	return fmt.Sprintf("%s: %s: %s", f.Path, f.Rule, f.Message)
}

// The rule names, kept together so that a report and a test agree on them.
const (
	RuleLink        = "link"
	RuleExternal    = "external"
	RuleHead        = "head"
	RuleTranslation = "translation"
	RuleWeight      = "weight"
)

const indexFile = "index.html"

// Run checks a built site and returns every finding, sorted by file and rule.
// An empty result means the site is publishable. The error is reserved for a
// site that could not be read at all.
func Run(site fs.FS, opt Options) ([]Finding, error) {
	if opt.BaseURL == "" {
		return nil, fmt.Errorf("check: base URL is empty")
	}
	if opt.ReferenceLocale == "" {
		return nil, fmt.Errorf("check: reference locale is empty")
	}

	files, err := readSizes(site)
	if err != nil {
		return nil, err
	}

	pages, err := readPages(site, files)
	if err != nil {
		return nil, err
	}
	if _, ok := pages[address(opt.ReferenceLocale, "")]; !ok {
		return nil, fmt.Errorf("check: reference locale %q has no front page", opt.ReferenceLocale)
	}

	var findings []Finding
	for _, page := range pages {
		findings = append(findings, checkLinks(page, files)...)
		findings = append(findings, checkExternal(page)...)
		findings = append(findings, checkHead(page, pages, opt)...)
		findings = append(findings, checkWeight(page, files, opt)...)
	}
	findings = append(findings, checkTranslations(pages, opt)...)
	findings = append(findings, checkStylesheets(site, files)...)

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		if findings[i].Rule != findings[j].Rule {
			return findings[i].Rule < findings[j].Rule
		}
		return findings[i].Message < findings[j].Message
	})
	return findings, nil
}

// readSizes lists every file in the site with its size, which is both the
// existence test for a link and the input to the weight budget.
func readSizes(site fs.FS) (map[string]int64, error) {
	sizes := make(map[string]int64)
	err := fs.WalkDir(site, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("stat %s: %w", name, err)
		}
		sizes[name] = info.Size()
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("check: read site: %w", err)
	}
	if len(sizes) == 0 {
		return nil, fmt.Errorf("check: the site is empty")
	}
	return sizes, nil
}

// readPages parses every HTML file in the site.
func readPages(site fs.FS, files map[string]int64) (map[string]*page, error) {
	pages := make(map[string]*page)
	for name := range files {
		if path.Base(name) != indexFile {
			continue
		}
		data, err := fs.ReadFile(site, name)
		if err != nil {
			return nil, fmt.Errorf("check: read %s: %w", name, err)
		}
		parsed, err := parsePage(name, data)
		if err != nil {
			return nil, fmt.Errorf("check: %s: %w", name, err)
		}
		pages[parsed.address] = parsed
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("check: the site holds no page")
	}
	return pages, nil
}

// checkLinks reports every reference that leads to a file the site does not
// have.
func checkLinks(p *page, files map[string]int64) []Finding {
	var findings []Finding
	for _, ref := range p.references {
		if ref.external() {
			continue
		}
		target := ref.resolve(p.file)
		if target == "" {
			continue
		}
		if _, ok := files[target]; !ok {
			findings = append(findings, Finding{
				Path:    p.file,
				Rule:    RuleLink,
				Message: fmt.Sprintf("<%s %s=%q> leads to %s, which the site does not have", ref.element, ref.attribute, ref.value, target),
			})
		}
	}
	return findings
}

// checkExternal reports anything the page fetches from another origin. A link a
// reader may follow is fine; a file the browser loads on its own is not, because
// it would hand a third party every visit to this page.
func checkExternal(p *page) []Finding {
	var findings []Finding
	for _, ref := range p.references {
		if !ref.subresource || !ref.external() {
			continue
		}
		findings = append(findings, Finding{
			Path:    p.file,
			Rule:    RuleExternal,
			Message: fmt.Sprintf("<%s %s=%q> loads from another origin", ref.element, ref.attribute, ref.value),
		})
	}
	return findings
}

// checkHead reports a head that a search result, a shared link or a screen
// reader could not be built from.
func checkHead(p *page, pages map[string]*page, opt Options) []Finding {
	var findings []Finding
	report := func(format string, args ...any) {
		findings = append(findings, Finding{Path: p.file, Rule: RuleHead, Message: fmt.Sprintf(format, args...)})
	}

	switch {
	case p.lang == "":
		report("<html> carries no lang")
	case p.locale != "" && p.lang != p.locale:
		report("<html lang=%q> disagrees with the locale %q the page is served under", p.lang, p.locale)
	}
	if p.dir == "" {
		report("<html> carries no dir")
	}
	if p.title == "" {
		report("the page has no title")
	}
	if p.description == "" {
		report("the page has no description")
	}

	if want := opt.BaseURL + p.address; p.canonical != want {
		report("canonical is %q, and the page is served at %q", p.canonical, want)
	}

	for _, want := range expectedAlternates(p, pages, opt) {
		if got, ok := p.alternates[want.Hreflang]; !ok {
			report("no alternate for %q, which the site has at %q", want.Hreflang, want.URL)
		} else if got != want.URL {
			report("alternate for %q is %q, and that page is at %q", want.Hreflang, got, want.URL)
		}
	}
	return findings
}

// alternate is one hreflang link a page is expected to carry.
type alternate struct {
	Hreflang string
	URL      string
}

// expectedAlternates works out which translations a page must point at, from
// the pages the site actually has.
func expectedAlternates(p *page, pages map[string]*page, opt Options) []alternate {
	var want []alternate
	for _, locale := range locales(pages) {
		if _, ok := pages[address(locale, p.name)]; !ok {
			continue
		}
		want = append(want, alternate{Hreflang: locale, URL: opt.BaseURL + address(locale, p.name)})
	}

	fallback := opt.BaseURL + address(opt.ReferenceLocale, p.name)
	if p.name == "" {
		fallback = opt.BaseURL + "/"
	}
	return append(want, alternate{Hreflang: "x-default", URL: fallback})
}

// checkTranslations reports a locale that is missing a page the reference
// locale has, which is how a language quietly loses its privacy policy.
func checkTranslations(pages map[string]*page, opt Options) []Finding {
	reference := sortedNames(pageNames(pages, opt.ReferenceLocale))

	var findings []Finding
	for _, locale := range locales(pages) {
		if locale == opt.ReferenceLocale {
			continue
		}
		have := pageNames(pages, locale)
		for _, name := range reference {
			if have[name] {
				continue
			}
			findings = append(findings, Finding{
				Path:    strings.TrimPrefix(address(locale, name), "/") + indexFile,
				Rule:    RuleTranslation,
				Message: fmt.Sprintf("locale %q is missing a page that %q has", locale, opt.ReferenceLocale),
			})
		}
	}
	return findings
}

// checkWeight reports a page that, with everything it pulls in, costs the
// reader more than the budget allows.
func checkWeight(p *page, files map[string]int64, opt Options) []Finding {
	if opt.MaxPageBytes == 0 {
		return nil
	}

	total := files[p.file]
	counted := map[string]bool{p.file: true}
	for _, ref := range p.references {
		if !ref.subresource || ref.external() {
			continue
		}
		target := ref.resolve(p.file)
		if target == "" || counted[target] {
			continue
		}
		counted[target] = true
		total += files[target]
	}

	if total <= opt.MaxPageBytes {
		return nil
	}
	return []Finding{{
		Path:    p.file,
		Rule:    RuleWeight,
		Message: fmt.Sprintf("the page and what it loads come to %d bytes, and the budget is %d", total, opt.MaxPageBytes),
	}}
}

// checkStylesheets reports a stylesheet that reaches for another origin. A font
// or an image pulled in from CSS is invisible in the HTML, and it would hand a
// third party every visit just as surely as a script tag.
func checkStylesheets(site fs.FS, files map[string]int64) []Finding {
	var findings []Finding
	for name := range files {
		if !strings.HasSuffix(name, ".css") {
			continue
		}
		data, err := fs.ReadFile(site, name)
		if err != nil {
			continue
		}
		// These are what is searched for, not what is used: a stylesheet that
		// names any of them is the finding.
		for _, marker := range []string{"http://", "https://", "url(//"} { // NOSONAR
			if !strings.Contains(string(data), marker) {
				continue
			}
			findings = append(findings, Finding{
				Path:    name,
				Rule:    RuleExternal,
				Message: fmt.Sprintf("the stylesheet mentions %q, so it may load from another origin", marker),
			})
		}
	}
	return findings
}

// locales lists the locales the site has, in a stable order.
func locales(pages map[string]*page) []string {
	seen := make(map[string]bool)
	for _, p := range pages {
		if p.locale != "" {
			seen[p.locale] = true
		}
	}
	list := make([]string, 0, len(seen))
	for locale := range seen {
		list = append(list, locale)
	}
	sort.Strings(list)
	return list
}

// pageNames lists the page names one locale has.
func pageNames(pages map[string]*page, locale string) map[string]bool {
	names := make(map[string]bool)
	for _, p := range pages {
		if p.locale == locale {
			names[p.name] = true
		}
	}
	return names
}

// address is the URL path a locale's page is served at.
func address(locale, name string) string {
	if name == "" {
		return "/" + locale + "/"
	}
	return "/" + locale + "/" + name + "/"
}

// sortedNames turns a set of page names into a stable list, so that a report
// reads the same way twice.
func sortedNames(set map[string]bool) []string {
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
