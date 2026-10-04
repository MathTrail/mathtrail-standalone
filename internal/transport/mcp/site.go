package mcpserver

import (
	"net"
	"net/url"
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/content"
)

// siteLanguages are the languages the site is written in, every page of it in
// each: those of its dictionaries.
var siteLanguages = []string{"en", "ru"}

// everyPageLanguage is the language every page of the site is written in, the
// one a lesson in a language the site does not speak is linked to.
const everyPageLanguage = "en"

// The parts of a topic's page a link may lead to: where it tells the mistakes
// made in the topic, and how to help at home.
const (
	anchorTraps = "#traps"
	anchorHome  = "#home"
)

// defaultPorts are the ports a browser leaves out of an origin, by scheme.
var defaultPorts = map[string]string{"https": "443", "http": "80"}

// siteOf is address as the origin of the site the topics' pages are on — its
// scheme and its host, in lower case and with no port a browser would leave
// out — and whether address is one at all: an address with a path, a query, a
// fragment or credentials names no site. The card holds an address it links to
// to the exact origin, so the words and the card name the site alike.
func siteOf(address string) (string, bool) {
	parsed, err := url.Parse(address)
	if err != nil || parsed.Host == "" || parsed.User != nil ||
		(parsed.Path != "" && parsed.Path != "/") || strings.ContainsAny(address, "?#") {
		return "", false
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "https" && scheme != "http" {
		return "", false
	}
	host := strings.ToLower(parsed.Hostname())
	if port := parsed.Port(); port != "" && port != defaultPorts[scheme] {
		host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return scheme + "://" + host, true
}

// pageLanguage is the language the topics' pages are linked in for a lesson
// in tag: the tag's own when the site is written in it, else the first
// language it narrows down to that the site is written in — pt-BR to pt —,
// and with none, or no lesson language chosen, the one every page is in.
func pageLanguage(tag *string) string {
	if tag == nil {
		return everyPageLanguage
	}
	for wanted := *tag; wanted != ""; {
		if slices.Contains(siteLanguages, wanted) {
			return wanted
		}
		cut := strings.LastIndex(wanted, "-")
		if cut < 0 {
			break
		}
		wanted = wanted[:cut]
	}
	return everyPageLanguage
}

// pageAddress is the address of topic's page on site in language, at anchor —
// none for the top of the page —, or nothing when the topic has no page there
// yet. Nothing of the child is in it: the site, the language, the topic.
func pageAddress(site, language string, topic *content.Topic, anchor string) string {
	if site == "" || !topic.SitePage || topic.Slug == "" {
		return ""
	}
	return site + "/" + language + "/topics/" + topic.Slug + "/" + anchor
}
