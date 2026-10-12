package tutor

import (
	"slices"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// ProbeEvery is how often a topic on pause comes back as a probe: one task in
// this many. A topic is put on pause when even its easiest task stands out of
// reach, and nothing the child answers elsewhere moves its own correction, so
// without a probe a topic the child stumbled in early would wait for the
// overall level alone, which the corridor keeps where it is.
const ProbeEvery = 6

// PausedTopics are the topics on pause for the child now, in catalog order:
// once the trial series is over, those the child has answered a task of, that
// the catalog teaches and that are out of reach. A topic the child has never
// answered is not on pause but not yet reached, and comes as the child grows.
// A pause belongs to the rule's turns, so none is on pause while they are not
// running: during the trial series; while the lessons are kept to a topic a
// person chose, which every task is set on; and while no topic at all is
// within reach, and the rule falls back on the easiest the catalog has, when
// every topic is out of reach and a probe of one would only stand in for
// another.
func PausedTopics(p *profile.Profile, catalog Catalog) []string {
	open, atTheBottom := reach(p, catalog)
	if atTheBottom || p.Ratings.InTrial() || LessonTopic(p, catalog) != "" {
		return nil
	}
	return pausedAmong(p, catalog, open)
}

// pausedAmong are the topics on pause, given the topics within reach.
func pausedAmong(p *profile.Profile, catalog Catalog, open []string) []string {
	var paused []string
	for _, topic := range catalog.TopicIDs() {
		if p.Topics[topic].Answers > 0 && len(catalog.LevelsOf(topic)) > 0 && !slices.Contains(open, topic) {
			paused = append(paused, topic)
		}
	}
	return paused
}

// Probes reports whether a task on this topic, chosen by mode, is a probe: set
// by the rule in a topic on pause now. What the task is follows from where the
// child stands rather than from anything kept with it, so the card that shows
// it and the answer that lifts it read it the same way.
func Probes(p *profile.Profile, catalog Catalog, topic string, mode profile.TutorMode) bool {
	return mode == profile.TutorRule && slices.Contains(PausedTopics(p, catalog), topic)
}

// JudgingOf is what the rule and the catalog know of the task on the card,
// which the profile records its answer by: the levels its topic is taught at,
// and whether it is a probe. With no task on the card there is nothing to
// judge, and an answer the task has had already is only told again.
func JudgingOf(p *profile.Profile, catalog Catalog) profile.Judging {
	task := p.CurrentTask
	if task == nil {
		return profile.Judging{}
	}
	return profile.Judging{
		Taught: catalog.LevelsOf(task.Topic),
		Probe:  task.Answered == nil && Probes(p, catalog, task.Topic, task.TutorMode),
	}
}

// probeDue says the task being chosen takes a probe's turn: the last of every
// ProbeEvery tasks behind the child, counted as the turns of the interests
// are, so that a task written ahead and one asked for in its place take the
// same turn. The turn of the sixth task is left out: with no task of the trial
// series skipped, it is the first after the series, which has only just found
// where the child stands. A skip counts among the tasks, so after skips in the
// series a probe can come on the first task after it.
func probeDue(p *profile.Profile) bool {
	tasks := tasksBehind(p)
	return !p.Ratings.InTrial() && tasks > rating.TrialAnswers && tasks%ProbeEvery == ProbeEvery-1
}
