// Command perf measures what the service's review of a task costs on this
// machine: how long the two halves of a review take on every reference task
// the checks accept as it is, how many steps and how long each run of its
// solver takes, what a review allocates, and how large the package the model
// is handed is at every point of the ladder a topic is taught at.
//
// The reviews are the experiment with injected defects' own: the same
// reference tasks, made into submissions the same way, reviewed by the
// service's reviewer in the service's sandbox. They run one at a time, so that
// no review waits for another's solver, in passes that each review every task
// once, so that a drift of the machine reaches every task alike.
//
// Usage:
//
//	perf [-out <directory>] [-passes <reviews of each task>]
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

// passesOverEveryTask is how many times every task is reviewed for its
// timings.
const passesOverEveryTask = 10

func main() {
	testing.Init()
	if err := flag.Set("test.benchtime", benchmarkPasses); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("perf", flag.ContinueOnError)
	flags.SetOutput(stderr)
	out := flags.String("out", "experiments/perf/results", "the directory the results are written to")
	passes := flags.Int("passes", passesOverEveryTask, "how many times every task is reviewed for its timings")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if err := measureInto(context.Background(), *out, *passes, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func measureInto(ctx context.Context, out string, passes int, stdout io.Writer) error {
	if out == "" || passes < 1 {
		return errors.New("usage: perf [-out <directory>] [-passes <reviews of each task, at least 1>]")
	}
	shipped, err := content.Load()
	if err != nil {
		return fmt.Errorf("perf: load the content: %w", err)
	}
	runner, err := reviewing.NewRunner()
	if err != nil {
		return err
	}
	tasks, err := accepted(ctx, shipped, runner)
	if err != nil {
		return err
	}
	timings, solverRuns, err := timeReviews(ctx, tasks, passes)
	if err != nil {
		return err
	}
	cost, err := reviewCost(ctx, tasks)
	if err != nil {
		return err
	}
	packs, err := packages(shipped)
	if err != nil {
		return err
	}
	found := measurements{tasks: len(tasks), passes: passes, timings: timings, solverRuns: solverRuns, cost: cost, packages: packs}
	if err := writeAll(out, &found); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%d tasks reviewed %d times each, %d packages, written to %s\n", len(tasks), passes, len(packs), out)
	return nil
}
