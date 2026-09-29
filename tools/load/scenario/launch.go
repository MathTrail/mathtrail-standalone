package scenario

import (
	"context"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/instance"
	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// Launched is a service brought up for a run.
type Launched struct {
	// Target is where the service is reached.
	Target session.Target
	// Asked is when the service was asked to come up, Began when its process
	// began and Healthy when it first answered its probe: zero for a service
	// that never did.
	Asked, Began, Healthy time.Time
	// Spent is what its instance spent between two moments, Peak the most
	// memory the instance has held so far, Alive whether the instance still
	// runs, and Stop ends the instance and is its life. Each is nil for a
	// service the run did not start, which is neither the run's to measure,
	// to watch nor to end.
	Spent func(from, to time.Time) (report.Spent, bool)
	Peak  func() (uint64, bool)
	Alive func() bool
	Stop  func(ctx context.Context) (report.Life, error)
}

// Launch brings a service up for a run. The error is for a service that
// could not be brought up at all: one that came up and never answered its
// probe is launched all the same, and what became of it is the run's to tell.
type Launch func(ctx context.Context) (Launched, error)

// At is the service at the target: somebody else's, already up, and the same
// service however many times it is launched.
func At(target session.Target) Launch {
	return func(context.Context) (Launched, error) {
		now := time.Now()
		return Launched{Target: target, Asked: now, Began: now, Healthy: now}, nil
	}
}

// FromImage is a container of the spec of its own, every time it is launched.
func FromImage(spec *instance.Spec) Launch {
	return func(ctx context.Context) (Launched, error) {
		container, err := instance.Start(ctx, spec)
		if err != nil {
			return Launched{}, err
		}
		return Launched{
			Target: container.Target, Asked: container.Asked, Began: container.Began, Healthy: container.Healthy,
			Spent: container.Spent, Peak: container.Peak, Alive: container.Alive, Stop: container.Stop,
		}, nil
	}
}

// served runs what is given against a service the launch brings up, and ends
// the service after, when the run started it: the runs it came to, each with
// what the instance spent while it went on, and the last with the life of the
// instance. A service that never answered its probe is a run of its own, and
// a failed one.
func served(ctx context.Context, launch Launch, name string, what func(up *Launched) []report.Run) ([]report.Run, error) {
	up, err := launch(ctx)
	if err != nil {
		return nil, err
	}
	var runs []report.Run
	if up.Healthy.IsZero() {
		runs = []report.Run{neverUp(ctx, name, &up)}
	} else {
		runs = what(&up)
	}
	return runs, up.end(ctx, runs)
}

// ended says whether the service is known to have ended: a run that cannot
// tell takes it as running.
func (l *Launched) ended() bool { return l.Alive != nil && !l.Alive() }

// neverUp is the run of a service that never answered its probe.
func neverUp(ctx context.Context, name string, up *Launched) report.Run {
	run := report.Run{Scenario: name + ": start", Began: up.Asked, Broken: []string{"the service never answered its probe"}}
	finish(ctx, &run)
	return run
}

// end ends the service when the run started it, and gives each of the runs
// what the instance spent while it went on, and the last of them the life of
// the instance. The cgroup counts the instance as a whole, so a run that went
// on while another of the same instance did is given no spending: what the
// instance spent then was theirs together.
func (l *Launched) end(ctx context.Context, runs []report.Run) error {
	if l.Stop == nil {
		return nil
	}
	for i := range runs {
		if l.Spent == nil || overlapped(runs, i) {
			continue
		}
		if spent, measured := l.Spent(runs[i].Began, runs[i].Ended); measured {
			runs[i].Spent = &spent
		}
	}
	life, err := l.Stop(ctx)
	if err != nil {
		return err
	}
	if len(runs) > 0 {
		last := &runs[len(runs)-1]
		last.Lives = append(last.Lives, life)
	}
	return nil
}

// overlapped says whether another of the runs went on while the i-th did.
func overlapped(runs []report.Run, i int) bool {
	for j := range runs {
		if j != i && runs[j].Began.Before(runs[i].Ended) && runs[i].Began.Before(runs[j].Ended) {
			return true
		}
	}
	return false
}
