package main

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// arXivDOIPrefix is where DataCite registers arXiv's preprints: arXiv 2503.16460
// is the DOI 10.48550/arXiv.2503.16460.
const arXivDOIPrefix = "10.48550/arXiv."

// DOIOf is the DOI an entry is checked by: its own, or the one arXiv registers
// for its preprint when it has none.
func DOIOf(e Entry) (string, error) {
	if doi := strings.TrimSpace(e.Get("doi")); doi != "" {
		return DOIFromID(doi), nil
	}
	isArXiv := strings.EqualFold(e.Get("archiveprefix"), "arxiv") || strings.EqualFold(e.Get("eprinttype"), "arxiv")
	if eprint := strings.TrimSpace(e.Get("eprint")); eprint != "" && isArXiv {
		return DOIFromID("arXiv:" + strings.TrimPrefix(strings.ToLower(eprint), "arxiv:")), nil
	}
	return "", errors.New("it has neither a DOI nor an arXiv id to check it by")
}

var arXivVersion = regexp.MustCompile(`v\d+$`)

// Check lists every way an entry disagrees with what the registry records about
// the work it cites. No problem means the entry may be cited as it stands.
func Check(e Entry, w *Work) []string {
	var problems []string
	problems = append(problems, checkTitle(e, w)...)
	problems = append(problems, checkAuthors(e, w)...)
	problems = append(problems, checkYear(e, w)...)
	problems = append(problems, checkVenue(e, w)...)
	for _, notice := range w.Notices {
		problems = append(problems, fmt.Sprintf("%s records a %s: the work is not to be cited as sound", w.Registry, notice))
	}
	return problems
}

func checkTitle(e Entry, w *Work) []string {
	title := normalize(e.Get("title"))
	if title == normalize(w.Title) || title == normalize(w.Title+" "+w.Subtitle) {
		return nil
	}
	return []string{fmt.Sprintf("title %q differs from %s's %q", e.Get("title"), w.Registry, joinTitle(w))}
}

func joinTitle(w *Work) string {
	if w.Subtitle == "" {
		return w.Title
	}
	return w.Title + ": " + w.Subtitle
}

// checkAuthors compares family names in order. An entry may end its list with
// "and others", and then only the authors it names are compared.
func checkAuthors(e Entry, w *Work) []string {
	if len(w.Authors) == 0 {
		return nil
	}
	listed, truncated := EntryAuthors(e.Get("author"))
	recorded := make([]string, len(w.Authors))
	for i, a := range w.Authors {
		recorded[i] = normalize(a.Family)
	}
	names := make([]string, len(listed))
	for i, a := range listed {
		names[i] = normalize(a)
	}
	switch {
	case truncated && len(names) > 0 && len(names) <= len(recorded) && slices.Equal(names, recorded[:len(names)]):
		return nil
	case !truncated && slices.Equal(names, recorded):
		return nil
	}
	return []string{fmt.Sprintf("authors %s differ from %s's %s", strings.Join(names, ", "), w.Registry, strings.Join(recorded, ", "))}
}

func checkYear(e Entry, w *Work) []string {
	if len(w.Years) == 0 {
		return nil
	}
	year, err := strconv.Atoi(strings.TrimSpace(e.Get("year")))
	if err == nil && slices.Contains(w.Years, year) {
		return nil
	}
	return []string{fmt.Sprintf("year %q is none of %s's %v", e.Get("year"), w.Registry, w.Years)}
}

// checkVenue compares the journal or the proceedings an entry names with the
// containers the registry records. The entry may abbreviate each word, as
// "Comput. Educ." abbreviates "Computers & Education", but must name as many
// words as the container, small words such as "of" and "and" aside. A venue the
// registry does not record cannot be checked, so it may not be claimed: a
// preprint that later appeared in a journal is cited by the journal's DOI.
func checkVenue(e Entry, w *Work) []string {
	venue := e.Get("journal")
	if venue == "" {
		venue = e.Get("booktitle")
	}
	if venue == "" {
		return nil
	}
	if len(w.Venues) == 0 {
		return []string{fmt.Sprintf("venue %q cannot be checked: %s records none for this work; cite the published version by its own DOI", venue, w.Registry)}
	}
	for _, v := range w.Venues {
		if abbreviates(venue, v) {
			return nil
		}
	}
	return []string{fmt.Sprintf("venue %q is none of %s's %q", venue, w.Registry, w.Venues)}
}

