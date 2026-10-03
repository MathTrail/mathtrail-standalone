// Package scenario is what a run does to the service. A lesson walks one
// child through the tools at a live pace. The attacks — saturation, limits
// and the adversarial solvers — send calls at a constant pace however slowly
// they are answered, and each is followed by a look at whether the service
// came back.
package scenario

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/MathTrail/mathtrail-standalone/internal/config"
	"github.com/MathTrail/mathtrail-standalone/tools/load/lesson"
	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
)

// Options are what a run of a scenario is asked to do. A scenario reads the
// options it needs, which Reads names, and leaves the others be.
type Options struct {
	// Scenario is the name of the scenario.
	Scenario string
	// Tasks is how many tasks a lesson asks for.
	Tasks int
	// Pace is the pause before every step of a lesson: a model writing a task,
	// a child reading one.
	Pace time.Duration
	// Timeout is how long one call may take before it counts as unanswered,
	// and how long the service has to come back after an attack.
	Timeout time.Duration
	// Rate is how many calls a second an attack sends: hand-ins in saturation
	// and adversarial; in limits, the calls of the greedy child and the
	// fetches of the greedy address, each.
	Rate float64
	// Duration is how long an attack goes on; in adversarial, each variant's.
	Duration time.Duration
	// Children is how many children an attack's hand-ins are shared among; in
	// limits, how many walk a lesson beside the greedy one, and how many
	// quiet addresses fetch the document beside the greedy address.
	Children int
	// QuietEvery is how often each quiet address of limits fetches the
	// document: well within the pace of an address.
	QuietEvery time.Duration
	// Variants are the costly solvers adversarial hands in, one after another.
	Variants []string
	// Steps is the step ceiling of the service, which the costly solvers are
	// sized to.
	Steps uint64
}

// The options a scenario may read, by the names of the flags that set them.
const (
	readsTasks    = "tasks"
	readsPace     = "pace"
	readsTimeout  = "timeout"
	readsRate     = "rate"
	readsDuration = "duration"
	readsChildren = "children"
	readsVariants = "variants"
	readsSteps    = "steps"
)

// The scenarios there are.
const (
	// Lesson is one child's lesson.
	Lesson = "lesson"
	// Saturation is more solvers than the sandbox has slots.
	Saturation = "saturation"
	// Limits is one child and one address past their paces, and the others
	// within theirs.
	Limits = "limits"
	// Adversarial is the costliest solvers the rules of the sandbox allow.
	Adversarial = "adversarial"
	// Cold is the service started again and again.
	Cold = "cold"
)

// The bounds of the options.
const (
	// maxRate is the most calls a second a run sends: a load of a tool, not a
	// flood, and a pace the runner keeps to the millisecond.
	maxRate = 1000
	// maxChildren is the most children a run signs in.
	maxChildren = 1000
)

// scenario is what a scenario starts from and which of the options it reads,
// what it does to the services it launches — which comes to one run, or to
// one for each thing it does at once or in turn — about how long that takes,
// what it refuses of the options besides what every scenario refuses, and
// whether it has to start every service it runs against itself.
type scenario struct {
	defaults func() Options
	reads    []string
	run      func(context.Context, *Options, Launch) ([]report.Run, error)
	lasts    func(*Options) time.Duration
	check    func(*Options) error
	starts   bool
}

// scenarios are the scenarios there are, by name.
var scenarios = map[string]scenario{
	Lesson: {
		// Five tasks at a pause of five seconds is a lesson of a minute and a
		// half: short enough to run by hand, and about the pace at which a
		// child's lesson calls the tools at its busiest.
		defaults: func() Options { return common(Lesson) },
		reads:    []string{readsTasks, readsPace, readsTimeout},
		run:      runLesson,
		lasts:    lessonLasts,
	},
	Saturation: {
		// Three hand-ins a second, about as many as the pace of an instance lets
		// through, of a solver that spends most of its budget, for half a
		// minute: whether the slots of an instance keep up with all the
		// sandbox can be sent.
		defaults: func() Options { return common(Saturation) },
		reads:    []string{readsRate, readsDuration, readsChildren, readsSteps, readsTimeout},
		run:      runSaturation,
		lasts:    func(o *Options) time.Duration { return o.Duration },
	},
	Limits: {
		// A child calling three times a second is three times the pace of an
		// account, and an address fetching as often nine times the pace of an
		// address; three other children walk lessons of two tasks, and three
		// quiet addresses fetch once every five seconds, well within theirs,
		// for the minute a pace is counted over.
		defaults: func() Options {
			o := common(Limits)
			o.Duration, o.Children, o.Tasks = time.Minute, 3, 2
			return o
		},
		reads: []string{readsRate, readsDuration, readsChildren, readsTasks, readsPace, readsTimeout},
		run:   runLimits,
		lasts: func(o *Options) time.Duration {
			return max(o.Duration, lessonLasts(o))
		},
		check: func(o *Options) error {
			if o.Duration <= o.QuietEvery || o.QuietEvery <= 0 {
				return fmt.Errorf("scenario: a quiet address fetches once every %v, which an attack of %v never comes to",
					o.QuietEvery, o.Duration)
			}
			return nil
		},
	},
	Adversarial: {
		// Every costly solver but the loop, whose share of the sandbox is
		// saturation's, for twenty seconds each.
		defaults: func() Options {
			o := common(Adversarial)
			o.Duration = 20 * time.Second
			return o
		},
		reads: []string{readsRate, readsDuration, readsChildren, readsVariants, readsSteps, readsTimeout},
		run:   runAdversarial,
		lasts: func(o *Options) time.Duration {
			return time.Duration(len(o.Variants)) * o.Duration
		},
	},
	Cold: {
		// A start is measured from the moment it is asked for, so every one
		// needs a service the run starts itself.
		defaults: func() Options { return common(Cold) },
		reads:    []string{readsTimeout},
		run:      runCold,
		lasts:    func(*Options) time.Duration { return starts * startLasts },
		starts:   true,
	},
}

