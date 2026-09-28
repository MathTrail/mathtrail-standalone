// Command load drives the MathTrail service with a scenario of a load run, and
// writes what the run came to as Markdown: every kind of answer the service
// gave and how long each took, what one unit of the run's work cost as the
// platform bills it against its free tier, and the facts that fail the run.
//
// Usage:
//
//	load -scenario lesson -url http://localhost:8080 [flags]
//
// The children of a run sign in through the development sign-in, each by a
// name of its own, so the service has to run with it, as `just run` runs it.
//
// It exits 0 when the run found nothing the service should never do, 1 when
// it found something, and 2 when it could not run, or was stopped before its
// end — the report of the part that ran is written all the same.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/MathTrail/mathtrail-standalone/tools/load/report"
	"github.com/MathTrail/mathtrail-standalone/tools/load/scenario"
	"github.com/MathTrail/mathtrail-standalone/tools/load/session"
)

// The ways the command ends.
const (
	exitClean  = 0
	exitHard   = 1
	exitCannot = 2
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// run is the command, from its arguments to its exit code: the report goes to
// stdout, and what went wrong before there was one to stderr.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("load", flag.ContinueOnError)
	flags.SetOutput(stderr)
	name := flags.String("scenario", "", "the scenario to run: "+strings.Join(scenario.Names(), ", "))
	address := flags.String("url", "", "where the service is reached, such as http://localhost:8080")
	host := flags.String("host", "", "the host the service knows itself by, when the URL names it otherwise")
	tasks := flags.Int("tasks", 0, "how many tasks a lesson asks for, when not the scenario's own")
	pace := flags.Duration("pace", 0, "the pause before every step of a lesson, when not the scenario's own")
	timeout := flags.Duration("timeout", 0, "how long one call may take, when not the scenario's own")
	cpus := flags.Float64("cpus", 1, "the vCPUs an instance is billed for")
	memory := flags.String("memory", "512m", "the memory an instance is billed for, as docker writes it: 512m, 1g")
	if err := flags.Parse(args); err != nil {
		return exitCannot
	}

	options, err := scenario.Defaults(*name)
	if err != nil {
		return cannot(stderr, err)
	}
	// Only what was asked for changes: every other option is the scenario's.
	flags.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "tasks":
			options.Tasks = *tasks
		case "pace":
			options.Pace = *pace
		case "timeout":
			options.Timeout = *timeout
		}
	})
	if *address == "" {
		return cannot(stderr, errors.New("-url names no service: where is it reached?"))
	}
	if !(*cpus > 0) || math.IsInf(*cpus, 0) {
		return cannot(stderr, fmt.Errorf("-cpus %v is no share of a processor an instance could have", *cpus))
	}
	gib, err := gibibytes(*memory)
	if err != nil {
		return cannot(stderr, err)
	}
	if err = options.Check(); err != nil {
		return cannot(stderr, err)
	}

	fmt.Fprintf(stderr, "load: %s against %s, about %s at a pause of %s\n",
		options.Scenario, *address, time.Duration(options.Steps())*options.Pace, options.Pace)
	result, err := scenario.Run(ctx, options, session.Target{URL: *address, Host: *host})
	if err != nil {
		return cannot(stderr, err)
	}
	hards, err := report.Write(stdout, []report.Run{result}, report.Instance{VCPU: *cpus, GiB: gib})
	switch {
	case err != nil:
		return cannot(stderr, err)
	case result.Stopped:
		// What ran is reported, but a run cut short proves nothing either way.
		return cannot(stderr, errors.New("the run was stopped before its end; the report is of the part that ran"))
	case len(hards) > 0:
		return exitHard
	}
	return exitClean
}

// cannot says why the command could not run, and is the exit code that says
// so.
func cannot(stderr io.Writer, err error) int {
	fmt.Fprintf(stderr, "load: %v\n", err)
	return exitCannot
}

// gibibytes reads a size as docker writes one — a number and one of the units
// b, k, m and g, each 1024 times the one before — in GiB.
func gibibytes(size string) (float64, error) {
	units := map[byte]float64{'b': 1 << 30, 'k': 1 << 20, 'm': 1 << 10, 'g': 1}
	if size == "" {
		return 0, errors.New("-memory names no size")
	}
	unit := size[len(size)-1]
	if 'A' <= unit && unit <= 'Z' {
		unit += 'a' - 'A'
	}
	per, known := units[unit]
	if !known {
		return 0, fmt.Errorf("-memory %q has no unit of b, k, m or g", size)
	}
	amount, err := strconv.ParseFloat(size[:len(size)-1], 64)
	// Written the other way round, a size that is not a number would pass.
	if err != nil || !(amount > 0) || math.IsInf(amount, 0) {
		return 0, fmt.Errorf("-memory %q is not a size", size)
	}
	return amount / per, nil
}
