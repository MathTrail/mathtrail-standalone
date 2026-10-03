// Command faultinject injects one defect at a time into the service's
// reference tasks and records which of the service's checks refuse each case.
//
// Every reference task is first made into a submission and reviewed as it is;
// only those the checks accept with every check run carry a defect, so that a
// refusal is the defect's doing. Each operator then applies to every such task
// it fits, the case is reviewed by the service's own reviewer, and the results
// are written as tables, a file of numbers for the paper and a sample of the
// out-of-scope cases for a person to read.
//
// Usage:
//
//	faultinject [-out <directory>]
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/research/experiments/reviewing"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("faultinject", flag.ContinueOnError)
	flags.SetOutput(stderr)
	out := flags.String("out", "experiments/faultinject/results", "the directory the results are written to")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if err := experimentInto(context.Background(), *out, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func experimentInto(ctx context.Context, out string, stdout io.Writer) error {
	if out == "" {
		return errors.New("usage: faultinject [-out <directory>]")
	}
	shipped, err := content.Load()
	if err != nil {
		return fmt.Errorf("faultinject: load the content: %w", err)
	}
	runner, err := reviewing.NewRunner()
	if err != nil {
		return err
	}
	ops := operators()
	found, err := runExperiment(ctx, shipped, runner, ops)
	if err != nil {
		return err
	}
	if err := writeAll(out, &found, ops); err != nil {
		return err
	}
	eligible := 0
	for i := range found.Sanity {
		if found.Sanity[i].eligible() {
			eligible++
		}
	}
	fmt.Fprintf(stdout, "%d reference tasks, %d accepted as they are; %d cases of %d operators, written to %s\n",
		len(found.Sanity), eligible, len(found.Cases), len(ops), out)
	return nil
}
