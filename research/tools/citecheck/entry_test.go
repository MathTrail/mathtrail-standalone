package main

import (
	"reflect"
	"testing"
)

func TestEntryForWritesEachKindOfWork(t *testing.T) {
	t.Parallel()
	none := func(string) bool { return false }
	chapter := Work{
		DOI: "10.1007/978-3-031-98414-3_23", Registry: "Crossref", Kind: "book-chapter",
		Title: "Beyond Final Answers", Subtitle: "Evaluating Large Language Models for Math Tutoring",
		Authors: []Person{{"Gupta", "Adit"}, {"Calò", "Tommaso"}}, Years: []int{2025},
		Venues:    []string{"Lecture Notes in Computer Science", "Artificial Intelligence in Education"},
		Publisher: "Springer Nature Switzerland", Pages: "323-337",
	}
	preprint := Work{
		DOI: "10.48550/arxiv.2503.16460", Registry: "DataCite", Kind: "Preprint",
		Title:   "Beyond Final Answers: Evaluating Large Language Models for Math Tutoring",
		Authors: []Person{{"Gupta", "Adit"}}, Years: []int{2025}, Publisher: "arXiv",
	}
	deposit := Work{
		DOI: "10.5281/zenodo.1", Registry: "DataCite", Kind: "JournalArticle", Title: "A deposit of an article",
		Authors: []Person{{"Doe", "J."}}, Years: []int{2020}, Venues: []string{"Journal of Things"}, Publisher: "Zenodo",
	}
	paper := Work{
		DOI: "10.1145/3000001", Registry: "Crossref", Kind: "proceedings-article", Title: "Adaptive Practice at Scale",
		Authors: []Person{{"Doe", "Jane"}}, Years: []int{2019}, Venues: []string{"Proceedings of the Learning Conference"}, Publisher: "ACM",
	}
	monograph := Work{
		DOI: "10.1007/978-0-000-00000-0", Registry: "Crossref", Kind: "monograph", Title: "Mathematical Olympiads for Young Children",
		Authors: []Person{{"Smith", "Ann"}}, Years: []int{2015}, Publisher: "Springer",
	}
	undated := Work{
		DOI: "10.1000/report.1", Registry: "Crossref", Kind: "report", Title: "Guidance on Tutoring",
		Authors: []Person{{Family: "An Agency"}}, Publisher: "The Agency",
	}
	uncontained := Work{
		DOI: "10.1007/978-0-000-00000-0_3", Registry: "Crossref", Kind: "book-chapter", Title: "Counting Problems",
		Authors: []Person{{"Lee", "Kim"}}, Years: []int{2018}, Publisher: "Springer",
	}
	tests := []struct {
		name string
		work Work
		want Entry
	}{
		{"a journal article", klinkenberg, Entry{Type: "article", Key: "klinkenberg2011computer", Fields: []Field{
			{"author", "Klinkenberg, S. and Straatemeier, M. and van der Maas, H.L.J."},
			{"title", "Computer adaptive practice of Maths ability using a new item response model for on the fly ability and difficulty estimation"},
			{"journal", `Computers \& Education`},
			{"year", "2011"},
			{"doi", "10.1016/j.compedu.2011.02.003"},
		}}},
		{"a chapter of proceedings in a series", chapter, Entry{Type: "incollection", Key: "gupta2025beyond", Fields: []Field{
			{"author", "Gupta, Adit and Calò, Tommaso"},
			{"title", "Beyond Final Answers: Evaluating Large Language Models for Math Tutoring"},
			{"booktitle", "Artificial Intelligence in Education"},
			{"series", "Lecture Notes in Computer Science"},
			{"publisher", "Springer Nature Switzerland"},
			{"year", "2025"},
			{"pages", "323--337"},
			{"doi", "10.1007/978-3-031-98414-3_23"},
		}}},
		{"an arXiv preprint", preprint, Entry{Type: "misc", Key: "gupta2025beyond", Fields: []Field{
			{"author", "Gupta, Adit"},
			{"title", "Beyond Final Answers: Evaluating Large Language Models for Math Tutoring"},
			{"year", "2025"},
			{"doi", "10.48550/arxiv.2503.16460"},
			{"eprint", "2503.16460"},
			{"archiveprefix", "arXiv"},
		}}},
		{"a journal article DataCite holds", deposit, Entry{Type: "article", Key: "doe2020deposit", Fields: []Field{
			{"author", "Doe, J."},
			{"title", "A deposit of an article"},
			{"journal", "Journal of Things"},
			{"year", "2020"},
			{"doi", "10.5281/zenodo.1"},
		}}},
		{"a paper of proceedings", paper, Entry{Type: "inproceedings", Key: "doe2019adaptive", Fields: []Field{
			{"author", "Doe, Jane"},
			{"title", "Adaptive Practice at Scale"},
			{"booktitle", "Proceedings of the Learning Conference"},
			{"publisher", "ACM"},
			{"year", "2019"},
			{"doi", "10.1145/3000001"},
		}}},
		{"a monograph", monograph, Entry{Type: "book", Key: "smith2015mathematical", Fields: []Field{
			{"author", "Smith, Ann"},
			{"title", "Mathematical Olympiads for Young Children"},
			{"publisher", "Springer"},
			{"year", "2015"},
			{"doi", "10.1007/978-0-000-00000-0"},
		}}},
		{"a report with no year", undated, Entry{Type: "misc", Key: "anagencyndguidance", Fields: []Field{
			{"author", "{An Agency}"},
			{"title", "Guidance on Tutoring"},
			{"doi", "10.1000/report.1"},
			{"publisher", "The Agency"},
		}}},
		{"a chapter with no container", uncontained, Entry{Type: "incollection", Key: "lee2018counting", Fields: []Field{
			{"author", "Lee, Kim"},
			{"title", "Counting Problems"},
			{"publisher", "Springer"},
			{"year", "2018"},
			{"doi", "10.1007/978-0-000-00000-0_3"},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := EntryFor(&tt.work, none)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EntryFor =\n%s\nwant\n%s", FormatEntry(got), FormatEntry(tt.want))
			}
			if problems := Check(got, &tt.work); len(problems) > 0 {
				t.Errorf("the written entry does not check out: %q", problems)
			}
		})
	}
}

