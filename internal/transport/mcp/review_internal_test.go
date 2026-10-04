package mcpserver

import (
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
)

// Every reason and every kind of step is told in words of its own: a trap by
// the catalog's description and its advice, a step in a topic with the address
// of the part of the topic's page it is about, in the language asked for, a
// topic to begin with the strong one it builds on, and a step for every topic
// with no topic before it. A step of a kind with no
// words is left out. A review with nothing in it says so, and the trial series
// has none to tell.
func TestEveryReasonAndStepIsToldInWords(t *testing.T) {
	t.Parallel()

	loaded, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	s := &Service{content: loaded, site: "https://mathtrail.app"}

	for _, tc := range []struct {
		name     string
		language string
		review   *progress.Review
		want     string
	}{
		{"every reason, and the steps of a trap, a fall and a low topic", "en", &progress.Review{
			Strong: []progress.Judged{
				{Topic: "logic.ordering", Reasons: []progress.Reason{progress.ReasonMastered, progress.ReasonHigh, progress.ReasonRose}},
			},
			Develop: []progress.Judged{
				{Topic: "counting.gaps", Trap: "off_by_one", Reasons: []progress.Reason{
					progress.ReasonLow, progress.ReasonFailures, progress.ReasonHints, progress.ReasonTrap, progress.ReasonFell,
				}},
				{Topic: "arithmetic.tricks", Reasons: []progress.Reason{progress.ReasonFell}, Moving: true},
				{Topic: "pigeonhole.basic", Reasons: []progress.Reason{progress.ReasonLow}},
			},
			Early: []string{"time.clocks", "logic.knights_liars"},
			Steps: []progress.Step{
				{Kind: progress.StepTrap, Topic: "counting.gaps", Trap: "off_by_one"},
				{Kind: progress.StepRhythm, Topic: "arithmetic.tricks"},
				{Kind: progress.StepPractice, Topic: "pigeonhole.basic"},
				{Kind: progress.StepTrap, Trap: "double_count"},
			},
		}, "Review for the adult. Strong: Ordering (mastered; well above the overall level; rose this week). " +
			"To develop: Gaps and boundaries (well below the overall level; answered wrongly several times in a row; " +
			"the hint used in half its latest answers; keeps falling for this: Off by one when counting gaps, floors or saw cuts; " +
			"fell this week), Arithmetic with a trick (fell this week; its last answer was right, so a move has begun), " +
			"Pigeonhole principle (well below the overall level). " +
			"Too early to judge, which is no verdict yet: Clocks, Knights and liars. " +
			"What to do next: 1. Gaps and boundaries (https://mathtrail.app/en/topics/gaps-and-boundaries/#traps): " +
			"Before answering, draw a quick sketch: the posts as dots and the gaps between them, then count both. " +
			"2. Arithmetic with a trick (https://mathtrail.app/en/topics/arithmetic-with-a-trick/#home): " +
			"two or three short tasks a day, until it comes back. " +
			"3. Pigeonhole principle (https://mathtrail.app/en/topics/pigeonhole-principle/#home): " +
			"a few more tasks in it; MathTrail sets them where the child stands. " +
			"4. Write each option down once, in a fixed order, and look for repeats before counting them."},
		{"the hint alone, in a language the site is written in", "ru", &progress.Review{
			Strong:  []progress.Judged{},
			Develop: []progress.Judged{{Topic: "logic.knights_liars", Reasons: []progress.Reason{progress.ReasonHints}, Moving: true}},
			Early:   []string{},
			Steps:   []progress.Step{{Kind: progress.StepUnaided, Topic: "logic.knights_liars"}},
		}, "Review for the adult. To develop: Knights and liars (the hint used in half its latest answers; " +
			"its last answer was right, so a move has begun). What to do next: " +
			"1. Knights and liars (https://mathtrail.app/ru/topics/knights-and-liars/#home): try each task first without the hint."},
		{"a topic to begin, at the top of its page", "en", &progress.Review{
			Strong:  []progress.Judged{{Topic: "counting.gaps", Reasons: []progress.Reason{progress.ReasonMastered}}},
			Develop: []progress.Judged{}, Early: []string{},
			Steps: []progress.Step{{Kind: progress.StepBegin, Topic: "time.calendar", Base: "counting.gaps"}},
		}, "Review for the adult. Strong: Gaps and boundaries (mastered). What to do next: " +
			"1. Calendar and age (https://mathtrail.app/en/topics/calendar-and-age/): " +
			"a new topic to begin; its base, Gaps and boundaries, is a strength."},
		{"a step of a kind with no words", "en", &progress.Review{
			Strong: []progress.Judged{}, Develop: []progress.Judged{}, Early: []string{},
			Steps: []progress.Step{
				{Kind: progress.StepKind("dance"), Topic: "logic.ordering"},
				{Kind: progress.StepTrap, Trap: "double_count"},
			},
		}, "Review for the adult. What to do next: " +
			"1. Write each option down once, in a fixed order, and look for repeats before counting them."},
		{"nothing to say", "en", &progress.Review{
			Strong: []progress.Judged{}, Develop: []progress.Judged{}, Early: []string{}, Steps: []progress.Step{},
		}, "Review for the adult: nothing to point out yet."},
		{"the trial series", "en", nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := s.reviewText(tc.review, tc.language); got != tc.want {
				t.Errorf("reviewText() =\n%q\nwant\n%q", got, tc.want)
			}
		})
	}
}
