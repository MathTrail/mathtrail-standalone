package main

import (
	"math"
	"testing"
)

// answerIn is a child with one answer, in a calibration bin.
func answerIn(b int, predicted, observed float64) vector {
	var v vector
	v.bins[b] = bin{predicted: predicted, observed: observed, n: 1}
	return v
}

// The calibration error is read off a sample's answers pooled bin by bin, each
// bin weighed by its answers: two children whose misses in one bin go opposite
// ways are calibrated together, misses in two bins add up, and a sample with
// no answers gives no number.
func TestTheCalibrationErrorPoolsTheAnswersOfEachBin(t *testing.T) {
	t.Parallel()
	opposite := []vector{answerIn(5, 0.5, 1), answerIn(5, 0.5, 0)}
	apart := []vector{answerIn(5, 0.5, 1), answerIn(7, 0.7, 0)}
	cases := []struct {
		name     string
		children []vector
		sample   []int
		want     float64
		read     bool
	}{
		{"opposite misses in one bin", opposite, []int{0, 1}, 0, true},
		{"one of them drawn twice", opposite, []int{0, 0}, 0.5, true},
		{"misses in two bins", apart, []int{0, 1}, (0.5 + 0.7) / 2, true},
		{"no answers", []vector{{}, {}}, []int{0, 1}, 0, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, read := calibrationError(tc.children, tc.sample)
			if read != tc.read || math.Abs(got-tc.want) > 1e-12 {
				t.Errorf("calibrationError() = %v (read %v), want %v (read %v)", got, read, tc.want, tc.read)
			}
		})
	}
}

// offBy is an estimate that stands a set distance from the child in each
// topic.
type offBy struct {
	c  *child
	by map[string]float64
}

func (o offBy) overall() float64               { return o.c.theta }
func (o offBy) level(topic string) float64     { return o.c.level(topic) + o.by[topic] }
func (o offBy) chance(string, float64) float64 { return 0.5 }
func (o offBy) answered(string, float64, bool) {}

// offUntil is an error of one in every topic before answer k, and none after.
func offUntil(k int) func(int, string) float64 {
	return func(answer int, _ string) float64 {
		if answer < k {
			return 1
		}
		return 0
	}
}

// From a child's jump on, the answers the rule took to follow it are counted
// until its error, averaged with its signs over the topics answered since the
// jump, stays within the bound for a run of answers, which is not counted
// itself; a child the rule never settled counts every answer, and is counted
// apart, and a child with no jump counts in neither.
func TestFollowingAJumpCountsTheAnswersUntilTheErrorSettles(t *testing.T) {
	t.Parallel()
	opposite := map[string]float64{"a": 0.6, "b": -0.6}
	cases := []struct {
		name      string
		jumpAt    int
		errorAt   func(k int, topic string) float64
		answers   int
		followed  float64
		read      bool
		unsettled [2]float64
	}{
		{"no jump", -1, offUntil(0), 20, 0, false, [2]float64{0, 0}},
		{"followed from the jump", 0, offUntil(0), 20, 0, true, [2]float64{0, 1}},
		{"followed after five answers", 0, offUntil(5), 25, 5, true, [2]float64{0, 1}},
		{"counted from the jump on", 3, offUntil(8), 25, 5, true, [2]float64{0, 1}},
		{"signs cancel across topics", 0, func(_ int, topic string) float64 { return opposite[topic] }, 15, 1, true, [2]float64{0, 1}},
		{"never settled", 0, offUntil(30), 30, 30, true, [2]float64{1, 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			followed, read, unsettled := jumpFollowed(tc.jumpAt, tc.errorAt, tc.answers)
			if followed != tc.followed || read != tc.read || unsettled != tc.unsettled {
				t.Errorf("followed in %v answers (read %v), unsettled %v of %v; want %v (read %v), %v of %v",
					followed, read, unsettled[0], unsettled[1], tc.followed, tc.read, tc.unsettled[0], tc.unsettled[1])
			}
		})
	}
}

// jumpFollowed is what the measures of a jump make of a child of two topics
// who jumps at answer jumpAt, under an estimate that stands errorAt from the
// child at every answer, over the answers given: how many answers the
// estimate took to follow, whether that was read, and the share left
// unsettled.
func jumpFollowed(jumpAt int, errorAt func(k int, topic string) float64, answers int) (followed float64, read bool, unsettled [2]float64) {
	topics := []string{"a", "b"}
	c := &child{theta: 1, delta: map[string]float64{"a": 0.25, "b": -0.25}, jumpAt: jumpAt}
	est := offBy{c: c, by: map[string]float64{}}
	s := &session{c: c, est: est}
	r := newChildResult(c)
	for k := range answers {
		for _, topic := range topics {
			est.by[topic] = errorAt(k, topic)
		}
		r.followJump(s, topics[k%len(topics)], k)
	}
	followed, read = jumpAnswers(r)
	num, den := jumpUnsettled(r)
	return followed, read, [2]float64{num, den}
}

// An interval of no draws has no ends, and a number with such an interval is
// not drawn in a figure, where it would stand as a bar at zero.
func TestAnIntervalOfNoDrawsHasNoEnds(t *testing.T) {
	t.Parallel()
	low, high := interval(nil)
	if !math.IsNaN(low) || !math.IsNaN(high) {
		t.Fatalf("interval(nil) = %v, %v, want no ends", low, high)
	}
	if drawable(summary{value: 0.5, low: low, high: high, has: true}) {
		t.Error("a number with no ends to its interval is drawable, want it refused")
	}
}

// Past the metrics, an index names a pooled number, in the order the names of
// a run list them, so that a summary or a comparison of a pooled number reads
// that number and not its neighbour's.
func TestAnIndexPastTheMetricsReadsThePooledNumberItNames(t *testing.T) {
	t.Parallel()
	ms := metrics()
	names := metricNames(ms)
	var v vector
	v.bins[3] = bin{predicted: 0.3, observed: 1, n: 1}
	for j := range longestChain {
		v.attempts[j] = atAttempt{atRisk: 10}
	}
	// One declaration at each length a pooled number reads within, so that
	// each reads a chance of its own.
	for _, m := range chainLengths {
		v.attempts[m-1].declared = 1
	}
	children := []vector{v}
	seen := map[float64]string{}
	for k, p := range pooledMetrics() {
		at := len(ms) + k
		if names[at] != p.name {
			t.Fatalf("name %d is %s, want the pooled %s", at, names[at], p.name)
		}
		got, gotRead := readerOf(at, ms)(children, []int{0})
		want, wantRead := p.read(children, []int{0})
		if got != want || gotRead != wantRead {
			t.Errorf("readerOf(%d) = %v (read %v), want %s's %v (read %v)", at, got, gotRead, p.name, want, wantRead)
		}
		if other, taken := seen[want]; taken {
			t.Fatalf("%s and %s read the same %v here, so a mix-up would not show", other, p.name, want)
		}
		seen[want] = p.name
	}
}