func TestEntryForFindsAFreeKey(t *testing.T) {
	t.Parallel()
	taken := map[string]bool{"klinkenberg2011computer": true, "klinkenberg2011computerb": true}

	got := EntryFor(&klinkenberg, func(key string) bool { return taken[key] })

	if got.Key != "klinkenberg2011computerc" {
		t.Errorf("key = %s, want klinkenberg2011computerc", got.Key)
	}
}

func TestBaseKeySkipsWordsThatSayNothing(t *testing.T) {
	t.Parallel()
	work := Work{Title: "The Eighty Five Percent Rule for optimal learning", Authors: []Person{{Family: "Wilson"}}}

	if got := baseKey(&work, 2019); got != "wilson2019eighty" {
		t.Errorf("baseKey = %s, want wilson2019eighty", got)
	}
	if got := baseKey(&Work{Title: "Über das Rechnen"}, 2020); got != "anonymous2020uber" {
		t.Errorf("baseKey without authors = %s, want anonymous2020uber", got)
	}
	if got := baseKey(&work, 0); got != "wilsonndeighty" {
		t.Errorf("baseKey without a year = %s, want wilsonndeighty", got)
	}
	russian := Work{Title: "Об олимпиадных задачах", Authors: []Person{{Family: "Щербаков"}}}
	if got := baseKey(&russian, 2020); got != "shcherbakov2020olimpiadnykh" {
		t.Errorf("baseKey of a Russian work = %s, want shcherbakov2020olimpiadnykh", got)
	}
	if got := keyWord("Ёлкин-Йорк"); got != "elkiniork" {
		t.Errorf("keyWord(Ёлкин-Йорк) = %s, want elkiniork", got)
	}
}

