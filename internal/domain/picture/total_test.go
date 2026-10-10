package picture_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// The total under a picture of a solution is the equality the solution comes
// to: labels and numbers joined by the signs of arithmetic, ending in = and
// the answer, short enough to be written large under the picture.
func TestATotalIsTheEqualityTheSolutionComesTo(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		text     string
		decimals picture.Decimals
		labels   []string
		result   string
	}{
		{"sums", "2 + 2 + 2 = 6", picture.Point, nil, "6"},
		{"a difference", "9 − 3 = 6", picture.Point, nil, "6"},
		{"a division and a sum", "12 ÷ 3 + 1 = 5", picture.Point, nil, "5"},
		{"labels", "A + B = 12", picture.Point, []string{"A", "B"}, "12"},
		{"ending in a label", "12 − 7 = B", picture.Point, []string{"B"}, "B"},
		{"brackets", "(3 + 2) × 2 = 10", picture.Point, nil, "10"},
		{"a decimal comma", "3,5 + 1 = 4,5", picture.Comma, nil, "4,5"},
		{"a decimal point", "3.5 + 1 = 4.5", picture.Point, nil, "4.5"},
		{"a negative answer", "2 − 5 = −3", picture.Point, nil, "−3"},
		{"spaces around it", "  4 + 1 = 5 ", picture.Point, nil, "5"},
		{"as long as it may be", "1 + 2 + 3 + 4 + 5 + 6 + 7 = 28", picture.Point, nil, "28"},
		{"ending in a number longer than a label", "100 × 1000 = 100000", picture.Point, nil, "100000"},
		{"ending in a long decimal", "99 ÷ 8 = 12,375", picture.Comma, nil, "12,375"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			total, problems := picture.ReadTotal(tc.text, tc.decimals)
			if len(problems) != 0 {
				t.Fatalf("ReadTotal(%q) problems = %v, want none", tc.text, problems)
			}
			if !slices.Equal(total.Labels, tc.labels) || total.Result != tc.result {
				t.Errorf("ReadTotal(%q) = %+v, want the labels %v and the result %q", tc.text, total, tc.labels, tc.result)
			}
		})
	}
}

// A total that is not such an equality is refused by its path, and the
// refusal never repeats what it held.
func TestATotalThatIsNoEqualityIsRefused(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ name, text string }{
		{"empty", ""},
		{"a number alone", "6"},
		{"no equals sign", "2 + 2 + 2"},
		{"the answer left open", "2 + 2 + 2 = ?"},
		{"nothing after the equals sign", "2 + 2 + 2 ="},
		{"two numbers after it", "2 + 4 = 6 7"},
		{"an inequality", "6 > 5"},
		{"a word", "2 + 2 + 2 = 6 flags"},
		{"a word of the lesson", "2 + 2 + 2 = 6 флажков"},
		{"small letters", "x = 6"},
		{"a label too long", "ABCDEF = 6"},
		{"longer than it may be", "1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 = 36"},
		{"the other decimal mark", "3,5 + 1 = 4,5"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			total, problems := picture.ReadTotal(tc.text, picture.Point)
			if len(problems) != 1 || problems[0].Path != "solution_total" || total.Result != "" {
				t.Fatalf("ReadTotal(%q) = %+v, %v; want one refusal of solution_total", tc.text, total, problems)
			}
			if tc.text != "" && strings.Contains(problems[0].Rule, tc.text) {
				t.Errorf("the refusal %q repeats what the total held", problems[0].Rule)
			}
		})
	}
}

// Whatever sum of numbers a solution comes to, written as an equality ending
// in its value, the total is read, and ends in that value.
func TestATotalHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)
	properties.Property("an equality of numbers is read, and ends in its value", prop.ForAll(
		func(terms []int) bool {
			sum, written := 0, make([]string, 0, len(terms))
			for _, term := range terms {
				sum += term
				written = append(written, strconv.Itoa(term))
			}
			text := strings.Join(written, " + ") + " = " + strconv.Itoa(sum)
			if len([]rune(text)) > picture.MaxTotalCharacters {
				return true
			}
			total, problems := picture.ReadTotal(text, picture.Point)
			return len(problems) == 0 && total.Result == strconv.Itoa(sum)
		},
		gen.SliceOfN(3, gen.IntRange(0, 999)),
	))
	properties.TestingRun(t)
}

// A total comes from the chat's model and is untrusted: whatever it holds,
// reading it ends in a total or in one refusal of its path, never in a panic.
func FuzzReadTotal(f *testing.F) {
	for _, seed := range []string{"2 + 2 + 2 = 6", "9 − 3 = 6", "A = ?", "=", "= 5", "(((", "3,5 = 3,5", "\u202e = 1"} {
		f.Add(seed, false)
		f.Add(seed, true)
	}

	f.Fuzz(func(t *testing.T, text string, comma bool) {
		decimals := picture.Point
		if comma {
			decimals = picture.Comma
		}
		total, problems := picture.ReadTotal(text, decimals)
		switch {
		case len(problems) > 1:
			t.Errorf("ReadTotal(%q) = %d problems, want one at most", text, len(problems))
		case len(problems) == 1 && (problems[0].Path != "solution_total" || total.Result != ""):
			t.Errorf("ReadTotal(%q) = %+v, %v; want a refusal of solution_total and no total", text, total, problems)
		case len(problems) == 0 && (total.Result == "" || !strings.HasSuffix(strings.TrimSpace(text), total.Result)):
			t.Errorf("ReadTotal(%q) = %+v, want it to end in its result", text, total)
		}
	})
}
