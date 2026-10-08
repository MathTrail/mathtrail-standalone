package main

import (
	"encoding/base64"
	"math"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// The exact interval has closed forms at the edges: with no success in n the
// upper end is 1 − 0.025^(1/n), and with n of n the lower end is 0.025^(1/n).
func TestClopperPearsonAtTheEdges(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		k, n            int
		wantLow, wantHi float64
	}{
		{0, 10, 0, 1 - math.Pow(0.025, 0.1)},
		{10, 10, math.Pow(0.025, 0.1), 1},
		{0, 0, 0, 1},
	} {
		low, high := clopperPearson(test.k, test.n)
		if math.Abs(low-test.wantLow) > 1e-9 || math.Abs(high-test.wantHi) > 1e-9 {
			t.Errorf("clopperPearson(%d, %d) = (%v, %v), want (%v, %v)", test.k, test.n, low, high, test.wantLow, test.wantHi)
		}
	}
}

// Five of ten: the interval the tables of the exact method print, 0.187 to
// 0.813.
func TestClopperPearsonInTheMiddle(t *testing.T) {
	t.Parallel()
	low, high := clopperPearson(5, 10)
	if math.Abs(low-0.1871) > 1e-4 || math.Abs(high-0.8129) > 1e-4 {
		t.Errorf("clopperPearson(5, 10) = (%.4f, %.4f), want (0.1871, 0.8129)", low, high)
	}
}

func TestTheBootstrapIsRepeatableAndHoldsTheShare(t *testing.T) {
	t.Parallel()
	perHost := []share{{1, 2}, {2, 2}, {0, 1}, {3, 4}, {1, 1}}
	low, high := bootstrapByHost(perHost, seeded("test", "bootstrap"))
	again, againHigh := bootstrapByHost(perHost, seeded("test", "bootstrap"))
	if low != again || high != againHigh {
		t.Errorf("the same seed gave (%v, %v) and (%v, %v)", low, high, again, againHigh)
	}
	if pooled := 7.0 / 10; low > pooled || high < pooled || low < 0 || high > 1 {
		t.Errorf("interval (%v, %v) does not hold the pooled share %v", low, high, pooled)
	}
}

// The estimate from two sketches agrees with the service's own decision: a
// question and its copy are alike, two unrelated questions are not.
func TestSketchSimilarityMatchesTheService(t *testing.T) {
	t.Parallel()
	question := "Ann has 3 red pencils and 4 blue pencils. How many pencils does Ann have?"
	same, readable := sketchSimilarity(checks.Fingerprint(question, language), checks.Fingerprint(question, language))
	if !readable || same != 1 {
		t.Errorf("sketchSimilarity of a question with itself = %v, %v; want 1, true", same, readable)
	}
	other := "A clock strikes once at one o'clock and twice at two. How many strikes in a day?"
	apart, _ := sketchSimilarity(checks.Fingerprint(question, language), checks.Fingerprint(other, language))
	if apart >= checks.Threshold {
		t.Errorf("sketchSimilarity of two unrelated questions = %v, want below %v", apart, checks.Threshold)
	}
	if _, readable := sketchSimilarity("not base64!", checks.Fingerprint(question, language)); readable {
		t.Error("sketchSimilarity read a sketch that is not one")
	}
}

// The chance of at most k successes in n tries is certain from k = n on and
// none below k = 0, certain at a chance of 0 and none at a chance of 1 short of
// every try; in between it is the sum of the binomial terms: at most two of
// four fair tries is 11 in 16.
func TestTheBinomialChanceAtItsEdges(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		k, n    int
		p, want float64
	}{
		{"every try", 4, 4, 0.3, 1},
		{"more than every try", 5, 4, 0.3, 1},
		{"fewer than none", -1, 4, 0.3, 0},
		{"a chance of none", 2, 4, 0, 1},
		{"a certain chance", 2, 4, 1, 0},
		{"fair tries", 2, 4, 0.5, 11.0 / 16},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := binomialAtMost(test.k, test.n, test.p); math.Abs(got-test.want) > 1e-12 {
				t.Errorf("binomialAtMost(%d, %d, %v) = %v, want %v", test.k, test.n, test.p, got, test.want)
			}
		})
	}
}

// With no host to draw there is nothing to bound, and the interval is every
// share there is; an empty list has no value at any share of it.
func TestAnEmptyBootstrapBoundsNothing(t *testing.T) {
	t.Parallel()
	if low, high := bootstrapByHost(nil, seeded("test", "bootstrap")); low != 0 || high != 1 {
		t.Errorf("bootstrapByHost(nil) = (%v, %v), want (0, 1)", low, high)
	}
	if got := percentile(nil, 0.5); !math.IsNaN(got) {
		t.Errorf("percentile(nil) = %v, want NaN", got)
	}
}

// Two sketches are compared only when both can be read and they are of one
// length: anything else is no estimate at all, not an estimate of zero.
func TestSketchesThatCannotBeComparedGiveNoEstimate(t *testing.T) {
	t.Parallel()
	sketch := checks.Fingerprint("Ann has 3 red pencils and 4 blue pencils. How many pencils does Ann have?", language)
	short := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})
	for name, pair := range map[string][2]string{
		"the second unreadable": {sketch, "not base64!"},
		"the second shorter":    {sketch, short},
		"the first shorter":     {short, sketch},
		"both empty":            {"", ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if estimate, readable := sketchSimilarity(pair[0], pair[1]); readable || estimate != 0 {
				t.Errorf("sketchSimilarity = %v, %v; want 0, false", estimate, readable)
			}
		})
	}
}