// smallWords are left out when a venue is compared word by word: abbreviated
// titles drop them.
var smallWords = []string{"a", "an", "the", "of", "and", "in", "on", "for", "to", "de", "der", "und"}

// abbreviates says whether a venue as an entry writes it names the container:
// the same significant words in the same order, each written in full or cut
// short from its start.
func abbreviates(written, container string) bool {
	short, full := significantWords(written), significantWords(container)
	if len(short) == 0 || len(short) != len(full) {
		return false
	}
	for i := range short {
		if !strings.HasPrefix(full[i], short[i]) {
			return false
		}
	}
	return true
}

func significantWords(s string) []string {
	var words []string
	for _, w := range strings.Fields(normalize(s)) {
		if !slices.Contains(smallWords, w) {
			words = append(words, w)
		}
	}
	return words
}

// EntryAuthors splits an author field into family names, in order, and says
// whether it ends with "and others". A name is "Family, Given", or "Given
// Family" with the family name last; a braced name — an organisation — stays
// whole, even when it contains the word "and".
func EntryAuthors(field string) (families []string, truncated bool) {
	for _, name := range splitAuthors(field) {
		name = strings.TrimSpace(name)
		switch {
		case name == "":
		case strings.EqualFold(name, "others"):
			truncated = true
		default:
			families = append(families, familyName(name))
		}
	}
	return families, truncated
}

// wholeBraced says whether one pair of braces holds the whole name, as it holds
// an organisation's: its first brace closes at its last character.
func wholeBraced(name string) bool {
	if !strings.HasPrefix(name, "{") {
		return false
	}
	depth := 0
	for i := 0; i < len(name); i++ {
		switch name[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i == len(name)-1
			}
		}
	}
	return false
}

// depthZeroComma is where the first comma outside braces stands, or -1: a braced
// organisation may hold commas that part no family name from a given one.
func depthZeroComma(s string) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
		case ',':
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// splitAuthors splits an author field at every " and " outside braces.
func splitAuthors(field string) []string {
	var names []string
	depth, start := 0, 0
	for i := 0; i < len(field); i++ {
		switch field[i] {
		case '{':
			depth++
		case '}':
			depth--
		}
		if depth == 0 && isAnd(field, i) {
			names = append(names, field[start:i])
			start = i + len(" and ")
			i = start - 1
		}
	}
	return append(names, field[start:])
}

// isAnd says whether the word "and", with white space on both sides, starts at i.
func isAnd(field string, i int) bool {
	return i+5 <= len(field) && isSpace(field[i]) && strings.EqualFold(field[i+1:i+4], "and") && isSpace(field[i+4])
}

func isSpace(c byte) bool { return c == ' ' || c == '\t' || c == '\n' || c == '\r' }

func familyName(name string) string {
	if wholeBraced(name) {
		return name[1 : len(name)-1]
	}
	if comma := depthZeroComma(name); comma >= 0 {
		return strings.Trim(strings.TrimSpace(name[:comma]), "{}")
	}
	words := strings.Fields(name)
	return strings.Trim(words[len(words)-1], "{}")
}

// latexAccent is an accent command such as \"u, \'{e} or \c{c}: normalize keeps the
// letter and drops the accent, as it drops a Unicode one.
var latexAccent = regexp.MustCompile("\\\\(?:[\"'`^~=.]|[cvuHkrbd](?:\\s|\\b))\\s*\\{?([A-Za-z])\\}?")

var latexEscape = strings.NewReplacer(`\&`, "&", `\%`, "%", `\$`, "$", `\#`, "#", `\_`, "_",
	`\ss`, "ss", `\o`, "o", `\O`, "O", `\l`, "l", `\L`, "L", `\i`, "i", `\aa`, "a", `\ae`, "ae", "{", "", "}", "", "--", "-")

// normalize is the form two strings are compared in: LaTeX escapes and braces
// undone, HTML entities decoded, accents dropped, case folded, "&" read as
// "and", and punctuation and spacing ignored. "Calò" and "Calo" then match, as
// do "Computers &amp; Education" and "Computers \& Education".
func normalize(s string) string {
	s = latexEscape.Replace(latexAccent.ReplaceAllString(clean(s), "$1"))
	s = strings.ReplaceAll(s, "&", " and ")
	var b strings.Builder
	for _, r := range norm.NFKD.String(s) {
		switch {
		case unicode.Is(unicode.Mn, r):
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
		default:
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
