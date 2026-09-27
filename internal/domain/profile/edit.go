package profile

import (
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/language"
)

// Edit is a change to the child's details as the parent asked for it, by way
// of the chat's model. A field left nil stays as it is. A list that is given
// replaces the one kept, so an empty list clears it; an empty text clears the
// notes, and an empty language makes the cards follow the chat's language
// again.
type Edit struct {
	Pseudonym      *string
	Grade          *int
	Interests      []string
	ExcludedSkills []string
	Notes          *string
	UILanguage     *string
}

// NewStudent is the child's details as the parent first gave them, or the
// problems that keep them from being kept. A pseudonym and a grade are what
// a profile cannot be made without; everything else may start empty.
//
// A skill is kept only when known names it: the ids come from a model, and a
// skill the catalog does not have could keep nothing out of a task.
func NewStudent(e *Edit, known func(skill string) bool) (Student, []Problem) {
	blank := Student{ExcludedSkills: []string{}, Interests: []string{}}
	return blank.edited(e, known)
}

// Change makes the parent's edit to the child's details, and reports whether
// it changed anything. An edit with problems changes nothing, and neither
// does one that asks for what is already there.
//
// A change is a write: the profile is touched with it, so its revision moves
// on exactly when there is something to write. Only the details move. The
// grade among them is a label once the profile exists — the start, the level
// and the trial series stay where the answers put them.
func (p *Profile) Change(e *Edit, known func(skill string) bool, appVersion string, now time.Time) (bool, []Problem) {
	edited, problems := p.Student.edited(e, known)
	if len(problems) > 0 || edited.same(&p.Student) {
		return false, problems
	}
	p.Student = edited
	p.Touch(appVersion, now)
	return true, nil
}

// edited is the details with the edit made and cleaned, and the problems that
// keep them from being kept, sorted by field. Two rules are the edit's own and
// not the file's: a skill must be in the catalog, and a language must be one.
// A file somebody edited by hand is not made unreadable by either.
func (s *Student) edited(e *Edit, known func(skill string) bool) (Student, []Problem) {
	next := s.copied()
	if e.Pseudonym != nil {
		next.Pseudonym = Typed(*e.Pseudonym)
	}
	if e.Grade != nil {
		next.Grade = *e.Grade
	}
	if e.Notes != nil {
		next.Notes = Typed(*e.Notes)
	}
	if e.Interests != nil {
		next.Interests = typedList(e.Interests)
	}
	found := slices.Concat(
		next.editSkills(e.ExcludedSkills, known),
		next.editLanguage(e.UILanguage),
		next.problems(),
	)
	slices.SortStableFunc(found, func(a, b Problem) int { return strings.Compare(a.Field, b.Field) })
	return next, found
}

// editSkills puts the skills given in place, or says why they cannot be. A
// list merged of its repeats is counted by the rules of the file afterwards, so
// every skill of the catalog with one of them given twice still fits.
func (s *Student) editSkills(given []string, known func(skill string) bool) []Problem {
	if given == nil {
		return nil
	}
	// Checked where each entry stood in what was sent, so that the entry a
	// refusal names is the one the model wrote there.
	cleaned := typedEach(given)
	if rule := catalogRule(cleaned, known); rule != "" {
		return []Problem{{Field: "excluded_skills", Rule: rule}}
	}
	s.ExcludedSkills = distinct(cleaned)
	return nil
}

// editLanguage puts the language of the cards in place, or says why it cannot
// be.
func (s *Student) editLanguage(given *string) []Problem {
	if given == nil {
		return nil
	}
	tag, rule := LanguageTag(*given)
	switch {
	case rule != "":
		return []Problem{{Field: "ui_language", Rule: rule}}
	case tag == "":
		s.UILanguage = nil
	default:
		s.UILanguage = &tag
	}
	return nil
}

// copied is the details sharing nothing with the ones they came from, and with
// a list for every list: a file says "none" as an empty list, never as null,
// because what the rule copies from it reaches the model as it stands.
func (s *Student) copied() Student {
	next := *s
	next.ExcludedSkills = append([]string{}, s.ExcludedSkills...)
	next.Interests = append([]string{}, s.Interests...)
	if s.UILanguage != nil {
		chosen := *s.UILanguage
		next.UILanguage = &chosen
	}
	return next
}

// same reports whether two sets of details say the same thing. The interests
// are the same in the same order only, since the order is the order a task is
// dressed in them; the skills left out are a set, and the same skills in
// another order are the same skills.
func (s *Student) same(other *Student) bool {
	return s.Pseudonym == other.Pseudonym && s.Grade == other.Grade && s.Notes == other.Notes &&
		slices.Equal(s.Interests, other.Interests) &&
		slices.Equal(sorted(s.ExcludedSkills), sorted(other.ExcludedSkills)) &&
		sameLanguage(s.UILanguage, other.UILanguage)
}

