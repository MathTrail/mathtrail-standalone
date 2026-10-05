package main

import (
	"fmt"
	"math"
	"slices"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/report"
)

// oneTopic is a catalog of one topic, "t", taught at the first two grade
// levels.
type oneTopic struct{}

func (oneTopic) TopicIDs() []string { return []string{"t"} }
func (oneTopic) LevelsOf(string) []rating.GradeLevel {
	return []rating.GradeLevel{rating.Grades12, rating.Grades34}
}
func (oneTopic) TrapIDs() []string                               { return nil }
func (oneTopic) ExampleTraps(string, rating.GradeLevel) []string { return nil }

// An answer is weighed as the report weighs its line: after the trial series
// and without the hint, in the range of the child's answers it fell in — the
// first fifty, or the 101st to the 200th, and no other —, against the chance
// written to two places.
func TestAnAnswerIsWeighedAsTheReportWeighsItsLine(t *testing.T) {
	t.Parallel()

	r := newChildResult(nil)
	for _, a := range []struct {
		answers  int
		recorded profile.Recorded
	}{
		{3, profile.Recorded{Trial: 3, Correct: true, Probability: 0.6}},
		{6, profile.Recorded{Correct: true, Probability: 0.774}},
		{30, profile.Recorded{Correct: true, HintUsed: true, Probability: 0.8}},
		{50, profile.Recorded{Probability: 0.8}},
		{51, profile.Recorded{Correct: true, Probability: 0.8}},
		{100, profile.Recorded{Probability: 0.8}},
		{101, profile.Recorded{Correct: true, Probability: 0.6}},
		{200, profile.Recorded{Probability: 0.7}},
		{201, profile.Recorded{Correct: true, Probability: 0.8}},
	} {
		r.keepUp(&a.recorded, a.answers)
	}

	earlier, later := r.keptUp.earlier, r.keptUp.later
	if math.Abs(earlier.a-(0.23-0.8)) > 1e-12 || earlier.b != 2 || math.Abs(later.a-(0.4-0.7)) > 1e-12 || later.b != 2 {
		t.Errorf("weighed %+v earlier and %+v later, want {a:-0.57 b:2} and {a:-0.3 b:2}", earlier, later)
	}
}

// A mastery shown is followed as the report follows it: taken back at the
// second wrong answer in a row in its topic, a right answer between ending
// the run, and followed no further once it is; ended, not taken back, by its
// topic shown mastered again; and moved on by no answer in another topic.
func TestAMasteryShownIsFollowedAsTheReportFollowsIt(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, outcomes string
		// want is each mastery shown: the answers that followed it, and the
		// one it was taken back at, none when it was not.
		want [][2]int
	}{
		{"two wrong in a row", "RWW", [][2]int{{3, 3}}},
		{"a right answer between", "WRWR", [][2]int{{4, 0}}},
		{"answers after it was taken back", "WWRRW", [][2]int{{2, 2}}},
		{"shown again", "RWSWW", [][2]int{{3, 0}, {2, 2}}},
		{"another topic", "WxxW", [][2]int{{2, 2}}},
		{"no answer after it", "", [][2]int{{0, 0}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := newChildResult(nil)
			r.followShown("t", true, true)
			for _, outcome := range tc.outcomes {
				topic := "t"
				if outcome == 'x' {
					topic = "u"
				}
				r.followShown(topic, outcome == 'R' || outcome == 'S', outcome == 'S')
			}
			var got [][2]int
			for _, m := range r.shown {
				got = append(got, [2]int{m.Answers, m.TakenBackAt})
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("after %q the masteries are %v, want %v", tc.outcomes, got, tc.want)
			}
		})
	}
}

// A topic is shown mastered on the answer that brought it among those the
// progress shows: an answer the service declared a mastery on, while the
// progress did not show the topic before it and shows it after. A mastery the
// progress does not show — one held below the level the topic's tasks come
// from now — and a higher one declared while the topic was shown already
// start nothing.
func TestATopicIsShownMasteredAsTheServiceWritesThatItIs(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name                           string
		held                           rating.GradeLevel
		mastered, shownBefore, isShown bool
	}{
		{"declared and shown", rating.Grades34, true, false, true},
		{"declared below the level the tasks come from", rating.Grades12, true, false, false},
		{"declared while shown already", rating.Grades34, true, true, false},
		{"not declared", rating.Grades34, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := profile.New(profile.Student{Grade: 4, Pseudonym: "sim"}, "simulation", firstDay)
			p.Ratings.Theta = middleTaskOf(rating.Grades34) + levelAt(rating.CorridorMiddle)
			summary := p.Topics["t"]
			held, since := tc.held, profile.DateOf(firstDay)
			summary.MasteredLevel, summary.MasteredSince = &held, &since
			p.Topics["t"] = summary
			s := &session{w: &world{catalog: oneTopic{}}, p: p}
			r := newChildResult(nil)
			r.live(s, &profile.Recorded{Topic: "t", Correct: true, Mastered: tc.mastered}, &observedAnswer{shownBefore: tc.shownBefore})
			if shown := len(r.shown) == 1; shown != tc.isShown {
				t.Errorf("shown %d masteries, want one: %t", len(r.shown), tc.isShown)
			}
		})
	}
}

