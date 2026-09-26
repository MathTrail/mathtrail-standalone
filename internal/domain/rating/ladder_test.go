package rating_test

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// youngest is a difficulty of the youngest level, whose shift is zero: the
// level every number of the prototype was measured at.
func youngest(difficulty int) rating.Point {
	return rating.Point{GradeLevel: rating.Grades12, Difficulty: difficulty}
}

// The years a child can be in, and the levels they fall into.
func TestEveryGradeHasItsLevel(t *testing.T) {
	t.Parallel()

	for grade, want := range map[int]rating.GradeLevel{
		1: rating.Grades12, 2: rating.Grades12,
		3: rating.Grades34, 4: rating.Grades34,
		5: rating.Grades56, 6: rating.Grades56,
	} {
		if got, known := rating.GradeLevelOf(grade); !known || got != want {
			t.Errorf("GradeLevelOf(%d) = %q, %v, want %q", grade, got, known, want)
		}
	}

	// A year outside school is told apart rather than rounded to the nearest.
	for _, grade := range []int{0, -1, 7, 100} {
		if got, known := rating.GradeLevelOf(grade); known {
			t.Errorf("GradeLevelOf(%d) = %q, want no level at all", grade, got)
		}
	}
}

// The ladder as the specification prints it: every level 2.5 above the one
// before, difficulty 3 of the youngest at zero.
func TestTheLadder(t *testing.T) {
	t.Parallel()

	want := map[rating.GradeLevel][rating.Difficulties]float64{
		rating.Grades12: {-2, -1, 0, 1, 2},
		rating.Grades34: {0.5, 1.5, 2.5, 3.5, 4.5},
		rating.Grades56: {3, 4, 5, 6, 7},
	}
	for level, betas := range want {
		for difficulty := 1; difficulty <= rating.Difficulties; difficulty++ {
			point := rating.Point{GradeLevel: level, Difficulty: difficulty}
			if got := point.Beta(); got != betas[difficulty-1] {
				t.Errorf("β of difficulty %d of %s = %v, want %v", difficulty, level, got, betas[difficulty-1])
			}
		}
	}
}

// Neighbouring levels overlap and no two points coincide: that is what lets a
// child move from one level to the next without a gap and without two tasks of
// different levels meaning the same thing.
func TestNoTwoPointsOfTheLadderCoincide(t *testing.T) {
	t.Parallel()

	ladder := rating.Points(rating.GradeLevels()...)
	if len(ladder) != 3*rating.Difficulties {
		t.Fatalf("the ladder has %d points, want %d", len(ladder), 3*rating.Difficulties)
	}
	for i := 1; i < len(ladder); i++ {
		if ladder[i].Beta() <= ladder[i-1].Beta() {
			t.Errorf("%+v at %v does not stand above %+v at %v",
				ladder[i], ladder[i].Beta(), ladder[i-1], ladder[i-1].Beta())
		}
	}

	// The two hardest of a level stand above the easiest of the next.
	levels := rating.GradeLevels()
	for i := 1; i < len(levels); i++ {
		hardest := rating.Point{GradeLevel: levels[i-1], Difficulty: rating.Difficulties}
		easiest := rating.Point{GradeLevel: levels[i], Difficulty: 1}
		if hardest.Beta() <= easiest.Beta() {
			t.Errorf("%s does not overlap %s: its hardest at %v, the next level's easiest at %v",
				levels[i-1], levels[i], hardest.Beta(), easiest.Beta())
		}
	}
}

// A child begins at the shift of their grade's level, and a year outside
// school begins at the nearest one rather than nowhere.
func TestWhereAChildStarts(t *testing.T) {
	t.Parallel()

	for grade, want := range map[int]float64{-3: 0, 1: 0, 2: 0, 3: 2.5, 4: 2.5, 5: 5, 6: 5, 9: 5} {
		if got := rating.Start(grade); got != want {
			t.Errorf("Start(%d) = %v, want %v", grade, got, want)
		}
	}
}

// What else a level says about itself: the first year it is written for, and
// whether it is one of the three at all. A level that is none of them has no
// place on the ladder, and nothing computed from it can pass for one.
func TestWhatALevelSaysAboutItself(t *testing.T) {
	t.Parallel()

	for level, first := range map[rating.GradeLevel]int{rating.Grades12: 1, rating.Grades34: 3, rating.Grades56: 5} {
		if !level.Known() || level.FirstGrade() != first {
			t.Errorf("%s: known %v, first grade %d, want known and %d", level, level.Known(), level.FirstGrade(), first)
		}
	}
	for _, level := range []rating.GradeLevel{"", "7-8", "3-5", "1-2 "} {
		if level.Known() || level.FirstGrade() != 0 || !math.IsNaN(level.Shift()) {
			t.Errorf("%q: known %v, first grade %d, shift %v, want unknown, 0 and no number",
				level, level.Known(), level.FirstGrade(), level.Shift())
		}
		if beta := (rating.Point{GradeLevel: level, Difficulty: 3}).Beta(); !math.IsNaN(beta) {
			t.Errorf("β of a point of %q = %v, want no number", level, beta)
		}
	}
}

// The levels handed out are a copy: a caller that rewrites its list does not
// rewrite the ladder for everybody else.
func TestTheLevelsHandedOutAreACopy(t *testing.T) {
	t.Parallel()

	levels := rating.GradeLevels()
	levels[0] = "9-10"
	if again := rating.GradeLevels(); !slices.Equal(again, []rating.GradeLevel{rating.Grades12, rating.Grades34, rating.Grades56}) {
		t.Errorf("GradeLevels() = %v after a caller rewrote its copy", again)
	}
}

// A point is in reach from the level at which its chance is the corridor's
// floor: for the easiest point of each level, that level is the point's β
// plus 0.5108.
func TestAPointComesInReachAtTheFloorOfTheCorridor(t *testing.T) {
	t.Parallel()

	for _, level := range rating.GradeLevels() {
		easiest := rating.Point{GradeLevel: level, Difficulty: 1}
		t.Run(fmt.Sprint(level), func(t *testing.T) {
			t.Parallel()

			floor := easiest.Beta() + 0.5108
			if rating.InReach(floor-0.001, easiest) {
				t.Errorf("%+v is in reach at %v, below the corridor", easiest, floor-0.001)
			}
			if !rating.InReach(floor+0.001, easiest) {
				t.Errorf("%+v is out of reach at %v, in the corridor", easiest, floor+0.001)
			}
		})
	}
}
