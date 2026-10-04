package main

import (
	"fmt"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/version"
)

// design is what a run was given besides its seed and its name: how many
// children a cell draws, how many answers each gives, and the set of rules it
// runs.
type design struct {
	children, answers int
	set               string
}

// lines is the run as run.txt states it: what decides its numbers — the seed,
// the name, the children, the answers and the set of rules — and what
// computed them: the service's version, Go's, and the processor's
// architecture, on which a multiplication and an addition may be fused into
// one step that rounds once, which moves the last places of a number.
func (d design) lines() string {
	return strings.Join([]string{
		"seed=" + strconv.FormatUint(masterSeed, 10),
		"experiment=" + experiment,
		"children=" + strconv.Itoa(d.children),
		"answers=" + strconv.Itoa(d.answers),
		"rules=" + d.set,
		"version=" + version.Version,
		"go=" + runtime.Version(),
		"arch=" + runtime.GOARCH,
	}, "\n") + "\n"
}

// headline is a number the summary gives every rule: a metric read off the
// children of one generator, in logits or, for a share, in percent.
type headline struct {
	heading   string
	metric    string
	generator generator
	share     bool
}

// headlines are the numbers of the summary: how far the estimate stands from a
// child who stays put after 200 answers; the share of the tasks of a child who
// stays put, and of one who learns, that fall in the corridor; the share of the
// masteries declared to a child who stays put that are false; how far the
// estimate stands from a child who learns, from the 101st answer on, below
// zero when it lags behind; and the share of children who jump that the
// estimate has not caught up with by their last answer.
var headlines = []headline{
	{heading: "Error after 200 answers, G0", metric: "r1_rms_200", generator: staticChildren},
	{heading: "In the corridor, G0", metric: "r3_inside", generator: staticChildren, share: true},
	{heading: "In the corridor, G2", metric: "r3_inside", generator: learning, share: true},
	{heading: "Mastered falsely, G0", metric: "r4_false", generator: staticChildren, share: true},
	{heading: "Estimate less level, G2", metric: "r6_lag", generator: learning},
	{heading: "Not caught up after the jump, G3", metric: "r6_jump_unsettled", generator: jumping, share: true},
}

// noNumber stands in the summary for a number the run has none of: too few
// children or answers to read it off.
const noNumber = "—"

// summaryOf is the summary of a run as a Markdown table: a row for every rule,
// in the order the cells come in, with its headlines and their intervals. A
// headline whose metric or cell the run lacks is an error, not a column of
// dashes.
func summaryOf(all []cell, summaries [][]summary, names []string, d design) (string, error) {
	at := map[string]int{}
	var order []*rule
	for c := range all {
		at[all[c].name()] = c
		if !slices.ContainsFunc(order, func(r *rule) bool { return sameRule(r, all[c].rule) }) {
			order = append(order, all[c].rule)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# The student model on simulated children\n\n")
	fmt.Fprintf(&b, "Seed %d, experiment %s: %d children a generator, %d answers each. "+
		"G0 is children who stay put, G2 children who learn, G3 children who jump once. "+
		"Each number is read off all the children of its cell, "+
		"with a %.0f %% interval from %d samples of them drawn again.\n\n",
		masterSeed, experiment, d.children, d.answers, 100*(1-2*tail), resamples)
	columns := []string{"Rule"}
	for _, h := range headlines {
		columns = append(columns, h.heading)
	}
	fmt.Fprintf(&b, "| %s |\n|---%s|\n", strings.Join(columns, " | "), strings.Repeat("|---:", len(headlines)))
	for _, r := range order {
		row := []string{ruleName(r)}
		for _, h := range headlines {
			i := slices.Index(names, h.metric)
			c, found := at[r.name+"/"+string(r.shape)+"/"+string(h.generator)]
			if i < 0 || !found {
				return "", fmt.Errorf("learners: the summary reads %s on %s under %s/%s, which the run lacks", h.metric, h.generator, r.name, r.shape)
			}
			row = append(row, h.written(summaries[c][i]))
		}
		fmt.Fprintf(&b, "| %s |\n", strings.Join(row, " | "))
	}
	return b.String(), nil
}

// sameRule says whether two rules are one: the same name in the same
// structure, which is what names their cells.
func sameRule(a, b *rule) bool { return a.name == b.name && a.shape == b.shape }

// ruleName is a rule as the summary names it: as its cells name it, the
// service's own and the ceiling marked.
func ruleName(r *rule) string {
	name := r.name + "/" + string(r.shape)
	switch {
	case r.service:
		name += " (the service)"
	case r.ceiling:
		name += " (the ceiling)"
	}
	return name
}

// written is a headline's number with its interval, a share in percent to a
// tenth and a number of logits to a thousandth.
func (h headline) written(s summary) string {
	if !s.has {
		return noNumber
	}
	if h.share {
		percent := func(v float64) string { return strconv.FormatFloat(100*v, 'f', 1, 64) }
		return fmt.Sprintf("%s %% [%s, %s]", percent(s.value), percent(s.low), percent(s.high))
	}
	logits := func(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }
	return fmt.Sprintf("%s [%s, %s]", logits(s.value), logits(s.low), logits(s.high))
}
