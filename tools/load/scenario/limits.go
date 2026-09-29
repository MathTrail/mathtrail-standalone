package scenario

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// document is the path a greedy address and the quiet ones fetch: the
// authorization server's metadata, which anybody may ask for before signing
// in, and which is held to the pace of the address that asks.
const document = "/.well-known/oauth-authorization-server"

// runLimits runs four things at once for as long as the options say — a
// greedy child and a greedy address past their paces, and other children
// walking their lessons and quiet addresses fetching the document within
// theirs — and then looks at whether the service came back. Each of the four
// is a run of its own, held to what its pace should have done to it: the
// greedy ones held back, the others never. The children's calls carry no
// address: the endpoint they reach holds each account to its own pace, and
// only the sign-in and its documents hold an address to one.
//
// An address is what the tool writes into X-Forwarded-For, which the service
// reads as the address only when nothing stands in front of it: behind the
// platform, which writes the tool's own address after it, every address of a
// run is that one, and the quiet ones are held back with the greedy.
func runLimits(ctx context.Context, o *Options, launch Launch) ([]report.Run, error) {
	return served(ctx, launch, Limits, func(up *Launched) []report.Run {
		target := up.Target
		var group sync.WaitGroup
		var runs [4]report.Run
		group.Go(func() { runs[0] = greedyChild(ctx, o, target) })
		group.Go(func() { runs[1] = otherChildren(ctx, o, target) })
		group.Go(func() { runs[2] = greedyAddress(ctx, o, target) })
		group.Go(func() { runs[3] = quietAddresses(ctx, o, target) })
		group.Wait()
		return recovered(ctx, o, up, Limits, runs[:]...)
	})
}

// greedyChild reads its profile at the rate of the options, past the pace of
// any account.
func greedyChild(ctx context.Context, o *Options, target session.Target) report.Run {
	service := session.Open(target, o.Timeout)
	defer service.Close()
	child := service.Child(Limits + "-greedy-" + stamp())
	defer func() { _ = child.Close() }()

	run := report.Run{Scenario: Limits + ": the greedy child", Unit: "call", Began: time.Now()}
	hits, unsent := attack(ctx, every(o.Rate), o.Duration, inFlight, func(ctx context.Context, _ uint64) session.Answer {
		return child.Call(ctx, "get_profile", map[string]any{})
	})
	finish(ctx, &run)
	run.Calls, run.Unsent = callsOf(hits), unsent
	run.Requests, run.Reconnects = service.Requests(), child.Reconnects()
	run.Units = len(run.Calls)
	run.Broken = heldBack(&run, "calls of the greedy child")
	return run
}

// otherChildren walk their lessons at the pace of the options, each a child
// of its own, all at once: none of them should ever meet a pace.
func otherChildren(ctx context.Context, o *Options, target session.Target) report.Run {
	service := session.Open(target, o.Timeout)
	defer service.Close()

	lessons := make([]report.Run, o.Children)
	began := stamp()
	var group sync.WaitGroup
	for i := range lessons {
		group.Go(func() { lessons[i] = walkLesson(ctx, o, service, fmt.Sprintf("%s-%s-%d", Limits, began, i)) })
	}
	group.Wait()

	run := report.Run{Scenario: Limits + ": the other children", Unit: "accepted task"}
	for i := range lessons {
		run.Began = earliest(run.Began, lessons[i].Began)
		run.Calls = append(run.Calls, lessons[i].Calls...)
		run.Units += lessons[i].Units
		run.Reconnects += lessons[i].Reconnects
	}
	finish(ctx, &run)
	run.Requests = service.Requests()
	run.Broken = lessonExpected(&run, o.Tasks*o.Children)
	return run
}

// greedyAddress fetches the document at the rate of the options, past the
// pace of any address.
func greedyAddress(ctx context.Context, o *Options, target session.Target) report.Run {
	service := session.Open(target, o.Timeout)
	defer service.Close()

	run := report.Run{Scenario: Limits + ": the greedy address", Unit: "fetch", Began: time.Now()}
	hits, unsent := attack(ctx, every(o.Rate), o.Duration, inFlight, func(ctx context.Context, _ uint64) session.Answer {
		return service.Fetch(ctx, document, address(0))
	})
	finish(ctx, &run)
	run.Calls, run.Unsent = callsOf(hits), unsent
	run.Requests = service.Requests()
	run.Units = len(run.Calls)
	run.Broken = heldBack(&run, "fetches of the greedy address")
	return run
}

// quietAddresses fetch the document as often as the options say, each from an
// address of its own, as long as the greedy ones go on: none of them should
// ever meet a pace.
func quietAddresses(ctx context.Context, o *Options, target session.Target) report.Run {
	service := session.Open(target, o.Timeout)
	defer service.Close()

	run := report.Run{Scenario: Limits + ": the quiet addresses", Unit: "fetch", Began: time.Now()}
	var (
		mu    sync.Mutex
		group sync.WaitGroup
	)
	for i := 1; i <= o.Children; i++ {
		group.Go(func() {
			hits, unsent := attack(ctx, steady(o.QuietEvery), o.Duration, inFlight, func(ctx context.Context, _ uint64) session.Answer {
				return service.Fetch(ctx, document, address(i))
			})
			mu.Lock()
			defer mu.Unlock()
			run.Calls = append(run.Calls, callsOf(hits)...)
			run.Unsent += unsent
		})
	}
	group.Wait()
	finish(ctx, &run)
	run.Requests = service.Requests()
	run.Units = len(run.Calls)
	if other := kindsBut(&run, session.Answered); other != "" {
		run.Broken = []string{"fetches of the quiet addresses answered otherwise than served: " + other}
	}
	return run
}

// heldBack is what a greedy one's run did not come to that it should have:
// held back by its pace at least once, and answered as asked whenever it was
// not.
func heldBack(run *report.Run, whose string) []string {
	var broken []string
	if countOf(run, session.Paced) == 0 {
		broken = append(broken, whose+" were never held back by a pace")
	}
	if other := kindsBut(run, session.Answered, session.Paced); other != "" {
		broken = append(broken, whose+" answered otherwise than as asked or held back: "+other)
	}
	return broken
}

// address is the address the i-th of a run's addresses fetches from, the
// greedy one first: one of the block set aside for benchmarking networks,
// which is nobody's on the internet.
func address(i int) string {
	n := i + 1
	return fmt.Sprintf("198.18.%d.%d", n/256, n%256)
}

// earliest is the earlier of two moments, a zero moment being neither.
func earliest(a, b time.Time) time.Time {
	if a.IsZero() || (!b.IsZero() && b.Before(a)) {
		return b
	}
	return a
}
