package progress

import (
	"cmp"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// FewestRepeats is the fewest answers a trap can be behind and still count as
// a mistake that repeats: a trap met once is a slip.
const FewestRepeats = 2

// Mistake is a trap the child keeps falling for: its catalog id, and how many
// answers of the history window it was behind.
type Mistake struct {
	Trap  string
	Times int
}

// Mistakes is the map of the child's misconceptions: the traps behind at least
// repeats answers of window, the profile's history window, the most frequent
// first and, of two as frequent, the one fallen for last first. The window is
// what makes a mistake one that repeats rather than one that once did: a trap
// the child has left behind drops out of it with the answers it was behind.
func Mistakes(window []profile.Answer, catalog tutor.Catalog, repeats int) []Mistake {
	counted := trapsIn(window, catalog)
	mistakes := []Mistake{}
	for trap, found := range counted {
		if found.times >= repeats {
			mistakes = append(mistakes, Mistake{Trap: trap, Times: found.times})
		}
	}
	slices.SortFunc(mistakes, func(a, b Mistake) int {
		if a.Times != b.Times {
			return cmp.Compare(b.Times, a.Times)
		}
		return cmp.Compare(counted[b.Trap].last, counted[a.Trap].last)
	})
	return mistakes
}

// Repeats says whether the child has fallen for trap at least repeats times
// among the answers of window, the profile's history window — a mistake that
// has come up before, and is worth a word of its own. It holds for exactly the
// traps on the map.
func Repeats(window []profile.Answer, catalog tutor.Catalog, trap string, repeats int) bool {
	return trapsIn(window, catalog)[trap].times >= repeats
}

// found is how many answers of a window a trap was behind, and the place of the
// last of them.
type found struct{ times, last int }

// trapsIn counts the traps of the catalog behind the answers of window. Only a
// wrong answer to an option has one: "I don't know" chose no option, and a task
// left without an answer has nothing to count. A trap the catalog no longer has
// is not counted at all, as a topic it no longer has is not in the progress.
func trapsIn(window []profile.Answer, catalog tutor.Catalog) map[string]found {
	known := map[string]bool{}
	for _, trap := range catalog.TrapIDs() {
		known[trap] = true
	}
	counted := map[string]found{}
	for at := range window {
		if trap := window[at].Trap; known[trap] {
			counted[trap] = found{times: counted[trap].times + 1, last: at}
		}
	}
	return counted
}
