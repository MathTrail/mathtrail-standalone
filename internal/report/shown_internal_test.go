package report

import (
	"maps"
	"slices"
	"testing"
	"time"
)

// A task is counted shown when the first question of a card once the hand-in
// that accepted it had begun brought it, counted from the hand-in's answer, and
// as nothing when the question the hand-in woke answered first: a later
// question that brings it again, a question that heard it still being
// written, one about another request, one before the hand-in began, one of a
// card drawn again long after, and a hand-in told again that accepted nothing
// are not counted.
func TestATaskIsShownByTheFirstQuestionThatBroughtItAfterItsHandIn(t *testing.T) {
	t.Parallel()

	at := func(seconds float64) time.Time {
		return time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC).Add(time.Duration(seconds * float64(time.Second)))
	}
	// The hand-in began at 8.5 s, and answered at 10 s.
	handIn := line{Message: eventToolCall, Tool: toolSubmitTask, Outcome: "ok", Screen: screenTask, TaskRequest: "req_a",
		InstructionsVersion: "v1", Client: "claude", Time: at(10), DurationMS: 1500}
	question := func(request, screen string, seconds float64) line {
		return line{Message: eventToolCall, Tool: toolReadTask, Outcome: "ok", Screen: screen, TaskRequest: request,
			InstructionsVersion: "v1", Client: "claude", Time: at(seconds)}
	}
	toldAgain := handIn
	toldAgain.Outcome, toldAgain.Time = "refused", at(20)

	cases := []struct {
		name  string
		lines []line
		want  map[group][]int64
	}{
		{
			name:  "the first question after the hand-in",
			lines: []line{handIn, question("req_a", screenTask, 30), question("req_a", screenTask, 10.5)},
			want:  map[group][]int64{{version: "v1", host: "claude"}: {500}},
		},
		{
			name:  "a question the hand-in woke, answered before the hand-in",
			lines: []line{question("req_a", screenTask, 9.8), handIn},
			want:  map[group][]int64{{version: "v1", host: "claude"}: {0}},
		},
		{
			name:  "a question that heard the task still being written",
			lines: []line{handIn, question("req_a", "coming", 10.2)},
			want:  map[group][]int64{},
		},
		{
			name:  "a question about another request",
			lines: []line{handIn, question("req_b", screenTask, 10.2)},
			want:  map[group][]int64{},
		},
		{
			name:  "a question before the hand-in began",
			lines: []line{question("req_a", screenTask, 8), handIn},
			want:  map[group][]int64{},
		},
		{
			name:  "a question of a card drawn again with its chat long after",
			lines: []line{handIn, question("req_a", screenTask, 10+20*60)},
			want:  map[group][]int64{},
		},
		{
			name:  "a hand-in told again",
			lines: []line{toldAgain, question("req_a", screenTask, 21)},
			want:  map[group][]int64{},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := shownAfterHandIn(tc.lines)
			if !maps.EqualFunc(got, tc.want, slices.Equal[[]int64]) {
				t.Errorf("shownAfterHandIn() = %v, want %v", got, tc.want)
			}
		})
	}
}
