package rating

import (
	"cmp"
	"math"
	"slices"
)

// GradeLevel is a band of school years whose tasks are written together: what
// a difficulty means, which topics exist and how long a sentence may be are
// decided for a level rather than for one year. The three levels stand on the
// scale a child's level is measured on, each shifted above the one before, so
// that their tasks make one ladder a child moves down and up by their answers.
type GradeLevel string

// The three grade levels, from the youngest up.
const (
	Grades12 GradeLevel = "1-2"
	Grades34 GradeLevel = "3-4"
	Grades56 GradeLevel = "5-6"
)

// Difficulties is how many difficulties a level has.
const Difficulties = 5

const (
	// levelShift is how far each level stands above the one before it. At
	// 2.5 the two hardest difficulties of a level overlap the two easiest of
	// the next and no two points of the ladder coincide: difficulty 5 of the
	// youngest level falls between difficulties 2 and 3 of the middle one.
	// Nothing has measured it yet, and the first children are what test it.
	levelShift = 2.5

	// gradesPerLevel is how many school years one level is written for.
	gradesPerLevel = 2
)

// gradeLevels are the levels in the order they stand on the ladder.
var gradeLevels = [...]GradeLevel{Grades12, Grades34, Grades56}

// GradeLevels lists the three levels, from the youngest up.
func GradeLevels() []GradeLevel { return slices.Clone(gradeLevels[:]) }

// GradeLevelOf is the level a school year falls into: 1 and 2 make the
// youngest, 3 and 4 the middle, 5 and 6 the oldest. A year outside 1 to 6 has
// no level, and the caller is told so rather than handed the nearest one.
func GradeLevelOf(grade int) (GradeLevel, bool) {
	if grade < 1 || grade > len(gradeLevels)*gradesPerLevel {
		return "", false
	}
	return gradeLevels[(grade-1)/gradesPerLevel], true
}

// Known reports whether this is one of the three levels.
func (l GradeLevel) Known() bool { return l.place() >= 0 }

// Shift is where the level stands on the scale: the β of its middle
// difficulty. A level that is none of the three stands nowhere, and its shift
// is no number, so that nothing worked out from it can pass for a place on the
// ladder.
func (l GradeLevel) Shift() float64 {
	place := l.place()
	if place < 0 {
		return math.NaN()
	}
	return levelShift * float64(place)
}

// FirstGrade is the youngest school year the level is written for, and zero
// for a level that is none of the three.
func (l GradeLevel) FirstGrade() int {
	place := l.place()
	if place < 0 {
		return 0
	}
	return place*gradesPerLevel + 1
}

// place is where the level stands among the three, from zero, or −1 for a
// level that is none of them.
func (l GradeLevel) place() int { return slices.Index(gradeLevels[:], l) }

// Point is one rung of the ladder: the level a task is written for and its
// difficulty inside that level.
type Point struct {
	GradeLevel GradeLevel
	Difficulty int
}

// Beta is where the point stands on the scale a child's level is measured on:
// the shift of its level, and its difficulty counted from the middle one, so
// that difficulty 3 of the youngest level sits at zero, 1 at −2 and 5 at +2. A
// point of a level that is none of the three stands nowhere, and its β is no
// number.
func (p Point) Beta() float64 {
	return p.GradeLevel.Shift() + float64(p.Difficulty-middleDifficulty)
}

// Points are every difficulty of these levels, the easiest first: the points a
// topic taught at those levels can be set at.
func Points(levels ...GradeLevel) []Point {
	points := make([]Point, 0, len(levels)*Difficulties)
	for _, level := range levels {
		for difficulty := 1; difficulty <= Difficulties; difficulty++ {
			points = append(points, Point{GradeLevel: level, Difficulty: difficulty})
		}
	}
	slices.SortStableFunc(points, easierFirst)
	return points
}

// easierFirst orders two points by where they stand on the ladder.
func easierFirst(a, b Point) int { return cmp.Compare(a.Beta(), b.Beta()) }

// Start is where a child of this grade begins: the shift of the level the
// grade falls into. It is the one thing the grade decides, and it is decided
// once — the answers move the child from there, and a grade changed later
// moves nothing. A grade outside 1 to 6 starts from the nearest level's shift
// rather than from no number at all.
func Start(grade int) float64 {
	level, _ := GradeLevelOf(min(max(grade, 1), len(gradeLevels)*gradesPerLevel))
	return level.Shift()
}

// InReach reports whether a child at this level solves a task at this point
// often enough to be set it: with a chance at least the corridor's floor, which
// is the point standing in the corridor or easier than it. A higher level only
// ever raises a chance, so a point in reach stays in reach as the child grows.
func InReach(level float64, point Point) bool {
	return Probability(level, point.Beta()) >= corridorLow
}
