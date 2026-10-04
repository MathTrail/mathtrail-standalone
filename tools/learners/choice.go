package main

import (
	"cmp"
	"math"
	"math/rand/v2"
	"slices"
)

// backupMargin is how much more of the way the backup must close than the
// main step chosen, with the whole interval of the difference, to be chosen
// over it: a gain smaller than that does not pay for new fields in the
// profile.
const backupMargin = 0.05

// rivalry is a candidate beside another, by the difference of their scores
// on the same children.
type rivalry struct {
	rc         *ruleCriterion
	difference summary
}

// stepChoice is what the criterion of a run comes to. Of the candidates that
// meet every constraint and are better than the service, the main step of the
// highest score is A*, and the main steps whose score's difference from A*'s
// holds nothing are its equals; the simplest of them is the main step chosen.
// The best backup is chosen over it only by the margin, or where no main step
// is there to choose. With no candidate, the exit is the floor of the highest
// score among those that meet the constraints of not worse and of the screen;
// with none of those, the service's step stays. A run of the held-out
// children confirms its one candidate, or not.
type stepChoice struct {
	decides, confirming bool
	candidates          []*ruleCriterion
	eligible            []*ruleCriterion
	nearMisses          []*ruleCriterion
	best, main          *ruleCriterion
	equals              []rivalry
	backup              *rivalry
	chosen              *ruleCriterion
	exit                bool
	margins             []margin
}

// nearMissCount is how many of the candidates that fail one constraint alone
// a choice lists when fewer than that many meet them all, for the decision
// run to be given candidates that come close.
const nearMissCount = 10

// choose reads the choice off a run's criterion.
func (cr *criterionRun) choose(read []ruleCriterion, confirming bool) stepChoice {
	c := stepChoice{decides: cr.decides, confirming: confirming}
	for i := range read {
		rc := &read[i]
		if !isCandidate(rc.rule) {
			continue
		}
		c.candidates = append(c.candidates, rc)
		switch missed := constraintsMissed(rc); {
		case missed == 0 && betterThanService(rc):
			c.eligible = append(c.eligible, rc)
		case missed == 1:
			c.nearMisses = append(c.nearMisses, rc)
		}
	}
	byScore := func(a, b *ruleCriterion) int { return cmp.Compare(b.score.value, a.score.value) }
	slices.SortStableFunc(c.eligible, byScore)
	slices.SortStableFunc(c.nearMisses, byScore)
	c.nearMisses = c.nearMisses[:min(len(c.nearMisses), nearMissCount)]
	c.chooseMain(cr)
	c.weighBackup(cr)
	if c.chosen == nil {
		c.chooseExit(read)
	}
	if c.chosen != nil {
		c.margins = marginsOf(c.chosen)
	}
	return c
}

// isCandidate says whether a rule is a candidate for the step, main or
// backup, rather than for the exit or none.
func isCandidate(r *rule) bool {
	return r.candidacy != nil && r.candidacy.kind != exitCandidate
}

// chooseMain finds A*, its equals, and the simplest of them.
func (c *stepChoice) chooseMain(cr *criterionRun) {
	for _, rc := range c.eligible {
		if rc.rule.candidacy.kind != mainCandidate {
			continue
		}
		if c.best == nil {
			c.best = rc
			continue
		}
		if d := cr.scoreDifference(rc.rule, c.best.rule); d.has && d.low <= 0 && d.high >= 0 {
			c.equals = append(c.equals, rivalry{rc: rc, difference: d})
		}
	}
	if c.best == nil {
		return
	}
	c.main = c.best
	for _, e := range c.equals {
		if simpler(e.rc.rule.candidacy, c.main.rule.candidacy) {
			c.main = e.rc
		}
	}
	c.chosen = c.main
}

// simpler says whether one candidate is simpler than another: it adds fewer
// numbers to the service's rule, or as many and no field to the profile where
// the other does, or as much of both and stands nearer the service's step.
func simpler(a, b *candidacy) bool {
	if a.added != b.added {
		return a.added < b.added
	}
	if a.fields != b.fields {
		return !a.fields
	}
	return a.distance < b.distance
}

// weighBackup sets the best backup against the main step chosen, and chooses
// it where it closes more of the way by the margin, or where there is no main
// step to choose.
func (c *stepChoice) weighBackup(cr *criterionRun) {
	for _, rc := range c.eligible {
		if rc.rule.candidacy.kind != backupCandidate {
			continue
		}
		if c.main == nil {
			c.backup, c.chosen = &rivalry{rc: rc}, rc
			return
		}
		gain := cr.scoreDifference(rc.rule, c.main.rule)
		c.backup = &rivalry{rc: rc, difference: gain}
		if gain.has && gain.low > backupMargin {
			c.chosen = rc
		}
		return
	}
}

