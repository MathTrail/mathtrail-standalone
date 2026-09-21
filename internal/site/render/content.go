package render

import (
	"bytes"
	"fmt"
	"strings"
)

// document is one source text: the fields the layout needs, and the body that
// follows them.
type document struct {
	title       string
	description string
	body        []byte
}

const frontMatterFence = "---"

// parseDocument splits a source file into its front matter and its body. The
// front matter is a fenced block of "key: value" lines at the top of the file;
// it is required, because a page with no title or description cannot be
// rendered into something a search engine or a chat will show correctly.
func parseDocument(source []byte) (document, error) {
	rest, ok := bytes.CutPrefix(source, []byte(frontMatterFence+"\n"))
	if !ok {
		return document{}, fmt.Errorf("front matter must open with %q on the first line", frontMatterFence)
	}

	header, body, ok := bytes.Cut(rest, []byte("\n"+frontMatterFence+"\n"))
	if !ok {
		return document{}, fmt.Errorf("front matter is never closed by %q on a line of its own", frontMatterFence)
	}

	doc := document{body: body}
	for _, line := range strings.Split(string(header), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return document{}, fmt.Errorf("front matter line %q is not key: value", line)
		}
		key, value = strings.TrimSpace(key), strings.TrimSpace(value)
		switch key {
		case "title":
			doc.title = value
		case "description":
			doc.description = value
		default:
			return document{}, fmt.Errorf("front matter key %q is not one this site understands", key)
		}
	}

	if doc.title == "" {
		return document{}, fmt.Errorf("front matter has no title")
	}
	if doc.description == "" {
		return document{}, fmt.Errorf("front matter has no description")
	}
	return doc, nil
}
