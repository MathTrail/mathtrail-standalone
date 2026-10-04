package main

import (
	"fmt"
	"strings"
)

// marginsShown is how many of the chosen rule's constraints the choice lists,
// the likeliest to fail on new children first.
const marginsShown = 8

// fewEligible is how many candidates meeting every constraint a choice wants
// before it stops listing the ones a single constraint short of it.
const fewEligible = 5

// writeChoice writes what the criterion comes to: the candidates that meet
// every constraint and are better than the baseline, A*, its equals and the
// simplest of them, the backup set against it, or the exit; and how likely the
// chosen rule's constraints are to hold on new children. A run of the
// held-out children writes its confirmation instead.
func writeChoice(b *strings.Builder, c *criterionChoice) {
	if c.confirming {
		writeConfirmation(b, c)
		return
	}
	b.WriteString("## The choice\n\n")
	if !c.decides {
		b.WriteString("As a rough look reads it, which decides nothing.\n\n")
	}
	writeEligible(b, c)
	switch {
	case c.chosen == nil:
		fmt.Fprintf(b, "No candidate meets every constraint and is better than %s, and no floor of the exit meets "+
			"the constraints of not worse and of the screen: **the service's step stays.**\n\n", c.crit.against)
	case c.exit && c.crit.exitIsBaseline:
		fmt.Fprintf(b, "No candidate meets every constraint and is better than %s. **The exit: %s**, %s itself, kept as it is.\n\n",
			c.crit.against, ruleName(c.chosen.rule), c.crit.against)
	case c.exit:
		fmt.Fprintf(b, "No candidate meets every constraint and is better than %s. **The exit: %s**, the floor of the "+
			"highest score, %s, among those that meet the constraints of not worse and of the screen.\n\n", c.crit.against, ruleName(c.chosen.rule), interval3(c.chosen.score))
	default:
		writeMain(b, c)
	}
	writeMargins(b, c)
	writeAgainstTop(b, c)
}

// writeAgainstTop lists the score of the main candidate of the highest score
// less every other rule's, on the same children: what each part of it gives,
// taken away, and how much more of the way it closes than a comparison.
func writeAgainstTop(b *strings.Builder, c *criterionChoice) {
	if c.top == nil || len(c.againstTop) == 0 {
		return
	}
	fmt.Fprintf(b, "### Every rule against the %[1]s of the highest score\n\nThe %[1]s of the highest score, whatever it meets: %[2]s, %[3]s. "+
		"Its score less every other rule's, on the same children:\n\n| Rule | Score | The highest %[1]s's less this |\n|---|---:|---:|\n",
		c.crit.main, ruleName(c.top.rule), interval3(c.top.score))
	for _, r := range c.againstTop {
		fmt.Fprintf(b, "| %s | %s | %s |\n", ruleName(r.rc.rule), interval3(r.rc.score), interval3(r.difference))
	}
	b.WriteString("\n")
}

// writeEligible lists the candidates that meet every constraint and are
// better than the baseline, the highest score first, and, when they are few,
// the candidates that miss one constraint alone.
func writeEligible(b *strings.Builder, c *criterionChoice) {
	fmt.Fprintf(b, "Candidates that meet every constraint and are better than %s: %d.\n\n", c.crit.against, len(c.eligible))
	if len(c.eligible) > 0 {
		fmt.Fprintf(b, "| Rule | Kind | Score | Numbers added | Fields in the profile | Distance from %s |\n|---|---|---:|---:|---|---:|\n", c.crit.distanceFrom)
		for _, rc := range c.eligible {
			fmt.Fprintf(b, "| %s | %s | %s | %d | %s | %s |\n", ruleName(rc.rule), rc.rule.candidacy.kind, interval3(rc.score),
				rc.rule.candidacy.added, yesOrNo(rc.rule.candidacy.fields), decimals3(rc.rule.candidacy.distance))
		}
		b.WriteString("\n")
	}
	if len(c.eligible) >= fewEligible || len(c.nearMisses) == 0 {
		return
	}
	b.WriteString("Candidates one constraint short, the highest score first:\n\n| Rule | Kind | Score | The constraint missed |\n|---|---|---:|---|\n")
	for _, rc := range c.nearMisses {
		fmt.Fprintf(b, "| %s | %s | %s | %s |\n", ruleName(rc.rule), rc.rule.candidacy.kind, interval3(rc.score), missedOf(rc))
	}
	b.WriteString("\n")
}

