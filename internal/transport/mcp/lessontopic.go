package mcpserver

import (
	"fmt"
	"slices"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// topicChoiceOut is what the card of a task needs to let the child or the
// adult keep the lessons to a topic: the topic chosen, or null while the rule
// chooses; the topics the review suggests, in its order; and the site the
// page of the topics is on, for the names of their groups to link to. Only the
// card reads it, so the model's words do not grow with it.
type topicChoiceOut struct {
	Chosen      *string  `json:"chosen"`
	Recommended []string `json:"recommended"`
	Site        *siteOut `json:"site"`
}

// topicChoiceOf is the choice of the topic as the card of a task shows it, or
// nil during the trial series: the series moves to a new topic each time, and
// a choice made then waits for its end, so the card offers none. What the
// review suggests is worked out as the progress works it out.
func (s *Service) topicChoiceOf(p *profile.Profile, now time.Time) (*topicChoiceOut, error) {
	if p.Ratings.InTrial() {
		return nil, nil
	}
	_, _, review, err := s.reviewed(p, profile.DateOf(now))
	if err != nil {
		return nil, err
	}
	return &topicChoiceOut{
		Chosen:      given(tutor.LessonTopic(p, s.content)),
		Recommended: review.Suggested(),
		Site:        s.siteOut(),
	}, nil
}

// siteOut is the site the topics' pages are on, as every payload names it.
func (s *Service) siteOut() *siteOut {
	return &siteOut{URL: s.site, Languages: slices.Clone(siteLanguages)}
}

// lessonTopicText says what the model needs to know of the topic of the
// lessons: which one the child or the adult chose, and whether the tasks keep
// to it yet. While the rule chooses — or the file names a topic the catalog
// does not have — there is nothing to say.
func (s *Service) lessonTopicText(p *profile.Profile) string {
	topic := tutor.ChosenTopic(p, s.content)
	switch {
	case topic == "":
		return ""
	case p.Ratings.InTrial():
		return fmt.Sprintf("The child or the adult chose to keep the lessons to %s once the trial series is over: "+
			"until then each task is on a new topic.", s.topicName(topic))
	}
	return fmt.Sprintf("The child or the adult keeps the lessons to %s: every task is on it, at the level and "+
		"difficulty the rule sets, until the choice is given back to the rule with an empty lesson_topic in "+
		"save_profile.", s.topicName(topic))
}

// topicChangedText is what the model is told when the card changed the topic
// of the lessons and nothing else: what the next tasks are on now, and that a
// task already being written keeps its own. The card's words asking the chat
// for a task on it follow these, unless the host refuses them; the model asks
// for the task when the child does, so that these words, read with whatever
// comes next instead, skip no task the child is still working on.
func (s *Service) topicChangedText(p *profile.Profile) string {
	lead := "On the card, the choice of the topic of the lessons was given back to the rule: it chooses the " +
		"topic of every task again."
	if p.Student.LessonTopic != "" {
		lead = "On the card, the child or the adult chose the topic of the lessons."
	}
	return joined(lead, s.lessonTopicText(p), s.topicStillText(p),
		"Do not explain the choice or say why the next task comes; ask for it with next_task when the child does.")
}

// topicStillText says that the task the child waits for keeps the topic it was
// asked on, when the lessons are now kept to another: a request open is handed
// back as it was, and the topic chosen starts with the task after it. A task
// being written ahead keeps nothing: it is let go once the next task is asked
// for.
func (s *Service) topicStillText(p *profile.Profile) string {
	lesson, request := tutor.LessonTopic(p, s.content), s.waitedOpen(p, s.now())
	if lesson == "" || request == nil || request.Brief.TargetConcept == lesson {
		return ""
	}
	return fmt.Sprintf("The task being written now stays on its own topic; the one after it comes on %s.", s.topicName(lesson))
}
