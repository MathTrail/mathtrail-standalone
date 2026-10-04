// Command learners runs simulated children through the service's own rule,
// profile and rating code, under the service's estimate of their level and
// under the rules it is compared with, and measures how each estimate places a
// child, keeps their tasks in the corridor, declares mastery and follows a
// child who changes.
//
// Every child is drawn from a seed of its own and is run under every rule with
// the same parameters and the same draws for its answers. The results are
// written as tables of every cell, of the comparisons and of every rule across
// the generators, a summary of the numbers a change to the rule is judged by,
// which is also printed, the criterion a new step is chosen by as every rule
// meets it, with the choice it comes to, and what the run was.
//
// A run is given a set of rules: the bench's own, unless it names another —
// the sweep of candidate steps and its refinement, which draw children of
// their own, the decision run, the parts of the chosen step, or its
// confirmation on the held-out children.
//
// Usage:
//
//	learners [-out <directory>] [-children <n>] [-answers <n>] [-rules <set>] [-seed <n>] [-experiment <name>] [-held-out]
package main

import (
	"cmp"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
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

// The command lines that ask for two things at once.
var (
	// errOwnSeeds asks for a set of rules that draws children of its own and
	// for a seed or a name of the run besides.
	errOwnSeeds = errors.New("learners: this set of rules draws children of its own, and takes no -seed or -experiment")
	// errHeldOutBesideRules asks for the held-out children, which only the
	// confirmation is run on, and for another set of rules.
	errHeldOutBesideRules = errors.New("learners: -held-out runs the confirmation of the chosen step, and takes no other -rules")
)

// parse reads the command line: the directory the results are written to,
// within it the set's own directory, and the run's design, with the set of
// rules it runs; and into masterSeed and experiment the seed and the name of
// the run — the set's own, for a set that draws children of its own, or else
// the command line's, or the paper's when it gives none.
func parse(args []string, stderr io.Writer) (string, design, error) {
	flags := flag.NewFlagSet("learners", flag.ContinueOnError)
	flags.SetOutput(stderr)
	out := flags.String("out", "results", "the directory the results are written to")
	children := flags.Int("children", 1000, "how many children each generator draws")
	answers := flags.Int("answers", 200, "how many answers each child gives")
	setName := flags.String("rules", benchSet, "the set of rules to run: "+strings.Join(setNames(), ", "))
	flags.Uint64Var(&masterSeed, "seed", paperSeed, "the seed every draw of the run comes from")
	flags.StringVar(&experiment, "experiment", paperExperiment, "the name of the run, which every seed it draws takes in")
	heldOut := flags.Bool("held-out", false, "confirm the chosen step on the seeds kept for it, and write under "+heldOutDirectory)
	if err := flags.Parse(args); err != nil {
		return "", design{}, err
	}
	given := map[string]bool{}
	flags.Visit(func(f *flag.Flag) { given[f.Name] = true })
	if *heldOut {
		if given["rules"] && *setName != confirmationSet {
			return refused(stderr, errHeldOutBesideRules)
		}
		*setName = confirmationSet
	}
	set, known := ruleSetNamed(*setName)
	if !known {
		return refused(stderr, fmt.Errorf("learners: no set of rules is named %q; the sets are %s", *setName, strings.Join(setNames(), ", ")))
	}
	if set.ownSeeds() {
		if given["seed"] || given["experiment"] {
			return refused(stderr, errOwnSeeds)
		}
		masterSeed, experiment = set.seed, set.experiment
	}
	if *out != "" && set.directory != "" {
		*out = filepath.Join(*out, set.directory)
	}
	return *out, design{children: *children, answers: *answers, set: set.name}, nil
}

// refused says why a command line is refused, and refuses it.
func refused(stderr io.Writer, err error) (string, design, error) {
	fmt.Fprintln(stderr, err)
	return "", design{}, err
}

// setNames are the names of the sets of rules.
func setNames() []string {
	var names []string
	for _, s := range ruleSets() {
		names = append(names, s.name)
	}
	return names
}

func experimentInto(out string, d design, stdout io.Writer) error {
	if out == "" || d.children < 1 || d.answers < 1 {
		return errors.New("usage: learners [-out <directory>] [-children <n>] [-answers <n>] [-rules <set>] [-seed <n>] [-experiment <name>] [-held-out]")
	}
	if last := checkpoints[len(checkpoints)-1]; d.answers < last {
		return fmt.Errorf("learners: the comparisons read the error after %d answers, so a child gives at least %d", last, last)
	}
	set, known := ruleSetNamed(cmp.Or(d.set, benchSet))
	if !known {
		return fmt.Errorf("learners: no set of rules is named %q", d.set)
	}
	rules, err := set.rules()
	if err != nil {
		return err
	}
	d.set = set.name
	w, err := newWorld(d.answers)
	if err != nil {
		return err
	}
	all := cellsOf(rules)
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

// newWorld is what every run shares: the shipped catalog, its topics, the
// levels each topic is taught at, the lowest first, and the points of its
// ladder, and a sealer with a key made for the run, which seals the tasks as
// the service does.
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
	levels := make(map[string][]rating.GradeLevel, len(topics))
	ladders := make(map[string][]rating.Point, len(topics))
	for _, topic := range topics {
		levels[topic] = slices.SortedStableFunc(slices.Values(shipped.LevelsOf(topic)), func(a, b rating.GradeLevel) int {
			return cmp.Compare(a.Shift(), b.Shift())
		})
		ladders[topic] = rating.Points(shipped.LevelsOf(topic)...)
	}
	return &world{
		catalog: shipped, topics: topics, levels: levels, ladders: ladders,
		sealer: ring.For(seal.PurposeTaskAnswer), answers: answers,
	}, nil
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