// writeMain writes how the main candidate and the choice were come to.
func writeMain(b *strings.Builder, c *criterionChoice) {
	if c.best != nil {
		fmt.Fprintf(b, "A*, the %s of the highest score: %s, %s.\n\n", c.crit.main, ruleName(c.best.rule), interval3(c.best.score))
		if len(c.equals) == 0 {
			fmt.Fprintf(b, "No %s's score less A*'s holds nothing: A* has no equals.\n\n", c.crit.main)
		} else {
			b.WriteString("Its equals, whose score less A*'s holds nothing:\n\n")
			for _, e := range c.equals {
				fmt.Fprintf(b, "- %s, %s less A*'s\n", ruleName(e.rc.rule), interval3(e.difference))
			}
			b.WriteString("\n")
		}
		m := c.main.rule.candidacy
		fmt.Fprintf(b, "The simplest of them, the %s: %s — %d numbers added to the service's rule, %s, "+
			"a distance of %s from the service's rule.\n\n", c.crit.main, ruleName(c.main.rule), m.added, fieldsPhrase(m.fields), decimals3(m.distance))
	}
	switch {
	case c.backup == nil:
		fmt.Fprintf(b, "No backup meets every constraint and is better than %s.\n\n", c.crit.against)
	case c.main == nil:
		fmt.Fprintf(b, "No %s meets every constraint and is better than %s; the backup does: %s.\n\n", c.crit.main, c.crit.against, ruleName(c.backup.rc.rule))
	default:
		verdict := "not chosen"
		if c.chosen == c.backup.rc {
			verdict = "chosen"
		}
		fmt.Fprintf(b, "The backup: %s, its score less the %s's %s, against a margin of %s: %s.\n\n",
			ruleName(c.backup.rc.rule), c.crit.main, interval3(c.backup.difference), decimals3(backupMargin), verdict)
	}
	fmt.Fprintf(b, "**The choice: %s.**\n\n", ruleName(c.chosen.rule))
}

// writeMargins lists the chosen rule's constraints the likeliest to fail on
// new children, and the chance all of them hold.
func writeMargins(b *strings.Builder, c *criterionChoice) {
	if c.chosen == nil || len(c.margins) == 0 {
		return
	}
	fmt.Fprintf(b, "### The chosen rule's constraints on new children\n\nThe chance each constraint holds on as many new children, "+
		"the likeliest to fail first: the value there falls around the value here with the standard error its interval shows, "+
		"and a check of not worse holds where the worse end of its interval does. The chance all %d hold, taken as independent: %s.\n\n",
		len(c.margins), percent1(allHold(c.margins)))
	b.WriteString("| Constraint | Generator | Measure | Value | Bound | Chance it holds |\n|---|---|---|---:|---:|---:|\n")
	for _, m := range c.margins[:min(len(c.margins), marginsShown)] {
		fmt.Fprintf(b, "| %s | %s | %s | %s | %s | %s |\n", m.rd.name, m.rd.generator, m.rd.metric, decimals3(m.rd.value), decimals3(m.rd.bound), percent1(m.chance))
	}
	b.WriteString("\n")
}

// writeConfirmation writes whether the one candidate of a run of the
// held-out children meets every constraint and is better than the baseline
// there.
func writeConfirmation(b *strings.Builder, c *criterionChoice) {
	b.WriteString("## The confirmation\n\n")
	if len(c.candidates) == 0 {
		b.WriteString("The run has no candidate to confirm.\n\n")
		return
	}
	for _, rc := range c.candidates {
		missed := constraintsMissed(rc)
		better := betterThanBaseline(rc)
		fmt.Fprintf(b, "%s: %d of %d constraints met, the score %s.", ruleName(rc.rule), len(rc.readings)-missed, len(rc.readings), interval3(rc.score))
		if missed == 0 && better {
			fmt.Fprintf(b, " **Confirmed:** it meets every constraint and is better than %s on the held-out children.\n\n", c.crit.against)
			continue
		}
		b.WriteString(" **Not confirmed:**")
		if missed > 0 {
			fmt.Fprintf(b, " it misses %s;", missedOf(rc))
		}
		if !better {
			fmt.Fprintf(b, " its score is not above %s's with its whole interval;", c.crit.against)
		}
		b.WriteString(" the exit chosen on the working children is taken.\n\n")
	}
}

// missedOf names the constraints a rule misses or the run cannot read.
func missedOf(rc *ruleCriterion) string {
	var missed []string
	for i := range rc.readings {
		rd := &rc.readings[i]
		switch rd.verdict {
		case met:
		case unread:
			missed = append(missed, fmt.Sprintf("%s %s unread", rd.generator, rd.metric))
		default:
			missed = append(missed, fmt.Sprintf("%s %s %s against %s", rd.generator, rd.metric, decimals3(rd.value), decimals3(rd.bound)))
		}
	}
	return strings.Join(missed, "; ")
}

func yesOrNo(yes bool) string {
	if yes {
		return "yes"
	}
	return "no"
}

func fieldsPhrase(fields bool) string {
	if fields {
		return "fields added to the profile"
	}
	return "no field added to the profile"
}
