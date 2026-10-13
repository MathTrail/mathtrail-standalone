package mcpserver

import (
	"slices"
	"strings"

	"golang.org/x/text/language"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// siteLanguages are the languages the site is written in, every page of it in
// each: those of its dictionaries.
var siteLanguages = []string{"ar", "de", "en", "es", "fr", "hi", "ja", "pt", "ru", "zh-Hans"}

// everyPageLanguage is the language every page of the site is written in, the
// one a lesson in a language the site does not speak is linked to.
const everyPageLanguage = "en"

// The parts of a topic's page a link may lead to: where it tells the mistakes
// made in the topic, and how to help at home.
const (
	anchorTraps = "#traps"
	anchorHome  = "#home"
)

// Norwegian named as a whole is read as Bokmål, the standard most of its
// readers write; Nynorsk is a standard of its own.
var (
	norwegian = language.MustParseBase("no")
	bokmal    = language.MustParseBase("nb")
)

// pageLanguage is the language the topics' pages are linked in for the
// child's lessons: the language the parent chose, when the site is written in
// it as a card would read it, and with none, or no language chosen, the one
// every page is in.
func pageLanguage(student *profile.Student) string {
	chosen, ok := student.ChosenLanguage()
	if !ok {
		return everyPageLanguage
	}
	return languageAmong(siteLanguages, chosen)
}

// languageAmong is the language of languages that chosen is read as, the way
// a card finds its dictionary, or the one every page is in when it is read as
// none of them. The tag is tried whole and then shorter — pt-BR, then pt —,
// and then with the script its language is usually written in, so that zh-CN
// finds zh-Hans; but never shorter than its script, so that zh-TW is not
// answered in the simplified characters a bare zh stands for. A language of a
// macrolanguage is read as it, cmn as zh, as the card's platform reads it;
// Norwegian named as a whole, no, is read as Bokmål, nb; a tag that names no
// language, und-RU, is read as none, however likely a language its region
// makes; and the tag's extensions, such as a numbering system, say nothing of
// the language and are left out.
func languageAmong(languages []string, chosen string) string {
	tag, err := language.All.Parse(chosen)
	if err != nil {
		return everyPageLanguage
	}
	if _, confidence := tag.Base(); confidence != language.Exact {
		return everyPageLanguage
	}
	written, likely := spellingsOf(tag)
	script := scriptOf(likely)
	for _, candidate := range slices.Concat(shorterTags(written), shorterTags(likely)) {
		if slices.Contains(languages, candidate) && scriptOf(candidate) == script {
			return candidate
		}
	}
	return everyPageLanguage
}

// spellingsOf are the two ways a chosen language is tried: as it was written,
// with no extensions and Norwegian as Bokmål, and with the script and the
// region its language is likely written and spoken in added.
func spellingsOf(tag language.Tag) (written, likely string) {
	base, _ := tag.Base()
	if base == norwegian {
		base = bokmal
	}
	as := []any{base}
	maximized := []any{base}
	if script, confidence := tag.Script(); confidence != language.No {
		maximized = append(maximized, script)
		if confidence == language.Exact {
			as = append(as, script)
		}
	}
	if region, confidence := tag.Region(); confidence != language.No {
		maximized = append(maximized, region)
		if confidence == language.Exact {
			as = append(as, region)
		}
	}
	as = append(as, tag.Variants())
	return composed(as), composed(maximized)
}

// composed is the tag the parts make, as BCP 47 writes it; parts read from a
// tag that parsed always make one.
func composed(parts []any) string {
	tag, err := language.Compose(parts...)
	if err != nil {
		return ""
	}
	return tag.String()
}

// scriptOf is the script a tag names, or the one its language is usually
// written in.
func scriptOf(tag string) language.Script {
	script, _ := language.Make(tag).Script()
	return script
}

// shorterTags are a tag and every tag its end can be cut to: zh-Hans-CN,
// zh-Hans, zh.
func shorterTags(tag string) []string {
	var tags []string
	for cut := tag; cut != ""; {
		tags = append(tags, cut)
		end := strings.LastIndex(cut, "-")
		if end < 0 {
			break
		}
		cut = cut[:end]
	}
	return tags
}

// pageAddress is the address of topic's page on site in locale, at anchor —
// none for the top of the page —, or nothing when the topic has no page there
// yet. Nothing of the child is in it: the site, the language, the topic.
func pageAddress(site, locale string, topic *content.Topic, anchor string) string {
	if !topic.SitePage || topic.Slug == "" {
		return ""
	}
	return site + "/" + locale + "/topics/" + topic.Slug + "/" + anchor
}
