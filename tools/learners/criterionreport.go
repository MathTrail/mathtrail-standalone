package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// criterionText is the criterion as a run reads it, in Markdown: every rule's
// score and how many of its constraints it meets, every rule's value against
// each goal's bound, what of not worse each rule does not meet, what the bench
// resolves on each check of not worse, and the choice it all comes to.
func criterionText(read []ruleCriterion, resolutions []checkResolution, choice *criterionChoice, d design, crit *criterion) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# The criterion, read on this run\n\n")
	fmt.Fprintf(&b, "Seed %d, experiment %s: %d children a generator, %d answers each. ", masterSeed, experiment, d.children, d.answers)
	if d.children < decisionChildren {
		fmt.Fprintf(&b, "A rough look, not a decision: a decision runs %d children a cell, and a rough look reads not worse as no clear harm. ", decisionChildren)
	}
	b.WriteString("The score is " + crit.way + ".\n\n")
	writeChoice(&b, choice)
	writeScores(&b, read)
	writeGoals(&b, read)
	writeNotWorse(&b, read, crit.against)
	writeResolutions(&b, resolutions, d.children, crit.against)
	return b.String()
}

func writeScores(b *strings.Builder, read []ruleCriterion) {
	b.WriteString("## Scores\n\n| Rule | Score | Constraints met | Not met | Unread |\n|---|---:|---:|---:|---:|\n")
	for _, rc := range read {
		counts := map[string]int{}
		for i := range rc.readings {
			counts[rc.readings[i].verdict]++
		}
		fmt.Fprintf(b, "| %s | %s | %d of %d | %d | %d |\n",
			ruleName(rc.rule), interval3(rc.score), counts[met], len(rc.readings), counts[notMet], counts[unread])
	}
	b.WriteString("\n")
}

