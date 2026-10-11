package main

import (
	"math"
	"slices"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
)

// For the service, the move a run records after an answer is the move the
// card shows: the topic's rating as the card draws it before the answer and
// after it, from the 6th answer on, in each window.
func TestTheCardsMoveIsTheCardsOwn(t *testing.T) {
	t.Parallel()
	w, err := newWorld(200)
	if err != nil {
		t.Fatal(err)
	}
	s := newSession(w, ruleNamed(t, "shrinking"), newChild(learningHalf, 0, w.topics))
	var want [len(screenWindows)][]uint16
	for k := range w.answers {
		before := map[string]float64{}
		for _, topic := range w.topics {
			before[topic] = s.p.LevelIn(topic)
		}
		if err := s.step(k); err != nil {
			t.Fatal(err)
		}
		at := windowOf(k)
		if at < 0 {
			continue
		}
		last, _ := s.p.LastAnswer()
		shown := func(level float64) int { return rating.Shown(rating.Elo(level)) }
		move := shown(s.p.LevelIn(last.Topic)) - shown(before[last.Topic])
		want[at] = append(want[at], uint16(max(move, -move)))
	}
	for at, moves := range s.result.screen.moves {
		if !slices.Equal(moves, want[at]) {
			t.Errorf("answers %s: recorded moves %v, want the card's %v", screenWindows[at].label(), moves, want[at])
		}
	}
}

// The percentile of the card's moves is read by the nearest rank over every
// answer of the sample's children together: of the moves 1 to 100, the 95th
// percentile is 95; a child drawn twice counts twice; a sample with no move
// gives no number.
func TestTheCardsMovesAreReadAtTheirPercentile(t *testing.T) {
	t.Parallel()
	var low, high vector
	for move := 1; move <= 100; move++ {
		if move <= 50 {
			low.moves[0] = append(low.moves[0], uint16(move))
		} else {
			high.moves[0] = append(high.moves[0], uint16(move))
		}
	}
	read := movesAt(0, 95)
	children := []vector{low, high}
	for _, tc := range []struct {
		sample []int
		want   float64
	}{
		{[]int{0, 1}, 95},
		{[]int{0, 0, 1}, 93},
		{[]int{1, 1}, 98},
	} {
		if got, has := read(children, tc.sample); !has || got != tc.want {
			t.Errorf("the sample %v: %v (read %v), want %v", tc.sample, got, has, tc.want)
		}
	}
	if got, has := read([]vector{{}}, []int{0}); has {
		t.Errorf("a sample with no move: %v, want no number", got)
	}
}

// stub is an estimator that stands where a test puts it, overall and in every
// topic.
type stub struct {
	at float64
}

func (s *stub) overall() float64               { return s.at }
func (s *stub) level(string) float64           { return s.at }
func (s *stub) chance(string, float64) float64 { return rating.Guess }
func (s *stub) answered(string, float64, bool) {}

// levelShownAs is the level whose rating the card shows as a number of
// points.
func levelShownAs(points float64) float64 { return (points - 1500) * math.Ln10 / 400 }

// A wrong answer's fall and a right answer that leaves the card where it stood
// are counted apart, by the card's own whole points: a right answer up 3, one
// that moves the rating less than a point, a wrong answer down 7, another down
// 2 and a right one that rounds to no move are falls of 9 over two wrong
// answers and two right answers of three that showed nothing.
func TestAWrongAnswersFallAndAStillCardAreCountedApart(t *testing.T) {
	t.Parallel()
	est := &stub{at: levelShownAs(1600)}
	s := &session{est: est}
	r := &childResult{}
	for i, a := range []struct {
		points  float64
		correct bool
	}{{1603, true}, {1603.2, true}, {1596, false}, {1594, false}, {1594.3, true}} {
		before := est.at
		est.at = levelShownAs(a.points)
		r.screen.after(s, "a", &observedAnswer{estimate: before, overall: before, correct: a.correct}, 5+i)
	}
	if num, den := fallAfterWrong(0)(r); num != 9 || den != 2 {
		t.Errorf("fallen %v points over %v wrong answers, want 9 over 2", num, den)
	}
	if num, den := stillAfterRight(0)(r); num != 2 || den != 3 {
		t.Errorf("%v right answers of %v left the card still, want 2 of 3", num, den)
	}
	if num, den := fallAfterWrong(1)(r); num != 0 || den != 0 {
		t.Errorf("a window with no answer: fallen %v over %v, want nothing", num, den)
	}
}

// Every run reads the card's fall after a wrong answer and the right answers
// that leave it still in both windows, as measures of what the child is shown,
// which say nothing of the ceiling.
func TestARunReadsWhatAWrongAndARightAnswerDoToTheCard(t *testing.T) {
	t.Parallel()
	ceiling := ceilingRule()
	for _, w := range screenWindows {
		for _, name := range []string{"r8_wrong_fall_" + w.name(), "r8_right_still_" + w.name()} {
			at := slices.IndexFunc(metrics(), func(m metric) bool { return m.name == name })
			if at < 0 {
				t.Errorf("no run reads %s", name)
				continue
			}
			if m := metrics()[at]; m.applies == nil || m.applies(ceiling) {
				t.Errorf("%s is read off the ceiling, want it read off the rules alone", name)
			}
		}
	}
}

// A rank changes when the rating crosses a rank's floor, and only then, and
// the changes are counted in a hundred answers of their window: a rating that
// steps across the floor of 1666 and back changes rank twice in five answers,
// forty times a hundred answers, and the card's moves are the steps.
func TestRankChangesAreCountedAcrossAFloor(t *testing.T) {
	t.Parallel()
	est := &stub{at: levelShownAs(1660)}
	s := &session{est: est}
	r := &childResult{}
	ratings := []float64{1664, 1667, 1670, 1665, 1662}
	for i, points := range ratings {
		before := est.at
		est.at = levelShownAs(points)
		r.screen.after(s, "a", &observedAnswer{estimate: before, overall: before}, 5+i)
	}
	if got := r.screen; got.rankChanges[0] != 2 || got.topicChanges[0] != 2 || got.answers[0] != 5 {
		t.Errorf("%d and %d changes of rank over %d answers, want 2 and 2 over 5", got.rankChanges[0], got.topicChanges[0], got.answers[0])
	}
	if got := r.screen.moves[0]; !slices.Equal(got, []uint16{4, 3, 3, 5, 3}) {
		t.Errorf("the card's moves %v, want 4, 3, 3, 5, 3", got)
	}
	if got, has := rankChanges(0, false)(r); !has || got != 40 {
		t.Errorf("rank changes %v a hundred answers (read %v), want 40", got, has)
	}
	if _, has := rankChanges(1, true)(r); has {
		t.Error("rank changes in a window with no answer were read, want none")
	}
}