// sorted is a list in order, sharing nothing with it.
func sorted(list []string) []string {
	copied := slices.Clone(list)
	slices.Sort(copied)
	return copied
}

func sameLanguage(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

// Typed is a text as the one who typed it meant it: a space of any kind — a
// tab, a line or paragraph break, a space of another width — reads as a plain
// one, any other control character is dropped, and so is a character that
// shows nothing at all or makes a text show in an order other than the one it
// is kept in. Nothing but the words is kept at either end, so a name made of
// nothing anybody can see is no name.
//
// The parent's notes are the text this matters for most: the one thing a
// person writes and a model reads. They reach the model as a string of JSON,
// which nothing inside it can end, introduced as information about the child.
// The reason a model gives for a choice of its own is read the same way, since
// it is kept in the file and handed back with the brief. The joiners that some
// scripts and emoji are written with stay, since without them the words are
// not the words.
func Typed(text string) string {
	cleaned := strings.Map(func(r rune) rune {
		switch {
		case unicode.IsSpace(r):
			// A tab, a line break, a paragraph break or a space of any width
			// is a space.
			return ' '
		case unicode.IsControl(r), unseen(r):
			return -1
		}
		return r
	}, text)
	return strings.TrimSpace(cleaned)
}

// unseen reports whether a character shows nothing, or only steers how the
// text around it is shown: every character of format — a zero-width space, a
// soft hyphen, a byte-order mark, the marks, embeddings, overrides and isolates
// of direction, the tags that can spell out a whole hidden sentence — every
// other character Unicode says to ignore by default, and the selectors of a
// glyph's variant. None of them changes what the text says, and a note about a
// child has no use for a sentence nobody can see. The two joiners some scripts
// and emoji are written with are the exception: without them the words are not
// the words.
func unseen(r rune) bool {
	if r == '\u200c' || r == '\u200d' {
		return false
	}
	return unicode.In(r, unicode.Cf, unicode.Other_Default_Ignorable_Code_Point, unicode.Variation_Selector)
}

// typedList is a list of texts, each as the parent meant it, with the same
// text kept once.
func typedList(texts []string) []string { return distinct(typedEach(texts)) }

// typedEach is every text of a list as the parent meant it, each in its place.
func typedEach(texts []string) []string {
	cleaned := make([]string, 0, len(texts))
	for _, text := range texts {
		cleaned = append(cleaned, Typed(text))
	}
	return cleaned
}

// distinct keeps the first of every repeated entry, in the order given.
func distinct(entries []string) []string {
	kept := make([]string, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if _, repeated := seen[entry]; !repeated {
			seen[entry] = struct{}{}
			kept = append(kept, entry)
		}
	}
	return kept
}

// catalogRule says which entry of a list of skills names no skill of the
// catalog, by its place rather than by what it says: a model resending a list
// that holds a skill the catalog has since dropped learns which one to leave
// out.
func catalogRule(skills []string, known func(skill string) bool) string {
	for place, skill := range skills {
		if !known(skill) {
			return fmt.Sprintf("must name skills of the catalog, by their ids; entry %d names none", place+1)
		}
	}
	return ""
}

// notLanguages are the codes of BCP 47 that name no language a card could be
// written in: undetermined, several, none at all, and one not yet coded.
var notLanguages = []language.Base{
	language.MustParseBase("und"),
	language.MustParseBase("mul"),
	language.MustParseBase("zxx"),
	language.MustParseBase("mis"),
}

// LanguageTag reads a language as the service keeps one: a BCP 47 tag, in its
// canonical spelling, so that pt-br and pt-BR are one language. The tag has to
// name its language outright: "und-US" only guesses one from a country, and a
// private tag names none. What it returns is the tag, or the rule the text
// broke, in words that never repeat the text. An empty text is no tag and
// breaks no rule here: what "none" means — the chat's language for the cards,
// a missing argument for a task — is the caller's to say.
func LanguageTag(text string) (tag, rule string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", ""
	}
	if count := utf8.RuneCountInString(text); count > MaxLanguageTag {
		return "", fmt.Sprintf("must be at most %d characters, not %d", MaxLanguageTag, count)
	}
	parsed, err := language.Parse(text)
	base, confidence := parsed.Base()
	if err != nil || confidence != language.Exact || slices.Contains(notLanguages, base) {
		return "", "must be a BCP 47 tag that names a language, such as en, ru or pt-BR"
	}
	return parsed.String(), ""
}
