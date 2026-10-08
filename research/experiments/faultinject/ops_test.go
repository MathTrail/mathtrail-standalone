package main

import (
	"context"
	"errors"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// A defect that needs its host's own solver to confirm it is confirmed for the
// case its operator makes, and not for the host as it was; a solver that could
// not be run is an error, not a verdict either way.
func TestTheHostsSolverConfirmsTheDefect(t *testing.T) {
	t.Parallel()
	b := testBench(t)
	for _, c := range []struct {
		id      string
		confirm func(ctx context.Context, runner solver.Runner, m *mutant) (bool, error)
	}{
		{"D01a", solverMissesKey},
		{"D03a", solverFindsNothing},
	} {
		t.Run(c.id, func(t *testing.T) {
			t.Parallel()
			h, made := firstCase(t, b, operatorNamed(t, c.id))
			if confirmed, err := c.confirm(context.Background(), b.runner, &made); err != nil || !confirmed {
				t.Errorf("%s confirmed on %s = %v, %v; want true, nil", c.id, h.ID, confirmed, err)
			}
			unchanged := mutant{Sub: h.Base.Clone()}
			if confirmed, err := c.confirm(context.Background(), b.runner, &unchanged); err != nil || confirmed {
				t.Errorf("%s confirmed on %s as it was = %v, %v; want false, nil", c.id, h.ID, confirmed, err)
			}
			if _, err := c.confirm(cancelled(), b.runner, &made); !errors.Is(err, context.Canceled) {
				t.Errorf("%s confirmed with the run stopped = %v, want %v", c.id, err, context.Canceled)
			}
		})
	}
}

// The number put where the right answer stood is the nearest one no option
// says already, tried in the order n + 1, n − 1, n + 2, n − 2 and n + 3, an
// option saying it in any form; when all five are said there is none, and a
// text that is no number has none at all.
func TestTheNearestNumberNotSaidTakesTheAnswersPlace(t *testing.T) {
	t.Parallel()
	saying := func(texts ...string) map[string]bool {
		said := map[string]bool{}
		for _, text := range texts {
			said[solver.Key(text)] = true
		}
		return said
	}
	for _, c := range []struct {
		name, number string
		said         map[string]bool
		want         string
		found        bool
	}{
		{"nothing said", "5", saying(), "6", true},
		{"n + 1 said", "5", saying("6"), "4", true},
		{"n ± 1 said", "5", saying("6", "4"), "7", true},
		{"n ± 1 and n + 2 said", "5", saying("6", "4", "7"), "3", true},
		{"all but n + 3 said", "5", saying("6", "4", "7", "3"), "8", true},
		{"all five said", "5", saying("6", "4", "7", "3", "8"), "", false},
		{"said in another form", "5", saying("6.0", "+4"), "7", true},
		{"a decimal", "2.5", saying("3.5"), "1.5", true},
		{"a negative", "-1", saying("0"), "-2", true},
		{"no number", "five", saying(), "", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got, found := nearbyNumber(c.number, c.said); got != c.want || found != c.found {
				t.Errorf("nearbyNumber(%q) = %q, %v; want %q, %v", c.number, got, found, c.want, c.found)
			}
		})
	}
}

// Every option of every reference task, written another way, is the same
// answer to the service and another text: what makes the right answer appear
// twice without either option being a copy of the other.
func TestAnAnswerWrittenAnotherWayIsTheSameAnswer(t *testing.T) {
	t.Parallel()
	shipped, err := content.Load()
	if err != nil {
		t.Fatal(err)
	}
	for _, example := range shipped.Examples() {
		for letter, text := range example.Options {
			again := sameAnswer(text)
			if again == text || solver.Key(again) != solver.Key(text) {
				t.Errorf("%s %s: sameAnswer(%q) = %q, want another text with the key %q", example.ID, letter, text, again, solver.Key(text))
			}
		}
	}
}

// plainHost is a host of grades 5–6 without a drawing: three sentences with
// numbers in them, a whole number for the key, a trap of its own behind every
// wrong option, and a solver whose solve has a body.
func plainHost(id, topic string) *host {
	return &host{ID: id, Topic: topic, Level: rating.Grades56, Difficulty: 2, Base: submission{
		Task: checks.Task{
			Question:      "Ann has 3 coins. Ben has 4 coins. How many coins do they have together?",
			Options:       map[string]string{"A": "6", "B": "7", "C": "8", "D": "1", "E": "12"},
			CorrectAnswer: "B",
			Distractors: map[string]checks.Distractor{
				"A": {Trap: "t1", Text: "One coin was missed."}, "C": {Trap: "t2", Text: "One coin was counted twice."},
				"D": {Trap: "t3", Text: "That is the difference."}, "E": {Trap: "t4", Text: "That is the product."},
			},
		},
		Solver: "def solve(options):\n    return match(options, 3 + 4)\n",
	}}
}

// changed is a plain host with one thing about it changed.
func changed(change func(h *host)) *host {
	h := plainHost("h", "t1")
	change(h)
	return h
}

// inWordsHost is a plain host whose options are words rather than numbers.
func inWordsHost(id string) *host {
	return changed(func(h *host) {
		h.ID = id
		h.Base.Task.Options = map[string]string{"A": "red", "B": "blue", "C": "green", "D": "white", "E": "black"}
	})
}

// drawn gives a host a drawing of two labelled points and its structure.
func drawn(h *host) {
	h.Base.Task.Drawing = "A---B"
	h.Base.Task.DrawingStructure = &checks.DrawingStructure{Kind: "segment", Objects: []checks.DrawingObject{{ID: "a", Label: "A"}, {ID: "b", Label: "B"}}}
}