// writeGoals lists the goals, numbered, and gives every rule a row of its
// values against them, so that the table reads the same for two rules or for
// hundreds.
func writeGoals(b *strings.Builder, read []ruleCriterion) {
	b.WriteString("## Goals\n\nEvery rule's value against each goal's bound, and whether its interval reaches the bound, " +
		"stands on its edge or does not reach it. The goals:\n\n")
	var places []int
	for i := range read[0].readings {
		rd := &read[0].readings[i]
		if rd.kind != goalKind {
			continue
		}
		places = append(places, i)
		fmt.Fprintf(b, "%d. %s — %s, %s, bound %s\n", len(places), rd.name, rd.generator, rd.metric, decimals3(rd.bound))
	}
	b.WriteString("\n| Rule |")
	for n := range places {
		fmt.Fprintf(b, " %d |", n+1)
	}
	b.WriteString("\n|---|" + strings.Repeat("---|", len(places)) + "\n")
	for _, rc := range read {
		fmt.Fprintf(b, "| %s |", ruleName(rc.rule))
		for _, i := range places {
			fmt.Fprintf(b, " %s |", valueAndMark(&rc.readings[i]))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

// writeNotWorse lists what of not worse each rule does not meet, against the
// baseline, which the criterion names as against.
func writeNotWorse(b *strings.Builder, read []ruleCriterion, against string) {
	fmt.Fprintf(b, "## Not worse than %[1]s\n\nWhat of not worse each rule does not meet, or the run cannot read: its difference "+
		"from %[1]s, with its interval, and the tolerance it is read with.\n\n| Rule | Checks met | Not met or unread |\n|---|---:|---|\n", against)
	for _, rc := range read {
		checks, metCount := 0, 0
		var missed []string
		for i := range rc.readings {
			rd := &rc.readings[i]
			if rd.kind != notWorseKind {
				continue
			}
			checks++
			switch rd.verdict {
			case met:
				metCount++
			case unread:
				missed = append(missed, fmt.Sprintf("%s %s unread", rd.generator, rd.metric))
			default:
				missed = append(missed, fmt.Sprintf("%s %s %s [%s, %s] against %s",
					rd.generator, rd.metric, decimals3(rd.value), decimals3(rd.low), decimals3(rd.high), decimals3(rd.bound)))
			}
		}
		fmt.Fprintf(b, "| %s | %d of %d | %s |\n", ruleName(rc.rule), metCount, checks, strings.Join(missed, "; "))
	}
	b.WriteString("\n")
}

// writeResolutions lists what the bench resolves on each check of not worse,
// read against the baseline, which the criterion names as against.
func writeResolutions(b *strings.Builder, resolutions []checkResolution, children int, against string) {
	fmt.Fprintf(b, "## What the bench resolves\n\nEach check of not worse: the author's tolerance, the standard error of the paired "+
		"difference between %[1]s and %[2]s, the tolerance the check is read with — never under %.1[3]f standard errors — "+
		"and the chance a candidate exactly as good as %[2]s passes it, read as this run reads it and as a decision run does.\n\n",
		measuringRule, against, resolutionWidths)
	fmt.Fprintf(b, "| Generator | Measure | Author's tolerance | Standard error | Tolerance read | Passes, %d children | Passes, %d children |\n",
		children, decisionChildren)
	b.WriteString("|---|---|---:|---:|---:|---:|---:|\n")
	for i := range resolutions {
		res := &resolutions[i]
		if !res.has {
			fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s | %s |\n",
				res.check.generator, res.check.metric, decimals3(res.check.tolerance), noNumber, decimals3(res.tolerance), noNumber, noNumber)
			continue
		}
		atDecision := res.se * math.Sqrt(float64(children)/decisionChildren)
		toleranceAtDecision := max(res.check.tolerance, resolutionWidths*atDecision)
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s | %s |\n",
			res.check.generator, res.check.metric, decimals3(res.check.tolerance), strconv.FormatFloat(res.se, 'f', 4, 64),
			decimals3(res.tolerance), percent1(passChance(res.tolerance, res.se, children >= decisionChildren)), percent1(passChance(toleranceAtDecision, atDecision, true)))
	}
}

// criterionTable is the criterion as a run reads it, one row a reading and
// one a score, for whatever reads the criterion as data.
func criterionTable(read []ruleCriterion, crit *criterion) [][]string {
	table := [][]string{{"rule", "structure", "kind", "name", "generator", "metric", "bound", "value", "low", "high", "verdict", "mark"}}
	scoreName := "the share of the way to " + crit.endOfWay()
	for _, rc := range read {
		r := rc.rule
		score := rc.score
		scoreRow := []string{r.name, string(r.shape), "score", scoreName, "", "", "", "", "", "", unread, ""}
		if score.has {
			scoreRow = []string{r.name, string(r.shape), "score", scoreName, "", "", "",
				number(score.value), number(score.low), number(score.high), "", ""}
		}
		table = append(table, scoreRow)
		for i := range rc.readings {
			rd := &rc.readings[i]
			row := []string{r.name, string(r.shape), rd.kind, rd.name, string(rd.generator), rd.metric, number(rd.bound), "", "", "", rd.verdict, rd.mark}
			if rd.verdict != unread {
				row[7], row[8], row[9] = number(rd.value), number(rd.low), number(rd.high)
			}
			table = append(table, row)
		}
	}
	return table
}

// valueAndMark is a reading as the table of goals shows it: the value and
// the mark, or a dash for a value the run cannot read.
func valueAndMark(rd *reading) string {
	if rd.verdict == unread {
		return noNumber
	}
	return decimals3(rd.value) + ", " + rd.mark
}

// interval3 is a number with its interval, to three places, or a dash.
func interval3(s summary) string {
	if !s.has {
		return noNumber
	}
	return decimals3(s.value) + " [" + decimals3(s.low) + ", " + decimals3(s.high) + "]"
}

func decimals3(v float64) string { return strconv.FormatFloat(v, 'f', 3, 64) }

func percent1(share float64) string { return strconv.FormatFloat(100*share, 'f', 1, 64) + " %" }
