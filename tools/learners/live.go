package main

import (
	"math"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/tutor"
)

// The numbers the service's report reads off the log of live children, read
// off the simulated ones the same way, so that the two can be laid side by
// side: the later answers set against the earlier, and how soon a topic a
// child was shown as mastered is taken back. Each is read as the report reads
// a line of the log — the chance written to two places, the range of answers
// as the service names it, a mastery as the progress shows it — so that a
// difference between the bench and the live children is the children's, not
// the reading's.

// The ranges of the child's answers the later are set against the earlier
// in, as the report names them: from the first after the trial series to the
// 50th, and from the 101st to the 200th.
const (
	earlierTo = 50
	laterFrom = 101
	laterTo   = 200
)

// The answers in a topic after it was shown as mastered that the share taken
// back is read within, as the report reads it: the first few, and the first
// ten.
const (
	takenBackEarly = 5
	takenBackBy    = 10
)

// keptUp is how a child's answers in the two ranges came out from the chance
// they were promised: for each range, the differences — a right answer as 1,
// a wrong one as 0, less the chance — summed, over how many answers there
// were.
type keptUp struct{ earlier, later part }

// shownMastery is a topic a child was shown as mastered, as the report follows
// it through the log: how many of the child's answers in the topic it was
// followed through, at which of them it was taken back, none when it was not,
// and the run of wrong answers in the topic it stands at.
type shownMastery struct {
	followed, takenBackAt, wrongInARow int
}

// heldOn is, at one answer in a topic after it was shown as mastered, how many
// of a child's masteries were followed to it and how many it took back.
type heldOn struct {
	followed, takenBack int32
}

// showsMastered says whether the progress shows the topic as mastered for the
// child now, as the service decides whether to write that it does.
func (s *session) showsMastered(topic string) bool {
	return tutor.Mastered(s.p, s.w.catalog, topic)
}

// live measures an answer as the report reads its line, and the line of a
// topic mastered the service writes beside it when the answer brought the
// topic among those the progress shows mastered.
func (r *childResult) live(s *session, recorded *profile.Recorded, o *observedAnswer) {
	r.keepUp(recorded, s.p.Ratings.Answers)
	shown := recorded.Mastered && !o.shownBefore && s.showsMastered(recorded.Topic)
	r.followShown(recorded.Topic, recorded.Correct, shown)
}

// keepUp weighs an answer against the chance its task was promised, in the
// range of the child's answers it fell in, given how many the child has given
// with it. Only the answers the report weighs are weighed: after the trial
// series, and without the hint. Every task of the bench is the rule's.
func (r *childResult) keepUp(recorded *profile.Recorded, answers int) {
	if recorded.Trial != 0 || recorded.HintUsed {
		return
	}
	var in *part
	switch {
	case answers <= earlierTo:
		in = &r.keptUp.earlier
	case answers >= laterFrom && answers <= laterTo:
		in = &r.keptUp.later
	default:
		return
	}
	right := 0.0
	if recorded.Correct {
		right = 1
	}
	in.a += right - math.Round(recorded.Probability*100)/100
	in.b++
}

// followShown follows the masteries a child was shown through its answers, as
// the report follows them through the log. An answer in a topic whose mastery
// is followed moves it on, and takes it back at the second wrong answer in a
// row: a right answer, with the hint or without, ends the run. An answer that
// shows the topic mastered starts following it, and ends the mastery the topic
// was shown before, which was not taken back.
func (r *childResult) followShown(topic string, correct, shown bool) {
	if m := r.following[topic]; m != nil {
		m.followed++
		if correct {
			m.wrongInARow = 0
		} else {
			m.wrongInARow++
		}
		if m.wrongInARow >= profile.MasteryLostAfter {
			m.takenBackAt = m.followed
			delete(r.following, topic)
		}
	}
	if shown {
		m := &shownMastery{}
		r.following[topic] = m
		r.shown = append(r.shown, m)
	}
}

// heldOf counts a child's masteries at each of the first answers the share
// taken back is read within: a mastery counts at each answer it was followed
// to, and is taken back at the one it was taken back at. One the run stopped
// following early counts for the answers it was followed through.
func heldOf(r *childResult) [takenBackBy]heldOn {
	var on [takenBackBy]heldOn
	for _, m := range r.shown {
		for j := range min(m.followed, takenBackBy) {
			on[j].followed++
		}
		if m.takenBackAt > 0 && m.takenBackAt <= takenBackBy {
			on[m.takenBackAt-1].takenBack++
		}
	}
	return on
}

// liveMetrics are the numbers of the report that are a sum over what each
// child brings: how far the answers of each range came out from their promise,
// on average over the answers there, and how many masteries a child was shown.
func liveMetrics() []metric {
	return []metric{
		ratio("r9_kept_up_6_50", func(r *childResult) (num, den float64) { return r.keptUp.earlier.a, r.keptUp.earlier.b }),
		ratio("r9_kept_up_101_200", func(r *childResult) (num, den float64) { return r.keptUp.later.a, r.keptUp.later.b }),
		mean("r10_shown", func(r *childResult) (float64, bool) { return float64(len(r.shown)), true }),
	}
}

// livePooled are the numbers of the report read off the children as a whole:
// the later less the earlier, and the share of the masteries shown that are
// taken back within five answers in their topic and within ten.
func livePooled() []pooled {
	return []pooled{
		{name: "r9_kept_up_later_less_earlier", read: laterLessEarlier},
		{name: "r10_taken_back_5", read: takenBackWithin(takenBackEarly)},
		{name: "r10_taken_back_10", read: takenBackWithin(takenBackBy)},
	}
}

// laterLessEarlier is how much further the later answers came out from their
// promise than the earlier, over the children with answers in both ranges,
// each range on average over its answers: the report's difference, which sets
// each child against itself.
func laterLessEarlier(children []vector, sample []int) (float64, bool) {
	var both keptUp
	for _, c := range sample {
		in := children[c].keptUp
		if in.earlier.b == 0 || in.later.b == 0 {
			continue
		}
		both.earlier.a, both.earlier.b = both.earlier.a+in.earlier.a, both.earlier.b+in.earlier.b
		both.later.a, both.later.b = both.later.a+in.later.a, both.later.b+in.later.b
	}
	if both.earlier.b == 0 {
		return 0, false
	}
	return both.later.a/both.later.b - both.earlier.a/both.earlier.b, true
}

// takenBackWithin is the share of the masteries shown that are taken back
// within m answers in their topic: one less the Kaplan–Meier estimate of
// their being held that long, as the report reads it. A mastery the run stopped
// following early — at its end, or at its topic shown mastered again — counts
// for the answers it was followed through.
func takenBackWithin(m int) reader {
	return func(children []vector, sample []int) (float64, bool) {
		held, read := 1.0, false
		for j := range m {
			var followed, takenBack int
			for _, c := range sample {
				followed += int(children[c].held[j].followed)
				takenBack += int(children[c].held[j].takenBack)
			}
			if followed == 0 {
				break
			}
			read = true
			held *= 1 - float64(takenBack)/float64(followed)
		}
		return 1 - held, read
	}
}