// poolOf is a pool of these hosts and no catalogs.
func poolOf(hosts ...*host) *pool { return &pool{hosts: hosts} }

// An operator that finds nothing in a host to put its defect into makes no
// case of it, rather than a case without the defect.
func TestAnOperatorWithNothingToChangeMakesNoCase(t *testing.T) {
	t.Parallel()
	noSolve := func(h *host) { h.Base.Solver = "def other(options):\n    return []\n" }
	ofAnotherLevel := changed(func(h *host) { h.ID, h.Level = "other", rating.Grades34 })
	for _, c := range []struct {
		name  string
		apply func(m *maker) (mutant, bool)
		on    *host
		pool  *pool
	}{
		{"a line into a program with no solve", intoSolve("injected = 1"), changed(noSolve), poolOf()},
		{"a line into a solve with only a comment", intoSolve("injected = 1"), changed(func(h *host) {
			h.Base.Solver = "def solve(options):\n    # to be written\n"
		}), poolOf()},
		{"a renamed solve where there is none", unnamedSolver, changed(noSolve), poolOf()},
		{"a missing answer with no other task", noRightOption, inWordsHost("h"), poolOf()},
		{"a missing answer every other task says", noRightOption, inWordsHost("h"), poolOf(inWordsHost("other"))},
		{"a copy onto a task with a drawing", copied(verbatim), changed(drawn), poolOf(plainHost("other", "t1"))},
		{"a copy with no other task of its level", copied(verbatim), plainHost("h", "t1"), poolOf(ofAnotherLevel)},
		{"a repeat that finds no number", repeated(bumpNumbers), changed(func(h *host) {
			h.Base.Task.Question = "Which coin is the odd one out?"
		}), poolOf()},
		{"a stray point in a question with no sentence", strayPoint, changed(func(h *host) {
			drawn(h)
			h.Base.Task.Question = " "
		}), poolOf()},
		{"a long sentence from one that ends in quotes", longSentence, changed(func(h *host) {
			h.Base.Task.Question = `Ann says: "I have 3 coins." Ben says: "I have 4." Who has more?`
		}), poolOf()},
		{"hard words in a question with no sentence", hardWords, changed(func(h *host) { h.Base.Task.Question = "" }), poolOf()},
		{"traps swapped where every trap is one", trapsSwapped, changed(func(h *host) {
			for letter, distractor := range h.Base.Task.Distractors {
				distractor.Trap = "t1"
				h.Base.Task.Distractors[letter] = distractor
			}
		}), poolOf()},
		{"a task of another topic where there is none", offTopic, plainHost("h", "t1"), poolOf(plainHost("other", "t1"))},
	} {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if made, applies := c.apply(newMaker(c.on, c.pool, "test")); applies {
				t.Errorf("the operator made a case of %q: %+v", c.on.Base.Task.Question, made.Sub.Task)
			}
		})
	}
}

// A right answer that is a number is replaced by the nearest number no option
// says, drawing nothing from the other tasks.
func TestANumberAnswerIsReplacedByTheNearestNumberNotSaid(t *testing.T) {
	t.Parallel()
	made, applies := noRightOption(newMaker(plainHost("h", "t1"), poolOf(), "D03a"))
	// 7 is the answer, and 8 and 6 are options already.
	if got := made.Sub.Task.Options["B"]; !applies || got != "9" {
		t.Errorf("noRightOption = %q, %v; want the answer replaced by 9", got, applies)
	}
}

// A copy is made from another task of the host's level and topic or, when the
// topic has no other, of its level alone; never from a task of another level.
func TestACopyFallsBackToATaskOfTheSameLevel(t *testing.T) {
	t.Parallel()
	sameLevel := plainHost("same-level", "t2")
	sameLevel.Base.Task.Question = "Kim has 5 balls. Tom has 2 balls. How many balls do they have?"
	otherLevel := changed(func(h *host) { h.ID, h.Level = "other-level", rating.Grades34 })
	made, applies := copied(verbatim)(newMaker(plainHost("h", "t1"), poolOf(otherLevel, sameLevel), "D14a"))
	if !applies || made.Donor != sameLevel.ID || made.Sub.Task.Question != sameLevel.Base.Task.Question {
		t.Errorf("copied = %+v, %v; want the question of %s", made, applies, sameLevel.ID)
	}
}

// A rewrite of a question says whether it found what it changes, and leaves
// the question as it was when it did not.
func TestARewriteSaysWhetherItFoundWhatItChanges(t *testing.T) {
	t.Parallel()
	beforeTheLast := func(text string) (string, bool) { return beforeLastSentence(text, longWords) }
	for _, test := range []struct {
		name     string
		rewrite  func(question string) (string, bool)
		question string
		want     string
		found    bool
	}{
		{"exactly made vague", madeVague, "Exactly 3 coins are red.", "About 3 coins are red.", true},
		{"exactly in capitals made vague", madeVague, "EXACTLY 3 coins are red.", "ABOUT 3 coins are red.", true},
		{"a number made vague", madeVague, "Ann has 3 coins.", "Ann has about 3 coins.", true},
		{"nothing to make vague", madeVague, "Ann has some coins.", "Ann has some coins.", false},
		{"a number only in the question", conditionRemoved, "Ann has coins. Has she 3 or 4?", "Ann has coins. Has she 3 or 4?", false},
		{"only two sentences to swap", firstTwoSwapped, "Ann has 3 coins. How many?", "Ann has 3 coins. How many?", false},
		{"no sentence to come before", beforeTheLast, "", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got, found := test.rewrite(test.question); got != test.want || found != test.found {
				t.Errorf("%q became %q, %v; want %q, %v", test.question, got, found, test.want, test.found)
			}
		})
	}
}
