package mcpserver

import (
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
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

// pageLanguage is the language the topics' pages are linked in for the
// child's lessons: the language the parent chose, read as any tag is read,
// when the site is written in it, else the first language it narrows down to
// that the site is written in — pt-BR to pt —, and with none, or no language
// chosen, the one every page is in.
func pageLanguage(student *profile.Student) string {
	chosen, ok := student.ChosenLanguage()
	if !ok {
		return everyPageLanguage
	}
	for wanted := chosen; wanted != ""; {
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
	if !topic.SitePage || topic.Slug == "" {
		return ""
	}
	return site + "/" + language + "/topics/" + topic.Slug + "/" + anchor
}
