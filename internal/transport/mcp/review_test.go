package mcpserver_test

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// reviewPayload is the review of the progress, as a card reads it.
type reviewPayload struct {
	Review *struct {
		Strong []struct {
			Topic   string   `json:"topic"`
			Reasons []string `json:"reasons"`
		} `json:"strong"`
		Develop []struct {
			Topic   string   `json:"topic"`
			Reasons []string `json:"reasons"`
			Trap    string   `json:"trap"`
			Moving  bool     `json:"moving"`
		} `json:"develop"`
		Early []string `json:"early"`
		Steps []struct {
			Kind  string `json:"kind"`
			Topic string `json:"topic"`
			Trap  string `json:"trap"`
		} `json:"steps"`
	} `json:"review"`
}

// reviewedPetya is Petya long past the trial series: Ordering mastered and
// strong, Gaps and boundaries falling for one gap too few again and again,
// Clocks met only twice, and his other topics level with the overall one.
func reviewedPetya(p *profile.Profile) {
	p.Ratings.Answers = 40
	for id, topic := range p.Topics {
		topic.Answers, topic.Correct, topic.Delta = 6, 4, 0
		topic.WrongStreak, topic.TopStreak = 0, 0
		topic.MasteredSince, topic.MasteredLevel = nil, nil
		p.Topics[id] = topic
	}
	ordering := p.Topics["logic.ordering"]
	level, since := rating.Grades34, profile.DateOf(lessonDay.AddDate(0, 0, -3))
	ordering.MasteredSince, ordering.MasteredLevel = &since, &level
	p.Topics["logic.ordering"] = ordering
	p.Topics["time.clocks"] = profile.Topic{Answers: 2, Correct: 1}
	missed := p.Recent[len(p.Recent)-1]
	missed.Topic, missed.Correct, missed.Chosen, missed.Trap = "counting.gaps", false, "B", "off_by_one"
	p.Recent = append(p.Recent, missed, missed)
	gaps := p.Topics["counting.gaps"]
	gaps.WrongStreak = 2
	p.Topics["counting.gaps"] = gaps
}

// The progress past the trial series carries the review, by ids and codes, and
// tells it in words for the adult, each step with its advice — the same
// whichever tool asks.
func TestTheProgressCarriesTheReviewAndTellsIt(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, keptAs(t, "petya", reviewedPetya))
	forModel, forCard := call(t, session, "get_progress", nil), call(t, session, "read_progress", nil)
	if !bytes.Equal(rawPayload(t, forModel), rawPayload(t, forCard)) {
		t.Errorf("get_progress and read_progress differ:\n%s\n%s", rawPayload(t, forModel), rawPayload(t, forCard))
	}

	review := payloadOf[reviewPayload](t, forModel).Review
	if review == nil {
		t.Fatalf("the progress carries no review: %s", rawPayload(t, forModel))
	}
	if len(review.Strong) != 1 || review.Strong[0].Topic != "logic.ordering" ||
		!reflect.DeepEqual(review.Strong[0].Reasons, []string{"mastered"}) {
		t.Errorf("strong = %+v, want Ordering, mastered", review.Strong)
	}
	if len(review.Develop) != 1 || review.Develop[0].Topic != "counting.gaps" || review.Develop[0].Trap != "off_by_one" ||
		!reflect.DeepEqual(review.Develop[0].Reasons, []string{"failures", "trap"}) || review.Develop[0].Moving {
		t.Errorf("develop = %+v, want Gaps and boundaries, wrong in a row and falling for off_by_one", review.Develop)
	}
	if !reflect.DeepEqual(review.Early, []string{"time.clocks"}) {
		t.Errorf("early = %v, want Clocks", review.Early)
	}
	if len(review.Steps) != 1 || review.Steps[0].Kind != "trap" || review.Steps[0].Topic != "counting.gaps" ||
		review.Steps[0].Trap != "off_by_one" {
		t.Errorf("steps = %+v, want the advice for off_by_one in Gaps and boundaries alone", review.Steps)
	}

	words := textOf(t, forModel)
	for _, want := range []string{
		"Review for the adult. Strong: Ordering (mastered).",
		"To develop: Gaps and boundaries (answered wrongly several times in a row; keeps falling for this: Off by one when counting gaps, floors or saw cuts).",
		"Too early to judge, which is no verdict yet: Clocks.",
		"What to do next: 1. Gaps and boundaries: Before answering, draw a quick sketch: the posts as dots and the gaps between them, then count both.",
	} {
		if !strings.Contains(words, want) {
			t.Errorf("the words are %q, want them to say %q", words, want)
		}
	}
}

// During the trial series there is no review, in the payload or in words, and
// the mistakes that repeat are still told.
func TestTheTrialSeriesHasNoReview(t *testing.T) {
	t.Parallel()

	_, session := lesson(t, keptWith(t, "masha"))
	result := call(t, session, "get_progress", nil)

	if raw := rawPayload(t, result); bytes.Contains(raw, []byte(`"review"`)) {
		t.Errorf("the progress of the trial series is %s, want no review in it", raw)
	}
	if words := textOf(t, result); strings.Contains(words, "Review") {
		t.Errorf("the words are %q, want no review during the trial series", words)
	}
}
