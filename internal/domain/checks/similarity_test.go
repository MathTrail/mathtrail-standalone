package checks_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// goldenPairs are the prototype's own measurements: pairs of questions and
// what Postgres's similarity() said of them, exported by the prototype's code
// (T16). They are the reference the measure is held to, so no test rewrites
// them.
type goldenPairs struct {
	Pairs []struct {
		Name       string  `json:"name"`
		A          string  `json:"a"`
		B          string  `json:"b"`
		Similarity float64 `json:"similarity"`
	} `json:"pairs"`
}

// pgFloat is how close to Postgres's number ours has to come. Postgres works
// in single precision, so its answer is the nearest float32 to ours.
const pgFloat = 1e-6

func TestSimilarityIsPostgresTrigramSimilarity(t *testing.T) {
	t.Parallel()

	golden := readGoldenPairs(t)

	for _, pair := range golden.Pairs {
		t.Run(pair.Name, func(t *testing.T) {
			t.Parallel()

			got := checks.Similarity(pair.A, pair.B)
			if math.Abs(got-pair.Similarity) > pgFloat {
				t.Errorf("Similarity(%q, %q) = %.7f, want %.7f", pair.A, pair.B, got, pair.Similarity)
			}
		})
	}
}

// The construction is pg_trgm's: every run of letters and digits is a word,
// whatever separates it, and each word is padded before its trigrams are taken.
func TestWordsAreCutTheWayPostgresCutsThem(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		a, b string
		want float64
	}{
		{"an apostrophe separates words", "o'clock", "o clock", 1},
		{"a hyphen separates words", "a 5-litre jug", "a 5 litre jug", 1},
		{"a multiplication sign is not a word", "25 × 4", "25 4", 1},
		{"punctuation is not a word", "Ann, Ben; Kim!", "ann ben kim", 1},
		{"a one-letter word still counts", "a", "a", 1},
		{"the same letters in another order are another word", "ab", "ba", 0},
		{"nothing to compare is nothing in common", "", "", 0},
		{"punctuation alone is nothing to compare", "?!", "?!", 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := checks.Similarity(test.a, test.b); got != test.want {
				t.Errorf("Similarity(%q, %q) = %v, want %v", test.a, test.b, got, test.want)
			}
		})
	}
}

// A text written without spaces is cut into character bigrams instead: the
// same construction, one character padded on each side. The first pair is
// worked by hand — ten bigrams in common out of fifteen between them.
func TestTextWithoutSpacesIsCutIntoBigrams(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		a, b string
		want float64
	}{
		// Five bigrams in common out of nine; in trigrams it would be four out
		// of ten, which is how the two cuts are told apart.
		{"one character changed in the middle", "我有三个苹果。", "我有五个苹果。", 5.0 / 9},
		{"one character changed and one added", "小明有三个苹果和两个梨。", "小明有三个苹果和两个大桃。", 10.0 / 15},
		{"the same sentence", "小明有三个苹果。", "小明有三个苹果。", 1},
		{"punctuation is a boundary", "小明，小红。", "小明 小红", 1},
		{"Chinese against English shares nothing", "小明有三个苹果。", "Ming has three apples.", 0},
		{"Japanese in three scripts is one text", "りんごを三つ買いました。", "りんごを三つ買いました。", 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := checks.Similarity(test.a, test.b); math.Abs(got-test.want) > 1e-12 {
				t.Errorf("Similarity(%q, %q) = %v, want %v", test.a, test.b, got, test.want)
			}
		})
	}
}

// Which unit a text is measured in depends on the script most of its letters
// are written in, counted as two kinds: with spaces between words, or without.
func TestAScriptWithoutSpacesIsTheOneMostLettersAreIn(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		text string
		want bool
	}{
		{"English", "Four spaceships dock in pairs.", false},
		{"Russian", "Четыре корабля стыкуются парами.", false},
		{"Chinese", "小明有三个苹果。", true},
		{"Chinese with two Latin labels", "A 和 B 一共有几个苹果？", true},
		{"Japanese in kanji, hiragana and katakana", "リンゴを三つ買いました。", true},
		{"Thai", "มีแอปเปิลสามลูก", true},
		{"English naming one character", "The sign 和 means and.", false},
		{"digits are no script", "12 + 34 = 46", false},
		{"nothing at all", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := checks.Spaceless(test.text); got != test.want {
				t.Errorf("Spaceless(%q) = %v, want %v", test.text, got, test.want)
			}
		})
	}
}

// The question comes from the chat's model, and so does every past one the
// measure is run over: whatever the text, the measure ends, stays between
// nothing and everything, and says the same whichever side a text is on.
func FuzzSimilarity(f *testing.F) {
	f.Add("Four spaceships dock in pairs.", "FOUR   SPACESHIPS   DOCK   IN   PAIRS.")
	f.Add("小明有三个苹果。", "Ming has three apples.")
	f.Add("", "\x00\xff")
	f.Add("a", "a")

	f.Fuzz(func(t *testing.T, a, b string) {
		forward, backward := checks.Similarity(a, b), checks.Similarity(b, a)
		if forward < 0 || forward > 1 {
			t.Fatalf("Similarity(%q, %q) = %v, want a share between 0 and 1", a, b, forward)
		}
		if forward != backward {
			t.Fatalf("Similarity(%q, %q) = %v, and the other way round %v", a, b, forward, backward)
		}
		if self := checks.Similarity(a, a); self != 0 && self != 1 {
			t.Fatalf("Similarity(%q, itself) = %v, want 1, or 0 for a text with no word in it", a, self)
		}
	})
}

// readGoldenPairs reads the prototype's measured pairs.
func readGoldenPairs(t *testing.T) goldenPairs {
	t.Helper()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "golden", "trgm_similarity.json"))
	if err != nil {
		t.Fatalf("read the golden vectors: %v", err)
	}
	var golden goldenPairs
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("parse the golden vectors: %v", err)
	}
	if len(golden.Pairs) == 0 {
		t.Fatal("the golden vectors hold no pairs")
	}
	return golden
}
