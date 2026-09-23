package tutor

import (
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// traps names the mistakes the wrong options should lead to: first the ones
// this child actually makes in this topic, and then, if there are not enough
// of those, the ones the reference tasks of the topic are built around.
//
// A child who has never met the topic has no mistakes of their own yet, and
// the task still needs wrong options worth offering — that is what the second
// source is for.
func traps(p *profile.Profile, topic string, catalog Catalog) []string {
	chosen := ownTraps(p, topic, catalog)
	if len(chosen) > Traps {
		chosen = chosen[:Traps]
	}

	for _, trap := range catalog.ExampleTraps(topic, p.Student.Grade) {
		if len(chosen) == Traps {
			break
		}
		if !slices.Contains(chosen, trap) {
			chosen = append(chosen, trap)
		}
	}
	return chosen
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
	slices.SortStableFunc(ordered, func(a, b string) int { return counts[b] - counts[a] })
	return ordered
}
