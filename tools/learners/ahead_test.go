package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// writtenAheadOfTheAnswer is how the rule opens the rationale of a task chosen
// while the child still works on the task on the card.
const writtenAheadOfTheAnswer = "Written ahead, before the answer to the task on the card."

// Under the service's path written ahead, every task but a session's first is
// the one chosen while the task before it was on the card, before its answer
// and whatever that answer was. The service's own path chooses every task when
// it is asked for, and writes nothing ahead.
func TestATaskWrittenAheadIsTheOneChosenBeforeTheAnswer(t *testing.T) {
	t.Parallel()
	w, err := newWorld(40)
	if err != nil {
		t.Fatal(err)
	}
	for _, gen := range []generator{staticChildren, learning, jumping} {
		ahead := newSession(w, aheadRule(), newChild(gen, 0, w.topics))
		asked := newSession(w, serviceRule(), newChild(gen, 0, w.topics))
		for k := range w.answers {
			wantStepChosenAhead(t, fmt.Sprintf("%s, answer %d", gen, k+1), ahead, asked, k)
		}
	}
}

// wantStepChosenAhead takes answer k under the path written ahead and under
// the service's own, and holds the first to the task chosen while the task
// before it was on the card, the first answer aside, and the second to
// choosing nothing ahead.
func wantStepChosenAhead(t *testing.T, when string, ahead, asked *session, k int) {
	t.Helper()
	chosen := ahead.ahead
	if err := ahead.step(k); err != nil {
		t.Fatal(err)
	}
	if err := asked.step(k); err != nil {
		t.Fatal(err)
	}
	if (chosen == nil) != (k == 0) || asked.ahead != nil {
		t.Fatalf("%s: chosen ahead %v under the path written ahead and %v under the service's, "+
			"want every task but the first chosen ahead there and none here", when, chosen != nil, asked.ahead != nil)
	}
	if chosen != nil {
		wantAnsweredAsChosen(t, when, ahead, chosen)
	}
}

// wantAnsweredAsChosen holds the task a session's child last answered to the
// one chosen for it before the answer to the task before it.
func wantAnsweredAsChosen(t *testing.T, when string, s *session, chosen *chosen) {
	t.Helper()
	got, _ := s.p.LastAnswer()
	if !strings.HasPrefix(chosen.brief.Rationale, writtenAheadOfTheAnswer) || got.Topic != chosen.brief.TargetConcept ||
		got.GradeLevel != chosen.brief.GradeLevel || got.Difficulty != chosen.brief.Difficulty {
		t.Fatalf("%s: answered %s at %s/%d, want the task chosen before the answer before it, %+v",
			when, got.Topic, got.GradeLevel, got.Difficulty, chosen.brief)
	}
}

// The set written ahead runs the service's path twice on the same children —
// as the next task is asked for, and written ahead — and sets the one written
// ahead against the service's own, on children who stay put and on children
// who learn, and on how often the overall rank changes, its summary naming it
// for what it is. The run sets the seed and the name every other test draws
// with, so it does not run beside them.
func TestTheSetWrittenAheadSetsThePathAgainstTheService(t *testing.T) {
	out := t.TempDir()
	var stdout, stderr strings.Builder
	if code := runCommand([]string{"-out", out, "-children", "10", "-answers", "200", "-rules", "ahead"}, &stdout, &stderr); code != 0 {
		t.Fatalf("learners exited %d: %s", code, stderr.String())
	}
	dir := filepath.Join(out, "ahead")
	comparisons := readFile(t, filepath.Join(dir, "comparisons.csv"))
	for _, row := range []string{
		"1,r1_rms_200,shrinking/both/" + string(staticChildren) + ",shrinking_ahead/both/" + string(staticChildren) + ",",
		"2,r6_lag,shrinking/both/" + string(learning) + ",shrinking_ahead/both/" + string(learning) + ",",
		"5,r8_rank_6_20,shrinking/both/" + string(staticChildren) + ",shrinking_ahead/both/" + string(staticChildren) + ",",
		"5,r6_lag,shrinking/both/" + string(learningHalf) + ",shrinking_ahead/both/" + string(learningHalf) + ",",
		"5,r8_rank_150_200,shrinking/both/" + string(learningHalf) + ",shrinking_ahead/both/" + string(learningHalf) + ",",
	} {
		if !strings.Contains(comparisons, "\n"+row) {
			t.Errorf("comparisons.csv lacks a row opening %q:\n%s", row, comparisons)
		}
	}
	if summary := readFile(t, filepath.Join(dir, "summary.md")); !strings.Contains(summary,
		"| shrinking/both (the service) |") || !strings.Contains(summary, "| shrinking_ahead/both (the service, the next task written ahead) |") {
		t.Errorf("the summary names the service's two paths as\n%s\nwant both, the one written ahead as such", summary)
	}
}
