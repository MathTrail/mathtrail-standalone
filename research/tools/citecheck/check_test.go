package main

import (
	"slices"
	"strings"
	"testing"
)

// klinkenberg is the record Crossref holds for Klinkenberg et al. 2011.
var klinkenberg = Work{
	DOI: "10.1016/j.compedu.2011.02.003", Registry: "Crossref", Kind: "journal-article",
	Title:   "Computer adaptive practice of Maths ability using a new item response model for on the fly ability and difficulty estimation",
	Authors: []Person{{"Klinkenberg", "S."}, {"Straatemeier", "M."}, {"van der Maas", "H.L.J."}},
	Years:   []int{2011}, Venues: []string{"Computers & Education"},
}

func entry(fields ...string) Entry {
	e := Entry{Type: "article", Key: "k"}
	for i := 0; i+1 < len(fields); i += 2 {
		e.Fields = append(e.Fields, Field{Name: fields[i], Value: fields[i+1]})
	}
	return e
}

var goodKlinkenberg = []string{
	"author", "Klinkenberg, S. and Straatemeier, M. and van der Maas, H.L.J.",
	"title", "Computer adaptive practice of {Maths} ability using a new item response model for on the fly ability and difficulty estimation",
	"journal", `Computers \& Education`,
	"year", "2011",
}

// with is goodKlinkenberg with one field replaced.
func with(name, value string) Entry {
	e := entry(goodKlinkenberg...)
	for i := range e.Fields {
		if e.Fields[i].Name == name {
			e.Fields[i].Value = value
		}
	}
	return e
}

func TestCheck(t *testing.T) {
	t.Parallel()
	retracted := klinkenberg
	retracted.Notices = []string{"retraction (10.1/notice)"}
	unrecordedVenue := klinkenberg
	unrecordedVenue.Venues = nil
	organisation := klinkenberg
	organisation.Authors = []Person{{Family: "National Research Council, Mathematics Learning Study Committee"}}
	undated := klinkenberg
	undated.Years = nil
	tests := []struct {
		name  string
		entry Entry
		work  Work
		want  []string // a fragment of each problem expected, in order
	}{
		{"an entry as the registry records it", entry(goodKlinkenberg...), klinkenberg, nil},
		{"a changed title", with("title", "Computer adaptive practice of Maths ability"), klinkenberg, []string{"title"}},
		{"an extra author", with("author", "Klinkenberg, S. and Straatemeier, M. and van der Maas, H.L.J. and Maris, G."), klinkenberg, []string{"authors"}},
		{"a missing author", with("author", "Klinkenberg, S. and van der Maas, H.L.J."), klinkenberg, []string{"authors"}},
		{"authors in another order", with("author", "Straatemeier, M. and Klinkenberg, S. and van der Maas, H.L.J."), klinkenberg, []string{"authors"}},
		{"and others after the first authors", with("author", "Klinkenberg, S. and others"), klinkenberg, nil},
		{"the first authors only, without and others", with("author", "Klinkenberg, S. and Straatemeier, M."), klinkenberg, []string{"authors"}},
		{"and others after a wrong first author", with("author", "Maris, G. and others"), klinkenberg, []string{"authors"}},
		{"given names first", with("author", "S. Klinkenberg and M. Straatemeier and {van der Maas}, H. L. J."), klinkenberg, nil},
		{"a wrong year", with("year", "2012"), klinkenberg, []string{"year"}},
		{"a year that is not a number", with("year", "in press"), klinkenberg, []string{"year"}},
		{"a wrong journal", with("journal", "Computers in Human Behavior"), klinkenberg, []string{"venue"}},
		{"the journal spelled with and", with("journal", "Computers and Education"), klinkenberg, nil},
		{"the journal abbreviated", with("journal", "Comput. Educ."), klinkenberg, nil},
		{"a word of the journal only", with("journal", "Education"), klinkenberg, []string{"venue"}},
		{"another journal with as many words", with("journal", "Computers & Society"), klinkenberg, []string{"venue"}},
		{"a journal that holds the recorded one", with("journal", "Computers & Education Research"), klinkenberg, []string{"venue"}},
		{"an empty journal", with("journal", "{{}}"), klinkenberg, []string{"venue"}},
		{"authors with LaTeX accents", with("author", `Klinkenberg, S. and Str{\"a}atemeier, M. and van der Maas, H.L.J.`), klinkenberg, nil},
		{"a retracted work", entry(goodKlinkenberg...), retracted, []string{"retraction (10.1/notice)"}},
		{"several problems at once", with("year", "1999"), Work{Registry: "Crossref", Title: "Other", Years: []int{2011}}, []string{"title", "year", "venue"}},
		{"a venue the registry does not record", with("journal", "Findings of EMNLP"), unrecordedVenue, []string{"cannot be checked"}},
		{"others alone", with("author", "others"), klinkenberg, []string{"authors"}},
		{"an organisation with a comma", entry("author", "{National Research Council, Mathematics Learning Study Committee}", "title", klinkenberg.Title, "year", "2011"), organisation, nil},
		{"a year the registry does not record", entry(goodKlinkenberg...), undated, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := Check(tt.entry, &tt.work)

			if len(got) != len(tt.want) {
				t.Fatalf("problems = %q, want %d of them: %q", got, len(tt.want), tt.want)
			}
			for i, fragment := range tt.want {
				if !strings.Contains(got[i], fragment) {
					t.Errorf("problem %d = %q, want it to mention %q", i, got[i], fragment)
				}
			}
		})
	}
}

