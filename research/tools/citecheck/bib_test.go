package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseBibReadsEntries(t *testing.T) {
	t.Parallel()
	text := `% A comment before the entries.
@Article{klinkenberg2011computer,
  author = {Klinkenberg, S. and {van der Maas}, H.L.J.},
  title  = {Computer adaptive practice of {Maths} ability},
  year   = 2011,
}

% Another comment.
@misc{gupta2025beyond, eprint = {2503.16460}, archiveprefix = {arXiv}}
`
	entries, err := ParseBib(text)

	if err != nil {
		t.Fatalf("ParseBib: %v", err)
	}
	want := []Entry{
		{Type: "article", Key: "klinkenberg2011computer", Fields: []Field{
			{"author", "Klinkenberg, S. and {van der Maas}, H.L.J."},
			{"title", "Computer adaptive practice of {Maths} ability"},
			{"year", "2011"},
		}},
		{Type: "misc", Key: "gupta2025beyond", Fields: []Field{{"eprint", "2503.16460"}, {"archiveprefix", "arXiv"}}},
	}
	if !reflect.DeepEqual(entries, want) {
		t.Errorf("ParseBib = %+v\nwant %+v", entries, want)
	}
}

func TestParseBibRefusesWhatWouldMakeACitationAmbiguousOrBroken(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, text, want string
	}{
		{"a key used twice", "@misc{a, title = {x}}\n@misc{A, title = {y}}", "line 2: key A is already used at line 1"},
		{"a field given twice", "@misc{a, title = {x}, Title = {y}}", "twice"},
		{"an entry never closed", "@misc{a, title = {x}", "the entry a never closes"},
		{"a value never closed", "@misc{a, title = {x", "the value of title never closes"},
		{"a key with no comma", "@misc{a title = {x}}", "key"},
		{"a field with no =", "@misc{a, title {x}}", "not followed by ="},
		{"text outside an entry", "title = {x}", "starts with @"},
		{"two fields with no comma", "@misc{a, title = {x} year = {1}}", "followed by , or }"},
		{"a string macro", "@string{ce = {Computers}}", "@string is not used here"},
		{"a quote never closed", "@misc{a, title = \"x", "never closes its quote"},
		{"a quoted value with a stray brace", "@misc{a, title = \"}{\"}", "closes a brace it never opened"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ParseBib(tt.text)

			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("ParseBib error = %v, want one mentioning %q", err, tt.want)
			}
		})
	}
}

func TestFormatEntryAlignsFieldsAndReadsBack(t *testing.T) {
	t.Parallel()
	e := Entry{Type: "article", Key: "k", Fields: []Field{{"author", "A, B"}, {"doi", "10.1/x"}}}

	text := FormatEntry(e)

	if want := "@article{k,\n  author = {A, B},\n  doi    = {10.1/x},\n}\n"; text != want {
		t.Errorf("FormatEntry = %q, want %q", text, want)
	}
	back, err := ParseBib(text)
	if err != nil || !reflect.DeepEqual(back, []Entry{e}) {
		t.Errorf("reading it back = %+v, %v; want the entry", back, err)
	}
}

// FuzzParseBib checks that no text makes the parser panic, and that whatever
// it reads is written back into text it reads the same way.
func FuzzParseBib(f *testing.F) {
	f.Add("@misc{a, title = {x {y} z}, year = 2011}\n% c\n@article{b,\n  author = {A and B},\n}\n")
	f.Add("@misc{a,}")
	f.Add("@@{,}")
	f.Fuzz(func(t *testing.T, text string) {
		entries, err := ParseBib(text)
		if err != nil {
			return
		}
		var written strings.Builder
		for _, e := range entries {
			written.WriteString(FormatEntry(e))
		}
		again, err := ParseBib(written.String())
		if err != nil {
			t.Fatalf("ParseBib of written entries: %v\n%s", err, written.String())
		}
		if len(again) != len(entries) {
			t.Fatalf("read back %d entries, want %d", len(again), len(entries))
		}
		for i := range entries {
			if !reflect.DeepEqual(again[i], entries[i]) {
				t.Fatalf("entry %d read back as %+v, want %+v", i, again[i], entries[i])
			}
		}
	})
}

func TestParseBibReadsQuotedValuesAndSkipsComments(t *testing.T) {
	t.Parallel()
	text := "@comment{Nothing {here} is an entry.}\n@article{k, journal = \"Comput. {Educ.}\", year = 2011}\n"

	entries, err := ParseBib(text)

	want := []Entry{{Type: "article", Key: "k", Fields: []Field{{"journal", "Comput. {Educ.}"}, {"year", "2011"}}}}
	if err != nil || !reflect.DeepEqual(entries, want) {
		t.Errorf("ParseBib = %+v, %v; want %+v", entries, err, want)
	}
}
