package scenario

import (
	"context"
	"math"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// inFlight is the most calls an attack keeps in flight at once. Past it, the
// service is answering too slowly for more calls to say anything new, and the
// tool would only be spending itself.
const inFlight = 256

// hit is one call the runner sent: when it was due to go out, and what came
// back.
type hit struct {
	due    time.Time
	answer session.Answer
}

// attack sends calls at the pace the pacer sets, for as long as given, and is
// what came back of each call it sent and how many it could not send.
//
// Every call goes out in a goroutine of its own when it is due, whether or not
// the ones before it were answered, so that a service that answers slowly gets
// more calls in flight rather than fewer: a runner that waited for answers
// would slow down with the service and report the service faster than it is.
// A call due while the runner already has as many in flight as it may is
// counted as not sent rather than held back until there is room, and the
// pacer is told of it all the same, so that it never bursts to catch up. The
// calls still in flight when the time is up are waited for; each ends by its
// own deadline.
func attack(ctx context.Context, pacer vegeta.Pacer, over time.Duration, inFlight int,
	call func(ctx context.Context, n uint64) session.Answer,
) (hits []hit, unsent int) {
	var (
		mu     sync.Mutex
		flying sync.WaitGroup
	)
	room := make(chan struct{}, inFlight)
	began := time.Now()
	for scheduled := uint64(0); ; scheduled++ {
		elapsed := time.Since(began)
		wait, stop := pacer.Pace(elapsed, scheduled)
		if stop || elapsed+wait >= over {
			break
		}
		due := began.Add(elapsed + wait)
		if !pauseUntil(ctx, due) {
			break
		}

		select {
		case room <- struct{}{}:
		default:
			unsent++
			continue
		}
		flying.Go(func() {
			defer func() { <-room }()
			answer := call(ctx, scheduled)
			mu.Lock()
			hits = append(hits, hit{due: due, answer: answer})
			mu.Unlock()
		})
	}
	flying.Wait()
	return hits, unsent
}

// callsOf are the hits of an attack as the report counts them, each from the
// moment it was due.
func callsOf(hits []hit) []report.Call {
	calls := make([]report.Call, len(hits))
	for i := range hits {
		calls[i] = report.CallOf(&hits[i].answer, hits[i].due)
	}
	return calls
}

// every is the pace of calls sent at the rate given, in calls a second.
func every(rate float64) vegeta.Pacer {
	return steady(time.Duration(math.Round(float64(time.Second) / rate)))
}

// steady is the pace of one call every interval given.
func steady(interval time.Duration) vegeta.Pacer {
	return vegeta.ConstantPacer{Freq: 1, Per: interval}
}

// pauseUntil waits for the moment given, and says whether the run may go on.
func pauseUntil(ctx context.Context, moment time.Time) bool {
	return pause(ctx, time.Until(moment))
}
