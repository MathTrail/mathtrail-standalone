package main

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// scenarioMeasure is a measure the table of every rule across the generators
// shows: its metric, its heading, and whether it is a share, shown in percent,
// rather than a number of logits.
type scenarioMeasure struct {
	metric, heading string
	share           bool
}

// scenarioMeasures are the measures a step is judged by on every generator:
// the error after 200 answers and the share of tasks in the corridor, which no
// candidate may make worse anywhere; the share of masteries declared falsely,
// shown for the choice of mastery; the lag behind a child who learns; the
// share of children not caught up after a jump or a drop; and the error of the
// overall level after ten answers, which tells most where a child is placed a
// level off.
var scenarioMeasures = []scenarioMeasure{
	{metric: "r1_rms_200", heading: "The error after 200 answers, in logits"},
	{metric: "r3_inside", heading: "The tasks in the corridor", share: true},
	{metric: "r4_false", heading: "The masteries declared falsely", share: true},
	{metric: "r6_lag", heading: "The estimate less the level from the 101st answer on, in logits"},
	{metric: "r6_jump_unsettled", heading: "The children not caught up after a jump or a drop", share: true},
	{metric: "r7_error_10", heading: "The error of the overall level after ten answers, in logits"},
}

// scenariosText is every rule of a run across every generator, measure by
// measure, in Markdown: a row a rule, in the order the cells come in, and a
// column a generator, each the value alone, its interval being in cells.csv.
func scenariosText(all []cell, summaries [][]summary, names []string, d design) string {
	at := map[string]int{}
	var order []*rule
	for c := range all {
		at[all[c].name()] = c
		if !containsRule(order, all[c].rule) {
			order = append(order, all[c].rule)
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# Every rule on every generator\n\nSeed %d, experiment %s: %d children a generator, %d answers each. "+
		"Each number is read off all the children of its cell; its interval is in cells.csv.\n\n", masterSeed, experiment, d.children, d.answers)
	for _, m := range scenarioMeasures {
		value := func(r *rule, g generator) string {
			c, found := at[(&cell{rule: r, generator: g}).name()]
			i := slices.Index(names, m.metric)
			if !found || i < 0 || !summaries[c][i].has {
				return noNumber
			}
			return m.written(summaries[c][i].value)
		}
		writeMeasure(&b, m.heading, order, value)
	}
	return b.String()
}

// writeMeasure writes one measure's table: a row a rule, a column a
// generator.
func writeMeasure(b *strings.Builder, heading string, order []*rule, value func(r *rule, g generator) string) {
	fmt.Fprintf(b, "## %s\n\n| Rule |", heading)
	for _, g := range allGenerators {
		fmt.Fprintf(b, " %s |", g)
	}
	b.WriteString("\n|---|" + strings.Repeat("---:|", len(allGenerators)) + "\n")
	for _, r := range order {
		fmt.Fprintf(b, "| %s |", ruleName(r))
		for _, g := range allGenerators {
			fmt.Fprintf(b, " %s |", value(r, g))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// written is a measure's value as the table shows it: a share in percent to a
// tenth, a number of logits to a thousandth.
func (m scenarioMeasure) written(v float64) string {
	if m.share {
		return strconv.FormatFloat(100*v, 'f', 1, 64)
	}
	return decimals3(v)
}
