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
// every constraint and are better than the service, A*, its equals and the
// simplest of them, the backup set against it, or the exit; and how likely the
// chosen rule's constraints are to hold on new children. A run of the
// held-out children writes its confirmation instead.
func writeChoice(b *strings.Builder, c *stepChoice) {
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
		b.WriteString("No candidate meets every constraint and is better than the service, and no floor of the exit meets " +
			"the constraints of not worse and of the screen: **the service's step stays.**\n\n")
	case c.exit:
		fmt.Fprintf(b, "No candidate meets every constraint and is better than the service. **The exit: %s**, the floor of the "+
			"highest score, %s, among those that meet the constraints of not worse and of the screen.\n\n", ruleName(c.chosen.rule), interval3(c.chosen.score))
	default:
		writeMain(b, c)
	}
	writeMargins(b, c)
}

// writeEligible lists the candidates that meet every constraint and are
// better than the service, the highest score first, and, when they are few,
// the candidates that miss one constraint alone.
func writeEligible(b *strings.Builder, c *stepChoice) {
	fmt.Fprintf(b, "Candidates that meet every constraint and are better than the service: %d.\n\n", len(c.eligible))
	if len(c.eligible) > 0 {
		b.WriteString("| Rule | Kind | Score | Numbers added | Fields in the profile | Distance from the service's step |\n|---|---|---:|---:|---|---:|\n")
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

// writeMain writes how the main step and the choice were come to.
func writeMain(b *strings.Builder, c *stepChoice) {
	if c.best != nil {
		fmt.Fprintf(b, "A*, the main step of the highest score: %s, %s.\n\n", ruleName(c.best.rule), interval3(c.best.score))
		if len(c.equals) == 0 {
			b.WriteString("No main step's score less A*'s holds nothing: A* has no equals.\n\n")
		} else {
			b.WriteString("Its equals, whose score less A*'s holds nothing:\n\n")
			for _, e := range c.equals {
				fmt.Fprintf(b, "- %s, %s less A*'s\n", ruleName(e.rc.rule), interval3(e.difference))
			}
			b.WriteString("\n")
		}
		m := c.main.rule.candidacy
		fmt.Fprintf(b, "The simplest of them, the main step: %s — %d numbers added to the service's rule, %s, "+
			"a distance of %s from the service's step.\n\n", ruleName(c.main.rule), m.added, fieldsPhrase(m.fields), decimals3(m.distance))
	}
	switch {
	case c.backup == nil:
		b.WriteString("No backup meets every constraint and is better than the service.\n\n")
	case c.main == nil:
		fmt.Fprintf(b, "No main step meets every constraint and is better than the service; the backup does: %s.\n\n", ruleName(c.backup.rc.rule))
	default:
		verdict := "not chosen"
		if c.chosen == c.backup.rc {
			verdict = "chosen"
		}
		fmt.Fprintf(b, "The backup: %s, its score less the main step's %s, against a margin of %s: %s.\n\n",
			ruleName(c.backup.rc.rule), interval3(c.backup.difference), decimals3(backupMargin), verdict)
	}
	fmt.Fprintf(b, "**The choice: %s.**\n\n", ruleName(c.chosen.rule))
}

// writeMargins lists the chosen rule's constraints the likeliest to fail on
// new children, and the chance all of them hold.
func writeMargins(b *strings.Builder, c *stepChoice) {
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
// held-out children meets every constraint and is better than the service
// there.
func writeConfirmation(b *strings.Builder, c *stepChoice) {
	b.WriteString("## The confirmation\n\n")
	if len(c.candidates) == 0 {
		b.WriteString("The run has no candidate to confirm.\n\n")
		return
	}
	for _, rc := range c.candidates {
		missed := constraintsMissed(rc)
		better := betterThanService(rc)
		fmt.Fprintf(b, "%s: %d of %d constraints met, the score %s.", ruleName(rc.rule), len(rc.readings)-missed, len(rc.readings), interval3(rc.score))
		if missed == 0 && better {
			b.WriteString(" **Confirmed:** it meets every constraint and is better than the service on the held-out children.\n\n")
			continue
		}
		b.WriteString(" **Not confirmed:**")
		if missed > 0 {
			fmt.Fprintf(b, " it misses %s;", missedOf(rc))
		}
		if !better {
			b.WriteString(" its score is not above the service's with its whole interval;")
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