// chooseExit takes the floor of the highest score among those that meet the
// constraints of not worse and of the screen; the goals do not apply to it.
func (c *stepChoice) chooseExit(read []ruleCriterion) {
	for i := range read {
		rc := &read[i]
		if rc.rule.candidacy == nil || rc.rule.candidacy.kind != exitCandidate || !rc.score.has || !meetsNotWorseAndScreen(rc) {
			continue
		}
		if c.chosen == nil || rc.score.value > c.chosen.score.value {
			c.chosen, c.exit = rc, true
		}
	}
}

// constraintsMissed is how many of its constraints a rule does not meet, or
// the run cannot read.
func constraintsMissed(rc *ruleCriterion) int {
	missed := 0
	for i := range rc.readings {
		if rc.readings[i].verdict != met {
			missed++
		}
	}
	return missed
}

// betterThanService says whether a rule's score lies above the service's,
// none, with its whole interval.
func betterThanService(rc *ruleCriterion) bool { return rc.score.has && rc.score.low > 0 }

// meetsNotWorseAndScreen says whether a rule meets every check of not worse
// and every constraint of the screen.
func meetsNotWorseAndScreen(rc *ruleCriterion) bool {
	for i := range rc.readings {
		rd := &rc.readings[i]
		if (rd.kind == notWorseKind || rd.screen) && rd.verdict != met {
			return false
		}
	}
	return true
}

// margin is a constraint of the chosen rule as as many new children would
// read it: the chance it holds on them.
type margin struct {
	rd     *reading
	chance float64
}

// marginsOf are a rule's constraints by the chance each holds on new
// children, the likeliest to fail first.
func marginsOf(rc *ruleCriterion) []margin {
	var all []margin
	for i := range rc.readings {
		if rd := &rc.readings[i]; rd.verdict != unread {
			all = append(all, margin{rd: rd, chance: holdChance(rd)})
		}
	}
	slices.SortStableFunc(all, func(a, b margin) int { return cmp.Compare(a.chance, b.chance) })
	return all
}

// holdChance is the chance a constraint met on these children holds on as
// many new ones. The value there falls around the value here with the
// standard error the interval shows, and the value here around the truth with
// the same, so the two differ by √2 standard errors. A goal holds where the
// new value is within its bound; a check of not worse where, as a decision
// run reads it, the worse end of the new interval is within its tolerance.
// The noise of the bound itself on new children is left out.
func holdChance(rd *reading) float64 {
	se := (rd.high - rd.low) / 2 / normalQuantile
	room := rd.bound - rd.value
	switch {
	case rd.kind == notWorseKind && rd.higherIsBetter:
		room = rd.bound + rd.value - normalQuantile*se
	case rd.kind == notWorseKind:
		room -= normalQuantile * se
	case rd.atLeast:
		room = rd.value - rd.bound
	}
	if se <= 0 {
		if room >= 0 {
			return 1
		}
		return 0
	}
	return 0.5 * (1 + math.Erf(room/(2*se)))
}

// allHold is the chance every constraint holds on new children, taken as
// independent of one another.
func allHold(margins []margin) float64 {
	chance := 1.0
	for _, m := range margins {
		chance *= m.chance
	}
	return chance
}

// scoreDifference is one rule's score less another's, read on the same
// samples of each generator's children, with its interval: the service's part
// of both cancels, and what is left is how much more of the way the first
// closes on the same children.
func (cr *criterionRun) scoreDifference(a, b *rule) summary {
	first, hasFirst := cr.sheetOf(a)
	second, hasSecond := cr.sheetOf(b)
	if !hasFirst || !hasSecond {
		return summary{}
	}
	rng := seeded("criterion/"+a.name+"/"+string(a.shape)+"/less/"+b.name+"/"+string(b.shape)+"/score", "bootstrap")
	return bootstrapped(first.counts(), rng, func(samples [][]int) (float64, bool) {
		x, hasX := first.read(samples)
		y, hasY := second.read(samples)
		return x - y, hasX && hasY
	})
}

// counts are how many children each generator of a score sheet has.
func (sheet *scoreSheet) counts() []int {
	counts := make([]int, len(sheet.parts))
	for p, part := range sheet.parts {
		counts[p] = len(part.rule)
	}
	return counts
}

// bootstrapped reads a number off every child of each generator, and its
// interval off samples of each generator's children drawn again, apart.
func bootstrapped(counts []int, rng *rand.Rand, readOff func(samples [][]int) (float64, bool)) summary {
	whole := make([][]int, len(counts))
	samples := make([][]int, len(counts))
	for p, n := range counts {
		whole[p], samples[p] = everyone(n), make([]int, n)
	}
	value, has := readOff(whole)
	if !has {
		return summary{}
	}
	draws := make([]float64, 0, resamples)
	for range resamples {
		drawSamples(samples, rng)
		if v, ok := readOff(samples); ok {
			draws = append(draws, v)
		}
	}
	if len(draws) == 0 {
		return summary{}
	}
	low, high := interval(draws)
	return summary{value: value, low: low, high: high, has: true}
}
