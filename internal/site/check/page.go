package check

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"golang.org/x/net/html"
)

// page is one parsed HTML file: where it sits, what its head promises, and
// everything it points at.
type page struct {
	// file is the path below the site root, such as "en/privacy/index.html".
	file string
	// address is the URL path the file is served at, such as "/en/privacy/".
	address string
	// locale is the first path segment, empty for the page the apex serves.
	locale string
	// name is the page within the locale, empty for the locale's front page.
	name string

	lang        string
	dir         string
	title       string
	description string
	canonical   string
	alternates  map[string]string

	references []reference
}

// reference is one address a page points at, and whether the browser fetches it
// on its own or only if the reader asks.
type reference struct {
	element   string
	attribute string
	value     string
	// subresource is true when the browser loads this without being asked,
	// which is what makes another origin a leak rather than a link.
	subresource bool
}

// external reports whether a reference leaves this site. A scheme or a
// protocol-relative prefix is the whole test: everything else is ours.
func (r reference) external() bool {
	if strings.HasPrefix(r.value, "//") {
		return true
	}
	scheme, _, found := strings.Cut(r.value, ":")
	if !found || scheme == "" {
		return false
	}
	for i := 0; i < len(scheme); i++ {
		if !schemeByte(scheme[i], i == 0) {
			return false
		}
	}
	return true
}

// schemeByte reports whether a byte may stand at this position of a URL scheme.
// Only a letter may open one, which is what keeps a Windows-style path and a
// bare colon in a filename from reading as a scheme.
func schemeByte(c byte, first bool) bool {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		return true
	case first:
		return false
	default:
		return c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.'
	}
}

// resolve returns the file a reference asks for, relative to the site root. An
// empty result means there is nothing to look for — an anchor within the page,
// or an empty attribute.
func (r reference) resolve(fromFile string) string {
	target := r.value
	if i := strings.IndexAny(target, "#?"); i >= 0 {
		target = target[:i]
	}
	if target == "" {
		return ""
	}

	if strings.HasPrefix(target, "/") {
		target = strings.TrimPrefix(target, "/")
	} else {
		target = path.Join(path.Dir(fromFile), target)
	}

	// Every page of this site is a directory, so an address with no file
	// extension is asking for the index inside one.
	if strings.HasSuffix(r.value, "/") || path.Ext(target) == "" {
		target = path.Join(target, indexFile)
	}
	return path.Clean(target)
}

// subresourceRels are the link relations a browser acts on by itself. A
// canonical or an alternate is a statement about the page, not a file to fetch.
var subresourceRels = map[string]bool{
	"stylesheet":       true,
	"icon":             true,
	"shortcut icon":    true,
	"apple-touch-icon": true,
	"manifest":         true,
	"preload":          true,
	"modulepreload":    true,
	"prefetch":         true,
}

// srcElements are the elements whose src the browser fetches on its own.
var srcElements = map[string]bool{
	"script": true,
	"img":    true,
	"iframe": true,
	"source": true,
	"video":  true,
	"audio":  true,
	"embed":  true,
	"track":  true,
	"object": true,
}

// parsePage reads one built HTML file.
func parsePage(file string, data []byte) (*page, error) {
	root, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	p := &page{file: file, alternates: make(map[string]string)}
	p.locale, p.name, p.address = locate(file)
	p.walk(root)
	return p, nil
}

// locate works out where a file sits in the site from its path alone.
func locate(file string) (locale, name, address string) {
	trimmed := strings.TrimSuffix(strings.TrimSuffix(file, indexFile), "/")
	if trimmed == "" {
		return "", "", "/"
	}
	locale, name, _ = strings.Cut(trimmed, "/")
	if name == "" {
		return locale, "", "/" + locale + "/"
	}
	return locale, name, "/" + locale + "/" + name + "/"
}

// walk reads the head and collects every reference in the document.
func (p *page) walk(node *html.Node) {
	if node.Type == html.ElementNode {
		p.readElement(node)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		p.walk(child)
	}
}

// readElement takes from one element whatever the checks need from it.
func (p *page) readElement(node *html.Node) {
	attr := func(name string) string {
		for _, a := range node.Attr {
			if a.Key == name {
				return a.Val
			}
		}
		return ""
	}

	switch node.Data {
	case "html":
		p.lang, p.dir = attr("lang"), attr("dir")
	case "title":
		if node.FirstChild != nil {
			p.title = strings.TrimSpace(node.FirstChild.Data)
		}
	case "meta":
		if strings.EqualFold(attr("name"), "description") {
			p.description = strings.TrimSpace(attr("content"))
		}
	case "a":
		p.reference(node, "href", false)
	case "link":
		p.readLink(node, attr)
	}

	if srcElements[node.Data] {
		p.reference(node, "src", true)
	}
}

// readLink sorts a link element into the three things it can be: a statement of
// this page's own address, a pointer at a translation, or a file to fetch.
func (p *page) readLink(node *html.Node, attr func(string) string) {
	rel := strings.ToLower(strings.TrimSpace(attr("rel")))
	switch rel {
	case "canonical":
		p.canonical = attr("href")
	case "alternate":
		if lang := attr("hreflang"); lang != "" {
			p.alternates[lang] = attr("href")
		}
	default:
		p.reference(node, "href", subresourceRels[rel])
	}
}

// reference records one address the element points at.
func (p *page) reference(node *html.Node, attribute string, subresource bool) {
	for _, a := range node.Attr {
		if a.Key != attribute || a.Val == "" {
			continue
		}
		p.references = append(p.references, reference{
			element:     node.Data,
			attribute:   attribute,
			value:       a.Val,
			subresource: subresource,
		})
	}
}