// common are the options every scenario starts from.
func common(name string) Options {
	return Options{
		Scenario: name,
		Tasks:    5,
		Pace:     5 * time.Second,
		Timeout:  time.Minute,
		Rate:     3,
		Duration: 30 * time.Second,
		Children: 10,
		// Twelve fetches a minute, against the twenty an address may make.
		QuietEvery: 5 * time.Second,
		Variants:   []string{lesson.Appends, lesson.Helper, lesson.Tuples, lesson.Pairs, lesson.Sets, lesson.Product},
		Steps:      config.DefaultSolverSteps,
	}
}

// Names are the names of the scenarios there are, in order.
func Names() []string { return slices.Sorted(maps.Keys(scenarios)) }

// Reads are the options the scenario named reads, by the names of the flags
// that set them: an option it does not read is one it would leave be.
func Reads(name string) []string { return slices.Clone(scenarios[name].reads) }

// Starts says whether the scenario named starts every service it runs against
// itself: it cannot be run against a service somebody else started.
func Starts(name string) bool { return scenarios[name].starts }

// Defaults are the options of the scenario named, before any is changed.
func Defaults(name string) (Options, error) {
	known, found := scenarios[name]
	if !found {
		return Options{}, fmt.Errorf("scenario: no scenario %q; there are %s", name, strings.Join(Names(), ", "))
	}
	return known.defaults(), nil
}

// Check refuses options no run could keep to.
func (o *Options) Check() error {
	switch {
	case o.Tasks < 1 || o.Tasks > len(lesson.Written()):
		return fmt.Errorf("scenario: a lesson asks for 1 to %d tasks, not %d", len(lesson.Written()), o.Tasks)
	case o.Pace < 0:
		return fmt.Errorf("scenario: a pause of %v goes back in time", o.Pace)
	case o.Timeout <= 0:
		return fmt.Errorf("scenario: a call may take no time at all: %v", o.Timeout)
	case !(o.Rate > 0) || o.Rate > maxRate:
		return fmt.Errorf("scenario: a rate of %v calls a second is not between 0 and %d", o.Rate, maxRate)
	case o.Duration <= 0:
		return fmt.Errorf("scenario: an attack of %v is no attack", o.Duration)
	case !(o.Rate*o.Duration.Seconds() > 1):
		// The first call of an attack is due one interval in, so an attack no
		// longer than one interval sends nothing at all.
		return fmt.Errorf("scenario: an attack of %v at %v calls a second sends no call", o.Duration, o.Rate)
	case o.Children < 1 || o.Children > maxChildren:
		return fmt.Errorf("scenario: a run signs in 1 to %d children, not %d", maxChildren, o.Children)
	case o.Steps < lesson.MinSteps:
		return fmt.Errorf("scenario: the costly solvers are sized for %d steps at least, not %d", lesson.MinSteps, o.Steps)
	case len(o.Variants) == 0:
		return fmt.Errorf("scenario: no variant to hand in; there are %s", strings.Join(lesson.Costly(), ", "))
	}
	for _, variant := range o.Variants {
		if !slices.Contains(lesson.Costly(), variant) {
			return fmt.Errorf("scenario: no costly solver %q; there are %s", variant, strings.Join(lesson.Costly(), ", "))
		}
	}
	if known, found := scenarios[o.Scenario]; found && known.check != nil {
		return known.check(o)
	}
	return nil
}

// Lasts is about how long a run of the options takes, for whoever is waiting
// for it.
func (o *Options) Lasts() time.Duration {
	known, found := scenarios[o.Scenario]
	if !found {
		return 0
	}
	return known.lasts(o)
}

// Run runs the scenario of the options against the services the launch
// brings up, and is what it came to: one run, or one for each thing the
// scenario does at once or in turn. A run the scenario was asked to stop in
// the middle of is the part that ran, marked as stopped, and is held to no
// expectation: it did not get to meet them. A run that ended before is what
// it was. The error is for a scenario that could not go on: a service that
// could not be brought up, for one, with the runs that came before it.
func Run(ctx context.Context, o *Options, launch Launch) ([]report.Run, error) {
	known, found := scenarios[o.Scenario]
	if !found {
		return nil, fmt.Errorf("scenario: no scenario %q", o.Scenario)
	}
	if err := o.Check(); err != nil {
		return nil, err
	}

	runs, err := known.run(ctx, o, launch)
	for i := range runs {
		if runs[i].Stopped {
			runs[i].Broken = nil
		}
	}
	return runs, err
}

// finish ends a run: when it ended, and whether it was asked to stop before
// it could.
func finish(ctx context.Context, run *report.Run) {
	run.Ended = time.Now()
	run.Stopped = ctx.Err() != nil
}
