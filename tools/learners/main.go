// Command learners runs simulated children through the service's own rule,
// profile and rating code, under the service's estimate of their level and
// under the rules it is compared with, and measures how each estimate places a
// child, keeps their tasks in the corridor, declares mastery and follows a
// child who changes.
//
// Every child is drawn from a seed of its own and is run under every rule with
// the same parameters and the same draws for its answers. The results are
// written as tables of every cell and of the comparisons, a summary of the
// numbers a change to the rule is judged by, which is also printed, the
// criterion a new rule is chosen by as every rule meets it, and what the run
// was.
//
// Usage:
//
//	learners [-out <directory>] [-children <n>] [-answers <n>] [-seed <n>] [-experiment <name> | -held-out]
package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/MathTrail/mathtrail-standalone/content"
	"github.com/MathTrail/mathtrail-standalone/internal/domain/rating"
	"github.com/MathTrail/mathtrail-standalone/internal/infra/seal"
)

func main() {
	os.Exit(runCommand(os.Args[1:], os.Stdout, os.Stderr))
}

func runCommand(args []string, stdout, stderr io.Writer) int {
	out, d, err := parse(args, stderr)
	if err != nil {
		return 2
	}
	if err := experimentInto(out, d, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

// errSeedBesideHeldOut is a command line that asks for the held-out seeds and
// for a seed or a name of its own at once.
var errSeedBesideHeldOut = errors.New("learners: -held-out draws from seeds of its own, and takes no -seed or -experiment")

// parse reads the command line: the directory the results are written to and
// the run's design, and into masterSeed and experiment the seed and the name
// of the run, each the paper's when the command line does not give it, or the
// held-out ones, whose results go into a directory of their own.
func parse(args []string, stderr io.Writer) (string, design, error) {
	flags := flag.NewFlagSet("learners", flag.ContinueOnError)
	flags.SetOutput(stderr)
	out := flags.String("out", "results", "the directory the results are written to")
	children := flags.Int("children", 1000, "how many children each generator draws")
	answers := flags.Int("answers", 200, "how many answers each child gives")
	flags.Uint64Var(&masterSeed, "seed", paperSeed, "the seed every draw of the run comes from")
	flags.StringVar(&experiment, "experiment", paperExperiment, "the name of the run, which every seed it draws takes in")
	heldOut := flags.Bool("held-out", false, "draw from the seeds kept for confirming a choice, and write under "+heldOutDirectory)
	if err := flags.Parse(args); err != nil {
		return "", design{}, err
	}
	if *heldOut {
		seedGiven := false
		flags.Visit(func(f *flag.Flag) { seedGiven = seedGiven || f.Name == "seed" || f.Name == "experiment" })
		if seedGiven {
			fmt.Fprintln(stderr, errSeedBesideHeldOut)
			return "", design{}, errSeedBesideHeldOut
		}
		masterSeed, experiment = heldOutSeed, heldOutExperiment
		if *out != "" {
			*out = filepath.Join(*out, heldOutDirectory)
		}
	}
	return *out, design{children: *children, answers: *answers}, nil
}

func experimentInto(out string, d design, stdout io.Writer) error {
	if out == "" || d.children < 1 || d.answers < 1 {
		return errors.New("usage: learners [-out <directory>] [-children <n>] [-answers <n>] [-seed <n>] [-experiment <name> | -held-out]")
	}
	if last := checkpoints[len(checkpoints)-1]; d.answers < last {
		return fmt.Errorf("learners: the comparisons read the error after %d answers, so a child gives at least %d", last, last)
	}
	w, err := newWorld(d.answers)
	if err != nil {
		return err
	}
	all := cells()
	ms := metrics()
	results, err := runCells(w, all, d.children, ms)
	if err != nil {
		return err
	}
	summary, err := writeAll(out, all, results, ms, d)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "%s\n%d cells of %d children, %d answers each, written to %s\n", summary, len(all), d.children, d.answers, out)
	return nil
}

// newWorld is what every run shares: the shipped catalog, its topics and the
// points of each topic's ladder, and a sealer with a key made for the run,
// which seals the tasks as the service does.
func newWorld(answers int) (*world, error) {
	shipped, err := content.Load()
	if err != nil {
		return nil, fmt.Errorf("learners: load the content: %w", err)
	}
	key := make([]byte, seal.KeySize)
	if _, failed := rand.Read(key); failed != nil {
		return nil, fmt.Errorf("learners: make a sealing key: %w", failed)
	}
	ring, err := seal.NewKeyRing(base64.StdEncoding.EncodeToString(key), "")
	if err != nil {
		return nil, fmt.Errorf("learners: make the key ring: %w", err)
	}
	topics := shipped.TopicIDs()
	ladders := make(map[string][]rating.Point, len(topics))
	for _, topic := range topics {
		ladders[topic] = rating.Points(shipped.LevelsOf(topic)...)
	}
	return &world{catalog: shipped, topics: topics, ladders: ladders, sealer: ring.For(seal.PurposeTaskAnswer), answers: answers}, nil
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
