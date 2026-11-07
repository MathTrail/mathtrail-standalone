package progress_test

import (
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/progress"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// pictured is a catalog with one topic all of whose tasks are solved from the
// picture.
type pictured struct {
	tutor.Catalog
	topic string
}

func (c pictured) SolvedFromPicture(topic string) bool { return topic == c.topic }

// topicIn is the topic of a summary by its id, and whether it is listed.
func topicIn(summary *progress.Summary, id string) (progress.Topic, bool) {
	for _, topic := range summary.Topics {
		if topic.ID == id {
			return topic, true
		}
	}
	return progress.Topic{}, false
}

// The progress reads which topics are within reach as the rule reads it in the
// chat it is read in: a topic solved from the picture is within reach where
// the card is drawn, and in a chat of words it is not, nor listed at all until
// the child has met it; one the child has met stays listed, out of reach.
func TestTheProgressOfAChatOfWordsLeavesTopicsSolvedFromThePictureOutOfReach(t *testing.T) {
	t.Parallel()

	p := fixture(t, "petya")
	unmet, met := withinReach(t, p)
	for _, topic := range []string{unmet, met} {
		catalog := pictured{Catalog: embedded(t), topic: topic}
		cards := summaryOf(t, p, catalog)
		if listed, found := topicIn(&cards, topic); !found || !listed.InReach {
			t.Errorf("a chat with cards: %s listed %v, %+v; want it within reach", topic, found, listed)
		}
		words, err := progress.Of(p, catalog, today, profile.ChatOfWords)
		if err != nil {
			t.Fatalf("Of() error = %v, want nil", err)
		}
		listed, found := topicIn(&words, topic)
		if found && (listed.InReach || listed.Paused) || found != (topic == met) {
			t.Errorf("a chat of words: %s listed %v, %+v; want it listed only once met, and never within reach",
				topic, found, listed)
		}
	}
}

// withinReach is a topic within reach of the child that the child has not
// met, and one the child has answered.
func withinReach(t *testing.T, p *profile.Profile) (unmet, met string) {
	t.Helper()

	whole := summaryOf(t, p, embedded(t))
	for _, topic := range whole.Topics {
		switch {
		case !topic.InReach:
		case topic.Answers == 0 && topic.Skipped == 0 && unmet == "":
			unmet = topic.ID
		case topic.Answers > 0 && met == "":
			met = topic.ID
		}
	}
	if unmet == "" || met == "" {
		t.Fatalf("the summary holds no topic within reach unmet and none met: %+v", whole.Topics)
	}
	return unmet, met
}
