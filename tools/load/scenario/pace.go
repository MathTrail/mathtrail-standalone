package scenario

import (
	"context"
	"sync"
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"

	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

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

// pauseUntil waits for the moment given, and says whether the run may go on.
func pauseUntil(ctx context.Context, moment time.Time) bool {
	return pause(ctx, time.Until(moment))
}