func TestCheckAcceptsAccentsSubtitlesAndSeries(t *testing.T) {
	t.Parallel()
	chapter := Work{
		Registry: "Crossref", Title: "Beyond Final Answers", Subtitle: "Evaluating Large Language Models for Math Tutoring",
		Authors: []Person{{Family: "Calò"}}, Years: []int{2025},
		Venues: []string{"Lecture Notes in Computer Science", "Artificial Intelligence in Education"},
	}
	for _, tt := range []struct {
		name  string
		entry Entry
		work  Work
	}{
		{"accents and a subtitle", entry("author", "Calo, Tommaso", "title", "Beyond Final Answers: Evaluating Large Language Models for Math Tutoring", "booktitle", "Artificial Intelligence in Education", "year", "2025"), chapter},
		{"the series as the venue", entry("author", "Calò, T.", "title", "Beyond final answers -- evaluating large language models for math tutoring", "booktitle", "Lecture Notes in Computer Science", "year", "2025"), chapter},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Check(tt.entry, &tt.work); len(got) > 0 {
				t.Errorf("problems = %q, want none", got)
			}
		})
	}
}

func TestDOIOf(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		entry Entry
		want  string
	}{
		{"its own DOI", entry("doi", " 10.1016/j.compedu.2011.02.003 ", "eprint", "2503.16460", "archiveprefix", "arXiv"), "10.1016/j.compedu.2011.02.003"},
		{"an arXiv preprint", entry("eprint", "2503.16460v2", "archiveprefix", "arXiv"), "10.48550/arXiv.2503.16460"},
		{"biblatex's eprinttype", entry("eprint", "2402.15861", "eprinttype", "arxiv"), "10.48550/arXiv.2402.15861"},
		{"both fields for the archive", entry("eprint", "2402.15861", "archiveprefix", "arXiv", "eprinttype", "arxiv"), "10.48550/arXiv.2402.15861"},
		{"a DOI written as its address", entry("doi", "https://doi.org/10.1016/j.compedu.2011.02.003"), "10.1016/j.compedu.2011.02.003"},
		{"an eprint written with its archive", entry("eprint", "arXiv:2402.15861v3", "archiveprefix", "arXiv"), "10.48550/arXiv.2402.15861"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got, err := DOIOf(tt.entry); err != nil || got != tt.want {
				t.Errorf("DOIOf = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
	if _, err := DOIOf(entry("eprint", "2402.15861", "url", "https://example.com")); err == nil {
		t.Error("an eprint of no known archive: DOIOf succeeded, want an error")
	}
}

func TestEntryAuthors(t *testing.T) {
	t.Parallel()

	families, truncated := EntryAuthors("Gupta, Adit and  Reddig, J.\n and {Lab and Things} and Daniel Weitekamp AND Anders Andersen and others")

	if want := []string{"Gupta", "Reddig", "Lab and Things", "Weitekamp", "Andersen"}; !slices.Equal(families, want) || !truncated {
		t.Errorf("EntryAuthors = %q, %v; want %q, true", families, truncated, want)
	}
}

func TestNormalize(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		`Computers \& Education`:        "computers and education",
		"Computers &amp; Education":     "computers and education",
		"  Calò,  Tommaso ":             "calo tommaso",
		"Pelánek":                       "pelanek",
		`Pel{\'a}nek`:                   "pelanek",
		`M{\"u}ller and \c{C}elik`:      "muller and celik",
		"{Lab and Things}":              "lab and things",
		"{GSM-Symbolic}: <i>Limits</i>": "gsm symbolic limits",
	} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			if got := normalize(in); got != want {
				t.Errorf("normalize(%q) = %q, want %q", in, got, want)
			}
		})
	}
}
