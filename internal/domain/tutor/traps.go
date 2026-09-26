package tutor

import (
	"cmp"
	"math"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// traps names the mistakes the wrong options should lead to: first the ones
// this child actually makes in this topic, and then, if there are not enough
// of those, the ones the reference tasks of the topic are built around.
//
// A child who has never met the topic has no mistakes of their own yet, and
// the task still needs wrong options worth offering — that is what the second
// source is for, the reference tasks of the task's own level first and then
// of the levels nearest it, until there are enough. A topic with no reference
// tasks anywhere takes the catalog's first traps, because a brief that names
// none is one the model hands back and the checks refuse.
func traps(p *profile.Profile, topic string, level rating.GradeLevel, catalog Catalog) []string {
	chosen := ownTraps(p, topic, catalog)
	if len(chosen) > Traps {
		chosen = chosen[:Traps]
	}

	for _, nearby := range levelsNearest(level) {
		if len(chosen) == Traps {
			return chosen
		}
		chosen = topUp(chosen, catalog.ExampleTraps(topic, nearby))
	}
	if len(chosen) == 0 {
		ids := catalog.TrapIDs()
		chosen = ids[:min(Traps, len(ids))]
	}
	return chosen
}

// topUp adds the traps chosen does not name yet, in their order, while a brief
// has room for them.
func topUp(chosen, more []string) []string {
	for _, trap := range more {
		if len(chosen) == Traps {
			break
		}
		if !slices.Contains(chosen, trap) {
			chosen = append(chosen, trap)
		}
	}
	return chosen
}

// levelsNearest is every level, the nearest to this one first and the lower of
// two equally near before the higher. A level that is none of the three is
// near none of them, and gets them from the youngest up.
func levelsNearest(level rating.GradeLevel) []rating.GradeLevel {
	levels := rating.GradeLevels()
	slices.SortStableFunc(levels, func(a, b rating.GradeLevel) int {
		return cmp.Or(
			cmp.Compare(math.Abs(a.Shift()-level.Shift()), math.Abs(b.Shift()-level.Shift())),
			cmp.Compare(a.Shift(), b.Shift()),
		)
	})
	return levels
}

// ownTraps are the traps this child has fallen for in this topic, the most
// frequent first.
//
// The counts live in a map, and a map is walked in a different order every
// time. That is why the walk below is over the catalog and the counts are
// only asked by key: the tie between two equally frequent traps is settled by
// the order the catalog lists them in rather than by chance. Without that the
// brief would differ from one run to the next on the same profile, which is
// the one thing the rule promises never to do.
func ownTraps(p *profile.Profile, topic string, catalog Catalog) []string {
	counts := p.Topics[topic].Traps
	if len(counts) == 0 {
		return nil
	}

	ordered := make([]string, 0, len(counts))
	for _, trap := range catalog.TrapIDs() {
		if counts[trap] > 0 {
			ordered = append(ordered, trap)
		}
	}
	slices.SortStableFunc(ordered, func(a, b string) int { return cmp.Compare(counts[b], counts[a]) })
	return ordered
}
