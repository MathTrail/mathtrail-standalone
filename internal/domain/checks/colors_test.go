package checks_test

import (
	"encoding/json"
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// painted is a picture of flags painted with colours, each called by its word:
// a flag to a colour, its other stripe the unknown.
func painted(words map[string]string) json.RawMessage {
	flags := make([][]string, 0, len(words))
	for _, paint := range slices.Sorted(maps.Keys(words)) {
		flags = append(flags, []string{paint, "?"})
	}
	raw, err := json.Marshal(map[string]any{
		"kind": "flags", "colors": words, "groups": []any{map[string]any{"flags": flags}},
	})
	if err != nil {
		panic(err)
	}
	return raw
}

// A colour is named where the wording writes its word in any form that keeps
// the word's stem: another case or number, capitals, ё written as е, and a
// word of two whose both parts are there.
func TestAColourIsNamedInAnyFormThatKeepsItsStem(t *testing.T) {
	t.Parallel()

	for _, test := range []struct{ name, question, word string }{
		{"as the wording writes it", "How many flags have a red stripe on top?", "red"},
		{"in another case", "Сколько флажков с полосой красной краски?", "красная"},
		{"in another number", "Сколько синих флажков?", "синяя"},
		{"with ё written as е", "Нижняя полоса желтая. Сколько флажков?", "жёлтая"},
		{"with е written as ё", "Нижняя полоса жёлтая. Сколько флажков?", "желтая"},
		{"in capitals", "RED stripes come first. How many flags?", "red"},
		{"declined in German", "Wie viele Flaggen haben einen roten Streifen?", "rot"},
		{"a word of two", "Có bao nhiêu lá cờ có sọc xanh lá?", "xanh lá"},
		{"in a script written without spaces", "白い旗はいくつありますか。", "白"},
		{"in a script that marks its vowels over its letters", "कितने झंडों में पीली पट्टी है?", "पीला"},
		{"after a hyphen", "Флажок тёмно-красного цвета.", "красная"},
		{"joined to the article, in Arabic", "كم علماً لونه الأحمر؟", "أحمر"},
		{"inside a word of Thai, which writes no spaces", "มีธงสีแดงกี่ผืน", "สีแดง"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if problems := matched(test.question, painted(map[string]string{"red": test.word})); len(problems) != 0 {
				t.Errorf("PictureMatch() = %v, want the colour named", problems)
			}
		})
	}
}

// A colour the wording does not name, or names only in a form whose stem is
// another, is counted, never quoted: the word may be the right option's.
func TestAColourTheWordingDoesNotNameIsCountedNeverQuoted(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name, question string
		words          map[string]string
		count          string
	}{
		{"two words the wording never says", "How many flags can be made?",
			map[string]string{"red": "scarlet", "blue": "azure"}, "does not name 2 of the colours"},
		{"one of two words of a colour", "Có bao nhiêu cờ màu xanh?", map[string]string{"green": "xanh lá"},
			"does not name 1 of the colours"},
		{"a stem that changes as the word is declined", "كم علماً لونه حمراء؟", map[string]string{"red": "أحمر"},
			"does not name 1 of the colours"},
		{"a word longer than the wording's", "赤の旗はいくつありますか。", map[string]string{"red": "赤い"},
			"does not name 1 of the colours"},
		{"another word for the colour", "Wie viele grüne Flaggen gibt es?", map[string]string{"green": "green"},
			"does not name 1 of the colours"},
		{"a word inside another word", "How many coloured flags can be made, and which ones?",
			map[string]string{"red": "red", "white": "white"}, "does not name 1 of the colours"},
		{"a word ending another word", "Сколько прекрасных флажков можно сделать?", map[string]string{"red": "красная"},
			"does not name 1 of the colours"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problems := matched(test.question, painted(test.words))
			if len(problems) != 1 || !strings.Contains(problems[0].Message, test.count) {
				t.Fatalf("PictureMatch() = %v, want one problem that %s", problems, test.count)
			}
			for word := range maps.Values(test.words) {
				if strings.Contains(problems[0].Message, word) {
					t.Errorf("message = %q quotes %q, want the colours counted alone", problems[0].Message, word)
				}
			}
		})
	}
}

// The colours of the picture of the solution may be named in the solution as
// well as in the question, and one named in neither is refused.
func TestTheColoursOfThePictureOfTheSolutionAreNamedInTheWordingOrTheSolution(t *testing.T) {
	t.Parallel()

	named := accepted()
	named.draft.Task.Solution += " A red flag and a blue one make a pair."
	named.draft.Task.SolutionPicture = painted(map[string]string{"red": "red", "blue": "blue"})
	if outcome := named.review(t); !outcome.Accepted() {
		t.Errorf("colours the solution names are refused: %v", outcome.Problems)
	}

	unnamed := accepted()
	unnamed.draft.Task.SolutionPicture = painted(map[string]string{"red": "crimson"})
	outcome := unnamed.review(t)
	if codes := codesOf(&outcome); !slices.Equal(codes, []checks.Code{checks.CodeDrawingMismatch}) ||
		!mentions(outcome.Problems, "neither task.question nor task.solution names 1 of the colours") {
		t.Errorf("a colour named in neither = %v, want it refused", outcome.Problems)
	}
}
