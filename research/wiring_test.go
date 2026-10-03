// Package research_test holds the check the whole research module rests on:
// experiments call the product's own packages, including those under
// internal/, instead of re-implementing them.
package research_test

import (
	"math"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// A fresh child meeting a task of the middle difficulty is expected to answer
// correctly with probability 0.2 + 0.8·σ(0) = 0.6. The value only proves that
// the research module compiles against the product's code as it is.
func TestProductPackagesImport(t *testing.T) {
	t.Parallel()

	got := rating.Probability(0, rating.Point{GradeLevel: rating.Grades12, Difficulty: 3}.Beta())
	if math.Abs(got-0.6) > 1e-12 {
		t.Fatalf("the chance at difficulty 3 of grades 1–2 = %v, want 0.6", got)
	}
}
