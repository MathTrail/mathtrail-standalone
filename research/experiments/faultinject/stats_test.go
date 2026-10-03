package main

import (
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
