package main

import (
	"context"
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"slices"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// masterSeed is where every random choice of the experiment comes from.
const masterSeed = 20261001

// experiment names the experiment in every seed it draws.
const experiment = "E-A1"

// kind says what a class of defects tests.
type kind string

const (
	// mechanism is a defect that breaks a rule a check states in its own
	// terms: a check that holds its rule refuses every such case.
	mechanism kind = "mechanism"
	// discovery is a defect a child would see whose refusal follows from no
	// check's definition.
	discovery kind = "discovery"
	// outOfScope is a defect the service is not built to see.
	outOfScope kind = "out_of_scope"
)

// operator turns one host into one case carrying one defect.
type operator struct {
	ID       string
	Class    string
	Kind     kind
	Expected []checks.Code
	// Apply makes the case, and says whether the operator applies to this
	// host at all.
	Apply func(m *maker) (mutant, bool)
	// Valid confirms with the host's own solver that a case is the defect it
	// is named after; nil when the operator needs no such confirmation.
	Valid func(ctx context.Context, runner solver.Runner, m *mutant) (bool, error)
}

// mutant is a host with one defect in it.
type mutant struct {
	Sub submission
	// LeaveOut are the reference questions left out of the comparison beside
	// the host's own: a donor whose whole task the case took.
	LeaveOut []string
	// Source is the question the case's question was made from, for the
	// similarity between them; empty when it was not made from one.
	Source string
	// Donor is the reference task the case took something from.
	Donor string
}

// pool is everything an operator may draw on besides its host: the shipped
// catalogs and the other eligible hosts.
type pool struct {
	content *content.Content
	hosts   []*host
}

// donors are the eligible hosts other than this one that keep, in id order.
func (p *pool) donors(h *host, keep func(*host) bool) []*host {
	var found []*host
	for _, other := range p.hosts {
		if other.ID != h.ID && keep(other) {
			found = append(found, other)
		}
	}
	return found
}

// maker is what an operator works with: its host, the pool and the case's own
// random choices.
type maker struct {
	host *host
	pool *pool
	rand *rand.Rand
}

// newMaker prepares the case of one operator on one host, with the random
// choices that case and no other draws.
func newMaker(h *host, p *pool, operatorID string) *maker {
	return &maker{host: h, pool: p, rand: seeded(h.ID+"/"+operatorID, "operator")}
}

// seeded is the random stream of one unit of the experiment for one purpose.
// The choices it makes pick letters, donors and types for a defect, and need
// to be repeatable rather than unpredictable.
func seeded(unit, purpose string) *rand.Rand {
	hash := fnv.New64a()
	fmt.Fprintf(hash, "%d|%s|%s|%s", masterSeed, experiment, unit, purpose)
	return rand.New(rand.NewPCG(hash.Sum64(), masterSeed)) //nolint:gosec // a seeded, repeatable choice of defects, not a secret
}

// start is a fresh copy of the host's submission to put a defect into.
func (m *maker) start() submission { return m.host.Base.Clone() }

// choose picks one of these, by the case's seed.
func choose[T any](m *maker, from []T) T { return from[m.rand.IntN(len(from))] }

// wrongLetters are the letters of the host's wrong options, in order.
func (m *maker) wrongLetters() []string {
	return slices.DeleteFunc(solver.Letters(), func(letter string) bool { return letter == m.host.Base.Task.CorrectAnswer })
}

// wrongLetter is one of the host's wrong options, by the case's seed.
func (m *maker) wrongLetter() string { return choose(m, m.wrongLetters()) }

// operators are every operator of the experiment, in the order of their
// classes.
func operators() []operator {
	return slices.Concat(answerOperators(), formOperators(), copyOperators(), drawingOperators(), scopeOperators())
}
