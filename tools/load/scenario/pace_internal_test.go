package scenario

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// answeredAfter is a call that is answered after the time given, and counts
// how many of its kind are in flight at the most.
func answeredAfter(took time.Duration, flying, most *atomic.Int32) func(context.Context, uint64) session.Answer {
	return func(context.Context, uint64) session.Answer {
		now := flying.Add(1)
		for {
			seen := most.Load()
			if now <= seen || most.CompareAndSwap(seen, now) {
				break
			}
		}
		started := time.Now()
		time.Sleep(took)
		flying.Add(-1)
		return session.Answer{Kind: session.Answered, Started: started, Ended: time.Now()}
	}
}

// The runner keeps its pace however slowly the calls are answered: a call goes
// out when it is due, with the ones before it still in flight. The bounds are
// wide, since a shared machine may run the runner late, and what is held is
// what a runner that waited for answers could not do.
func TestCallsGoOutAtThePaceWhateverTheAnswersTake(t *testing.T) {
	t.Parallel()

	var flying, most atomic.Int32
	hits, unsent := attack(t.Context(), vegeta.ConstantPacer{Freq: 50, Per: time.Second}, time.Second, 1000,
		answeredAfter(400*time.Millisecond, &flying, &most))

	if len(hits) < 30 || len(hits) > 55 || unsent != 0 {
		t.Errorf("%d calls sent and %d not, want about 50 sent — fifty a second for a second — and none held back",
			len(hits), unsent)
	}
	if got := most.Load(); got < 10 {
		t.Errorf("at most %d calls were in flight at once, want the pace kept while the answers took 400ms", got)
	}
}

// A call due while the runner has as many in flight as it may is counted as
// not sent, and the pace goes on: nothing waits for room to come free.
func TestACallPastTheRoomIsCountedNotDelayed(t *testing.T) {
	t.Parallel()

	var flying, most atomic.Int32
	started := time.Now()
	hits, unsent := attack(t.Context(), vegeta.ConstantPacer{Freq: 50, Per: time.Second}, 300*time.Millisecond, 2,
		answeredAfter(time.Second, &flying, &most))
	took := time.Since(started)

	if len(hits) != 2 || unsent < 5 {
		t.Errorf("%d calls sent and %d not, want the two there was room for sent and the rest counted", len(hits), unsent)
	}
	if took > 3*time.Second {
		t.Errorf("the run took %v, want it over once the two calls in flight were answered", took)
	}
}

// A call never goes out before it is due, and counts its time from the
// moment it was due: a call the runner sent late is not reported the faster
// for it.
func TestACallCountsFromTheMomentItWasDue(t *testing.T) {
	t.Parallel()

	var flying, most atomic.Int32
	hits, _ := attack(t.Context(), vegeta.ConstantPacer{Freq: 20, Per: time.Second}, 500*time.Millisecond, 100,
		answeredAfter(time.Millisecond, &flying, &most))
	if len(hits) < 5 {
		t.Fatalf("%d calls sent, want about ten: twenty a second for half a second", len(hits))
	}
	for _, h := range hits {
		if h.answer.Started.Before(h.due) {
			t.Errorf("a call went out at %v, before it was due at %v", h.answer.Started, h.due)
		}
		if call := report.CallOf(&h.answer, h.due); call.Took() < h.answer.Took() {
			t.Errorf("a call counts %v, want at least the %v it took from when it went out", call.Took(), h.answer.Took())
		}
	}
}

// A run asked to stop stops: the runner sends nothing after it is told to.
func TestARunAskedToStopStops(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()
	var flying, most atomic.Int32
	started := time.Now()
	hits, _ := attack(ctx, vegeta.ConstantPacer{Freq: 50, Per: time.Second}, time.Minute, 1000,
		answeredAfter(time.Millisecond, &flying, &most))
	if took := time.Since(started); took > 2*time.Second || len(hits) > 10 {
		t.Errorf("the run sent %d calls over %v, want it stopped after a tenth of a second", len(hits), took)
	}
}
