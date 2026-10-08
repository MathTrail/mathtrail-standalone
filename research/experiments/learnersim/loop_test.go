package main

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// emptyCatalog is a catalog that teaches nothing, from which no task can be
// set.
type emptyCatalog struct{}

func (emptyCatalog) TopicIDs() []string                              { return nil }
func (emptyCatalog) LevelsOf(string) []rating.GradeLevel             { return nil }
func (emptyCatalog) TrapIDs() []string                               { return nil }
func (emptyCatalog) ExampleTraps(string, rating.GradeLevel) []string { return nil }

// rulesOf are the distinct rules of some cells, each with the first cell it is
// run in.
func rulesOf(all []cell) []cell {
	seen := map[*rule]bool{}
	var first []cell
	for _, c := range all {
		if !seen[c.rule] {
			seen[c.rule] = true
			first = append(first, c)
		}
	}
	return first
}

// Every rule of the experiment starts a child where the service does, at the
// start of the child's grade overall and in every topic, so that the rules
// differ only in what they make of the answers.
func TestEveryRuleStartsAtTheStartOfTheChildsGrade(t *testing.T) {
	t.Parallel()
	w := &world{topics: []string{"a", "b", "c"}, answers: 1}
	for _, c := range rulesOf(cells()) {
		child := newChild(c.generator, 0, w.topics)
		s := newSession(w, c.rule, child)
		start := rating.Start(child.grade)
		if got := s.est.overall(); got != start {
			t.Errorf("%s: the overall level starts at %v, want %v", c.name(), got, start)
		}
		for _, topic := range w.topics {
			if got := s.est.level(topic); got != start {
				t.Errorf("%s: the level in %s starts at %v, want %v", c.name(), topic, got, start)
			}
		}
	}
}

// Under every rule but the service's own, an answer moves no level of the rule
// against the outcome, and the levels the harness writes into the profile,
// where the service's rule and its mastery read them, are the rule's own.
func TestEveryRulesLevelsFollowTheOutcomeIntoTheProfile(t *testing.T) {
	t.Parallel()
	w, err := newWorld(30)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range rulesOf(cells()) {
		if c.rule.service {
			continue
		}
		t.Run(c.name(), func(t *testing.T) {
			t.Parallel()
			followTheOutcome(t, w, &c)
		})
	}
}

// followTheOutcome runs the first child of a cell through its answers, and
// fails the test at an answer that moves a level of the rule against the
// outcome, at levels written into the profile that are not the rule's, and
// when no answer moved a level at all.
func followTheOutcome(t *testing.T, w *world, c *cell) {
	t.Helper()
	s := newSession(w, c.rule, newChild(c.generator, 0, w.topics))
	moves := 0
	for k := range w.answers {
		before := sessionLevels(s)
		scored := s.result.observed
		if err := s.step(k); err != nil {
			t.Fatal(err)
		}
		moves += checkFollowed(t, s, before, s.result.observed > scored, k)
		// The harness writes the levels before the next task is chosen;
		// writing them now, when it would anyway, changes nothing it does.
		if s.writes(k + 1) {
			s.writeLevels()
			checkWritten(t, s, k)
		}
	}
	if moves == 0 {
		t.Errorf("no answer of %d moved a level", w.answers)
	}
}

// sessionLevels are the rule's levels in every topic of a session.
func sessionLevels(s *session) []float64 {
	levels := make([]float64, 0, len(s.w.topics))
	for _, topic := range s.w.topics {
		levels = append(levels, s.est.level(topic))
	}
	return levels
}

// checkFollowed fails the test when an answer moved a level of the rule
// against its outcome, and says how many levels it moved.
func checkFollowed(t *testing.T, s *session, before []float64, correct bool, k int) int {
	t.Helper()
	moves := 0
	for i, level := range sessionLevels(s) {
		moved := level - before[i]
		if correct && moved < -1e-12 || !correct && moved > 1e-12 {
			t.Fatalf("answer %d, correct %v, moved the level in %s by %v", k+1, correct, s.w.topics[i], moved)
		}
		if moved != 0 {
			moves++
		}
	}
	return moves
}

// checkWritten fails the test when the profile's levels are not the rule's.
func checkWritten(t *testing.T, s *session, k int) {
	t.Helper()
	if got, want := s.p.Ratings.Theta, s.est.overall(); got != want {
		t.Fatalf("after answer %d the profile's overall level is %v, the rule's %v", k+1, got, want)
	}
	for _, topic := range s.w.topics {
		if got, want := s.p.LevelIn(topic), s.est.level(topic); math.Abs(got-want) > 1e-12 {
			t.Fatalf("after answer %d the profile's level in %s is %v, the rule's %v", k+1, topic, got, want)
		}
	}
}

// A run in which the service can set no task, or seal none, fails at the first
// answer, naming the child, the rule and the answer, rather than answering a
// task nobody set or one whose answer lies in the open.
func TestARunFailsAtTheFirstAnswerItCannotSetOrSeal(t *testing.T) {
	t.Parallel()
	shipped, err := newWorld(3)
	if err != nil {
		t.Fatal(err)
	}
	unsealed := *shipped
	unsealed.sealer = failingSealer{}
	cases := []struct {
		name  string
		w     *world
		cause string
	}{
		{"no task to set", &world{catalog: emptyCatalog{}, sealer: shipped.sealer, answers: 3}, "tutor: "},
		{"a task that cannot be sealed", &unsealed, errCannotSeal.Error()},
	}
	service := &rule{name: "shrinking", shape: both, service: true}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, failed := run(tc.w, service, newChild(staticChildren, 0, tc.w.topics))
			if want := "G0/0 under shrinking, answer 1: "; failed == nil || !strings.HasPrefix(failed.Error(), want) || !strings.Contains(failed.Error(), tc.cause) {
				t.Errorf("run() error = %v, want one that begins %q and gives the cause %q", failed, want, tc.cause)
			}
		})
	}
	if _, failed := run(&unsealed, service, newChild(staticChildren, 0, unsealed.topics)); !errors.Is(failed, errCannotSeal) {
		t.Errorf("run() error = %v, want the sealer's %v", failed, errCannotSeal)
	}
}