func TestEntryForWritesWhatItsOwnCheckAccepts(t *testing.T) {
	t.Parallel()
	works := map[string]Work{
		"an organisation as author": {Registry: "Crossref", Kind: "report", Title: "Guidance on generative AI in education",
			Authors: []Person{{Family: "Organisation for Economic Co-operation and Development"}, {Family: "Holmes", Given: "Wayne"}}, Years: []int{2023}},
		"a title with an unmatched brace": {Registry: "Crossref", Kind: "journal-article", Title: `The set {0,1 and the \ of it`,
			Authors: []Person{{Family: "Author", Given: "A."}}, Years: []int{2020}, Venues: []string{"Journal"}},
		"a title with mathematics": {Registry: "DataCite", Kind: "Preprint", Title: `The $O(n^2)$ bound for $\alpha$-trees, roughly ~ tight`,
			Authors: []Person{{Family: "Author", Given: "A."}}, Years: []int{2024}},
		"a title the registry ends with punctuation": {Registry: "Crossref", Kind: "journal-article", Title: "When and where do we apply what we learn?: A taxonomy for far transfer.",
			Authors: []Person{{Family: "Barnett", Given: "Susan M."}}, Years: []int{2002}, Venues: []string{"Psychological Bulletin"}},
	}
	for name, work := range works {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			written := FormatEntry(EntryFor(&work, func(string) bool { return false }))

			entries, err := ParseBib(written)
			if err != nil || len(entries) != 1 {
				t.Fatalf("the written entry cannot be read back: %v\n%s", err, written)
			}
			if problems := Check(entries[0], &work); len(problems) > 0 {
				t.Errorf("problems = %q, want none\n%s", problems, written)
			}
		})
	}
}

func TestProtectAcronymsBracesWordsWithCapitals(t *testing.T) {
	t.Parallel()

	got := protectAcronyms("MATHWELL: Generating Math Word Problems with LLMs in Teacher–Student Large-Scale GSM-Symbolic LLM-based Tasks")

	if want := "{MATHWELL:} Generating Math Word Problems with {LLMs} in Teacher–Student Large-Scale {GSM-Symbolic} {LLM-based} Tasks"; got != want {
		t.Errorf("protectAcronyms = %q, want %q", got, want)
	}
}

func TestTidyTitleDropsWhatAStyleWouldPrintTwice(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"Desirable difficulties in theory and practice.":                          "Desirable difficulties in theory and practice",
		"When and where do we apply what we learn?: A taxonomy for far transfer.": "When and where do we apply what we learn? A taxonomy for far transfer",
		"Stop!: A title":                                    "Stop! A title",
		"Schooling in the U.S.":                             "Schooling in the U.S.",
		"Numbers, shapes, etc.":                             "Numbers, shapes, etc.",
		"A reply to Smith et al.":                           "A reply to Smith et al.",
		"Does Far Transfer Exist?":                          "Does Far Transfer Exist?",
		"The Eighty Five Percent Rule for optimal learning": "The Eighty Five Percent Rule for optimal learning",
		"Rules of thumb: a note":                            "Rules of thumb: a note",
	} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			if got := tidyTitle(in); got != want {
				t.Errorf("tidyTitle(%q) = %q, want %q", in, got, want)
			}
		})
	}
}

func TestEntryForTidiesTheTitle(t *testing.T) {
	t.Parallel()
	work := Work{Registry: "Crossref", Kind: "journal-article", Title: "Desirable difficulties in theory and practice.",
		Authors: []Person{{Family: "Bjork", Given: "Robert A."}}, Years: []int{2020}}

	got := EntryFor(&work, func(string) bool { return false }).Get("title")

	if want := "Desirable difficulties in theory and practice"; got != want {
		t.Errorf("title = %q, want %q", got, want)
	}
}

func TestEscapeLaTeXKeepsATitleCompilable(t *testing.T) {
	t.Parallel()

	got := escapeLaTeX(`50% of $x^2$ ~ {y} & #1_a \b`)

	if want := `50\% of \$x\^{}2\$ \~{} y \& \#1\_a b`; got != want {
		t.Errorf("escapeLaTeX = %s, want %s", got, want)
	}
}
