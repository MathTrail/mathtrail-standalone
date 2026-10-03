// Command learnersim runs simulated children through the service's own rule,
// profile and rating code, under the service's estimate of their level and
// under baselines, and measures how each estimate places a child, keeps their
// tasks in the corridor, declares mastery and follows a child who changes.
//
// Every child is drawn from a seed of its own and is run under every rule with
// the same parameters and the same draws for its answers. The results are
// written as tables, a file of numbers for the paper, the chain of mastery
// computed over its grid, and the checks of the baselines against their
// papers.
//
// Usage:
//
//	learnersim [-out <directory>] [-children <n>] [-answers <n>]
package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

func main() {
	os.Exit(runCommand(os.Args[1:], os.Stdout, os.Stderr))
}

func runCommand(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("learnersim", flag.ContinueOnError)
	flags.SetOutput(stderr)
	out := flags.String("out", "experiments/learnersim/results", "the directory the results are written to")
	children := flags.Int("children", 1000, "how many children each generator draws")
	answers := flags.Int("answers", 200, "how many answers each child gives")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if err := experimentInto(*out, *children, *answers, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func experimentInto(out string, children, answers int, stdout io.Writer) error {
	if out == "" || children < 1 || answers < 1 {
		return errors.New("usage: learnersim [-out <directory>] [-children <n>] [-answers <n>]")
	}
	w, err := newWorld(answers)
	if err != nil {
		return err
	}
	all := cells()
	ms := metrics()
	results, err := runCells(w, all, children, ms)
	if err != nil {
		return err
	}
	if err := writeAll(out, all, results, ms, designNumbers(children, answers)); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%d cells of %d children, %d answers each, written to %s\n", len(all), children, answers, out)
	return nil
}

// newWorld is what every run shares: the shipped catalog and its topics, and
// a sealer with a key made for the run, which seals the tasks as the service
// does.
func newWorld(answers int) (*world, error) {
	shipped, err := content.Load()
	if err != nil {
		return nil, fmt.Errorf("learnersim: load the content: %w", err)
	}
	key := make([]byte, seal.KeySize)
	if _, failed := rand.Read(key); failed != nil {
		return nil, fmt.Errorf("learnersim: make a sealing key: %w", failed)
	}
	ring, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(key), "")
	if err != nil {
		return nil, fmt.Errorf("learnersim: make the key ring: %w", err)
	}
	return &world{catalog: shipped, topics: shipped.TopicIDs(), sealer: ring.For(seal.PurposeTaskAnswer), answers: answers}, nil
}

// job is one child of one cell.
type job struct {
	cell, child int
}

// runCells runs every child of every cell, as many at once as there are
// processors, and keeps of each what its summaries need.
func runCells(w *world, all []cell, children int, ms []metric) ([][]vector, error) {
	results := make([][]vector, len(all))
	for i := range results {
		results[i] = make([]vector, children)
	}
	jobs := make(chan job)
	var (
		wg    sync.WaitGroup
		once  sync.Once
		first error
	)
	for range runtime.GOMAXPROCS(0) {
		wg.Go(func() {
			for j := range jobs {
				c := &all[j.cell]
				r, err := run(w, c.rule, newChild(c.generator, j.child, w.topics))
				if err != nil {
					once.Do(func() { first = err })
					continue
				}
				results[j.cell][j.child] = vectorOf(r, ms)
			}
		})
	}
	for c := range all {
		for i := range children {
			jobs <- job{cell: c, child: i}
		}
	}
	close(jobs)
	wg.Wait()
	return results, first
}