// The share taken back counts a mastery the run stopped following early for
// the answers it was followed through. Of four children's masteries, one is
// taken back at the second answer, one followed through two answers alone, one
// taken back at the fourth and one followed through ten: 5/8 are taken back
// within ten answers, as the report reads them, and 1/4 within three.
// Counting the one followed early as held would give 1/2 within ten.
func TestTheShareTakenBackIsReadAsTheReportReadsIt(t *testing.T) {
	t.Parallel()

	var children []vector
	for _, m := range []report.Followed{{Answers: 2, TakenBackAt: 2}, {Answers: 2}, {Answers: 4, TakenBackAt: 4}, {Answers: 10}} {
		r := newChildResult(nil)
		r.shown = []*report.Followed{&m}
		children = append(children, vector{held: heldOf(r)})
	}
	for within, want := range map[int]float64{report.TakenBackBy: 5.0 / 8, 3: 1.0 / 4} {
		if got, read := takenBackWithin(within)(children, everyone(len(children))); !read || math.Abs(got-want) > 1e-12 {
			t.Errorf("taken back within %d answers: %v (read: %t), want %v", within, got, read, want)
		}
	}
	if _, read := takenBackWithin(report.TakenBackBy)(nil, nil); read {
		t.Error("no masteries read a share, want none")
	}
}

// The later less the earlier is read over the children with answers in both
// ranges, each range on average over its answers, as the report reads it: a
// child with earlier answers alone is left out.
func TestTheLaterLessTheEarlierSetsEachChildAgainstItself(t *testing.T) {
	t.Parallel()

	children := []vector{
		{keptUp: keptUp{earlier: part{a: 1, b: 4}, later: part{a: 2, b: 4}}},
		{keptUp: keptUp{earlier: part{a: -1, b: 2}}},
		{keptUp: keptUp{earlier: part{a: 0, b: 4}, later: part{a: -1, b: 2}}},
	}
	if got, read := laterLessEarlier(children, everyone(len(children))); !read || math.Abs(got-(1.0/6-1.0/8)) > 1e-12 {
		t.Errorf("the later less the earlier is %v (read: %t), want %v", got, read, 1.0/6-1.0/8)
	}
	if _, read := laterLessEarlier(children[1:2], everyone(1)); read {
		t.Error("a child with no later answers gave a difference, want none")
	}
}

// What the report reads in the log as a mastery taken back is the service's
// own losing of it: on any sequence of answers, a mastery followed since the
// progress showed it is taken back on the very answer the service loses it on,
// and on no other.
func TestAMasteryTakenBackIsOneTheServiceLost(t *testing.T) {
	t.Parallel()
	w, err := newWorld(sequenceLength)
	if err != nil {
		t.Fatal(err)
	}
	takenBack := 0
	properties := gopter.NewProperties(nil)
	properties.Property("a mastery followed is taken back on the answer the service loses it on", prop.ForAll(
		func(seq sequence) string {
			why, taken := takenBackAsTheServiceLosesIt(w, &seq)
			takenBack += taken
			return why
		},
		genSequence(len(w.topics)),
	))
	properties.TestingRun(t)
	if takenBack == 0 {
		t.Error("the sequences took no mastery back, so the property shows nothing of it")
	}
}

// takenBackAsTheServiceLosesIt runs a sequence under the service, follows the
// masteries shown as the report does, and says where following them first
// parted from the service's losing them, or nothing when it never did, with
// how many it took back.
func takenBackAsTheServiceLosesIt(w *world, seq *sequence) (why string, taken int) {
	s := newSession(w, serviceRule(), &child{grade: seq.grade})
	for k, a := range seq.answers {
		topic := w.topics[seq.topics[a.topic]]
		ladder := w.ladders[topic]
		point := ladder[a.point%len(ladder)]
		before := observedAnswer{shownBefore: s.showsMastered(topic)}
		followed := s.result.following[topic]
		recorded, err := answerScripted(s, k, topic, point, a.draw < seq.skill, a.hint)
		if err != nil {
			return err.Error(), taken
		}
		s.result.live(s, &recorded, &before)
		tookBack := followed != nil && followed.TakenBackAt > 0
		if tookBack {
			taken++
		}
		if followed != nil && tookBack != recorded.Unmastered {
			return fmt.Sprintf("answer %d, %s: taken back %t, the service lost it %t", k+1, topic, tookBack, recorded.Unmastered), taken
		}
	}
	return "", taken
}
