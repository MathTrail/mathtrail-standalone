package main

import (
	"slices"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// EntryFor writes the entry refs.bib keeps for a work, from its record alone,
// so that no field comes from memory. The key is the first author's family
// name, the year and the first word of the title that says something, as in
// klinkenberg2011computer; taken says which keys the file already uses, and a
// letter is added until the key is free.
func EntryFor(w *Work, taken func(key string) bool) Entry {
	entry := Entry{Type: bibType(w)}
	add := func(name, value string) {
		if value = strings.TrimSpace(value); value != "" {
			entry.Fields = append(entry.Fields, Field{Name: name, Value: value})
		}
	}
	add("author", authorField(w.Authors))
	add("title", protectAcronyms(escapeLaTeX(tidyTitle(joinTitle(w)))))
	switch entry.Type {
	case "article":
		add("journal", escapeLaTeX(first(w.Venues)))
	case "incollection", "inproceedings":
		add("booktitle", escapeLaTeX(last(w.Venues)))
		if len(w.Venues) > 1 {
			add("series", escapeLaTeX(first(w.Venues)))
		}
		add("publisher", escapeLaTeX(w.Publisher))
	case "book":
		add("publisher", escapeLaTeX(w.Publisher))
	}
	year := earliest(w.Years)
	if year > 0 {
		add("year", strconv.Itoa(year))
	}
	add("volume", w.Volume)
	add("number", w.Issue)
	add("pages", strings.ReplaceAll(w.Pages, "-", "--"))
	add("doi", w.DOI)
	if id, isArXiv := strings.CutPrefix(strings.ToLower(w.DOI), strings.ToLower(arXivDOIPrefix)); isArXiv {
		add("eprint", id)
		add("archiveprefix", "arXiv")
	} else if entry.Type == "misc" {
		add("publisher", escapeLaTeX(w.Publisher))
	}
	entry.Key = freeKey(baseKey(w, year), taken)
	return entry
}

// bibType maps a registry's type to the BibTeX type an entry gets: Crossref
// names types in lower case with hyphens, DataCite in camel case. A chapter
// of a proceedings volume and one of an edited book look alike to Crossref,
// so both become incollection; an author who knows better changes the type,
// which the check never reads.
func bibType(w *Work) string {
	switch w.Kind {
	case "journal-article", "JournalArticle":
		return "article"
	case "proceedings-article", "ConferencePaper":
		return "inproceedings"
	case "book-chapter", "book-section", "book-part", "BookChapter":
		return "incollection"
	case "book", "monograph", "edited-book", "reference-book", "Book":
		return "book"
	default:
		return "misc"
	}
}

// authorField writes each author as "Family, Given". An author with no given
// name — an organisation — is braced, so that BibTeX and the check read it as
// one name rather than a family name after given ones.
func authorField(authors []Person) string {
	names := make([]string, len(authors))
	for i, a := range authors {
		if a.Given == "" {
			names[i] = "{" + escapeLaTeX(a.Family) + "}"
			continue
		}
		names[i] = escapeLaTeX(a.Family) + ", " + escapeLaTeX(a.Given)
	}
	return strings.Join(names, " and ")
}

// stopWords are the title words a key skips: they say nothing about the work.
// The Russian ones are listed as keyWord transliterates them.
var stopWords = []string{"a", "an", "the", "on", "of", "in", "for", "and", "to", "with", "from", "by", "towards", "toward", "is", "are", "how", "what", "when", "why", "do", "does",
	"o", "ob", "v", "vo", "i", "na", "k", "s", "po", "dlya", "kak", "chto", "ili"}

func baseKey(w *Work, year int) string {
	author := "anonymous"
	if len(w.Authors) > 0 {
		if a := keyWord(w.Authors[0].Family); a != "" {
			author = a
		}
	}
	word := ""
	for _, t := range strings.Fields(w.Title) {
		if k := keyWord(t); k != "" && !slices.Contains(stopWords, k) {
			word = k
			break
		}
	}
	when := "nd" // no date, as a bibliography writes it
	if year > 0 {
		when = strconv.Itoa(year)
	}
	return author + when + word
}

// keyWord folds a word into the letters a key may use: lower-case ASCII, no
// accents, no spaces or punctuation. Cyrillic letters are transliterated, so
// a Russian author gets a key of their own rather than "anonymous".
func keyWord(s string) string {
	var b strings.Builder
	for _, r := range norm.NFKD.String(strings.ToLower(s)) {
		switch latin, isCyrillic := cyrillic[r]; {
		case isCyrillic:
			b.WriteString(latin)
		case r < unicode.MaxASCII && unicode.IsLetter(r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// cyrillic transliterates the Russian alphabet as passports do. NFKD has
// already split ё and й into е and и with a combining mark, which is dropped.
var cyrillic = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ж': "zh", 'з': "z", 'и': "i",
	'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t",
	'у': "u", 'ф': "f", 'х': "kh", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch", 'ъ': "",
	'ы': "y", 'ь': "", 'э': "e", 'ю': "iu", 'я': "ia",
}

func freeKey(base string, taken func(string) bool) string {
	key := base
	for suffix := 'b'; taken(key); suffix++ {
		key = base + string(suffix)
	}
	return key
}

func earliest(years []int) int {
	if len(years) == 0 {
		return 0
	}
	return slices.Min(years)
}

func last(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[len(values)-1]
}

// tidyTitle drops what some registries add to a title and a style would
// print twice: a final period, since a style ends the title with its own, and
// the colon after a question or exclamation mark that joins a subtitle. A
// final period that belongs to an abbreviation, as in "the U.S." or "et al.",
// stays.
func tidyTitle(title string) string {
	title = strings.NewReplacer("?:", "?", "!:", "!").Replace(strings.TrimSpace(title))
	lastWord, isSentence := strings.CutSuffix(title[strings.LastIndex(title, " ")+1:], ".")
	if isSentence && !strings.Contains(lastWord, ".") && !slices.Contains(abbreviations, strings.ToLower(lastWord)) {
		title = strings.TrimSuffix(title, ".")
	}
	return title
}

// abbreviations are the words a title may end with whose period is their own.
var abbreviations = []string{"al", "etc", "inc", "ltd", "co", "no", "vs", "jr", "sr", "st"}

// protectAcronyms braces every word of a title with two capitals or more, such
// as MATHWELL or LLMs, so that a style that changes a title's case keeps them.
// A compound is braced for an acronym among its parts, as in GSM-Symbolic, and
// not for its parts' capitals, as in Teacher–Student, which a style may lower.
// The check ignores braces, so a protected title still matches its record.
func protectAcronyms(title string) string {
	words := strings.Fields(title)
	for i, word := range words {
		parts := strings.FieldsFunc(word, func(r rune) bool { return r == '-' || r == '–' || r == '—' || r == '/' })
		if slices.ContainsFunc(parts, isAcronym) {
			words[i] = "{" + word + "}"
		}
	}
	return strings.Join(words, " ")
}

func isAcronym(word string) bool {
	capitals := 0
	for _, r := range word {
		if unicode.IsUpper(r) {
			capitals++
		}
	}
	return capitals >= 2
}

var latexSpecial = strings.NewReplacer("&", `\&`, "%", `\%`, "$", `\$`, "#", `\#`, "_", `\_`, "^", `\^{}`, "~", `\~{}`,
	"{", "", "}", "", `\`, "")

// escapeLaTeX writes the characters LaTeX reads as commands the way a
// bibliography must carry them. Braces and backslashes are dropped: a registry
// title with one brace unmatched would otherwise leave a file no BibTeX reader
// can read, and the check ignores braces anyway. A title's mathematics comes
// out as text, $x^2$ as \$x\^{}2\$: it compiles, and an author may set it in
// math mode again, which the check accepts.
func escapeLaTeX(s string) string { return latexSpecial.Replace(s) }
