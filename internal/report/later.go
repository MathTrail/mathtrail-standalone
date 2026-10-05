package report

import (
	"cmp"
	"maps"
	"math"
	"slices"
	"strconv"
)

// The estimate is set against itself as the answers pile up: each child's
// answers early on and later, each weighed against the chance promised, and
// the later less the earlier. Whatever leans the same way in both ranges falls
// out of the difference — the answers the weighing leaves out, a model that
// misses the difficulty it is asked for —, and a child is set against itself
// alone, so that the children who stop before the later range are not read as
// an estimate that falls behind.

// The two ranges of the child's answers set against each other: the last
// answer of the earlier, and the first and the last of the later. The earlier
// ends before the overall level's step comes to its floor; the later stays
// past it, and ends where a child of the learners' bench stops answering, so
// that the live numbers and the bench's can be laid side by side.
const (
	EarlierTo = 50
	LaterFrom = 101
	LaterTo   = 200
)

// everyHost stands for the hosts of a version taken together. It is no word
// the service writes, so that it is never taken for a client's name.
const everyHost = "(every host)"

// side is which of the two ranges a range of the child's answers is wholly
// within, if either.
type side int

const (
	neither side = iota
	earlier
	later
)

// sideOf is the range a range of the child's answers, as the service names
// it, falls wholly within: 6-20 and 21-50 earlier, 101-200 later, and neither
// for a range that crosses a boundary, one open at its end, or one named
// otherwise.
func sideOf(named string) side {
	first, hasFirst := firstAnswerOf(named)
	last, hasLast := lastAnswerOf(named)
	switch {
	case !hasFirst || !hasLast:
		return neither
	case last <= EarlierTo:
		return earlier
	case first >= LaterFrom && last <= LaterTo:
		return later
	}
	return neither
}

// crossesAnEdge says whether a range of the child's answers, as the service
// names it, holds answers on both sides of where one of the two ranges begins
// or ends, and so is read in neither though part of it belongs to one: a range
// the service writes only if it counts in other ranges than these. One open at
// its end runs on past every edge after its first answer.
func crossesAnEdge(named string) bool {
	first, hasFirst := firstAnswerOf(named)
	if !hasFirst {
		return false
	}
	last, hasLast := lastAnswerOf(named)
	if !hasLast {
		last = math.MaxInt
	}
	for _, edge := range []int{EarlierTo, LaterFrom - 1, LaterTo} {
		if first <= edge && edge < last {
			return true
		}
	}
	return false
}

// crossingEdges are the ranges of the child's answers the lines name that
// cross where one of the two ranges begins or ends, in their order.
func (c *counts) crossingEdges() []string {
	var crossing []string
	for key := range c.keptUp {
		if crossesAnEdge(key.answers) && !slices.Contains(crossing, key.answers) {
			crossing = append(crossing, key.answers)
		}
	}
	slices.SortFunc(crossing, compareAnswerRanges)
	return crossing
}

// pair is one child's answers in the two ranges.
type pair struct{ earlier, later ofChild }

// pairsByGroup are, for every group and for every version's hosts together,
// each child's answers in the two ranges, from the cells of how the estimate
// keeps up. A host that hands a task to no child a family has, a child such a
// host handed a task to, and an answer that names no child are left out.
func (c *counts) pairsByGroup() map[group]map[string]*pair {
	pairs := map[group]map[string]*pair{}
	for key, cell := range c.keptUp {
		within := sideOf(key.answers)
		if within == neither || slices.Contains(notChildren, key.host) {
			continue
		}
		for _, g := range []group{key.group, {version: key.version, host: everyHost}} {
			if pairs[g] == nil {
				pairs[g] = map[string]*pair{}
			}
			c.addChildren(pairs[g], cell, within)
		}
	}
	return pairs
}

// addChildren adds each child's answers in a cell to that child's range on
// the side the cell is on. A child a host of notChildren handed a task to, and
// an answer that names no child, are left out.
func (c *counts) addChildren(byChild map[string]*pair, cell *keptUp, within side) {
	for user, answered := range cell.byChild {
		if user == "" || c.testing[user] {
			continue
		}
		p := byChild[user]
		if p == nil {
			p = &pair{}
			byChild[user] = p
		}
		p.add(within, answered)
	}
}

// add adds a child's answers in a range to the side the range is on.
func (p *pair) add(within side, answered *ofChild) {
	to := &p.earlier
	if within == later {
		to = &p.later
	}
	to.answers += answered.answers
	to.sum += answered.sum
}

// comparison is the children who answered in both ranges, set against
// themselves: their answers in each range, together; how much further the
// later came out from their promise than the earlier, on average over the
// answers of each range; and its standard error, counted by child.
type comparison struct {
	children       int
	earlier, later ofChild
	difference     float64
	standardError  float64
	hasError       bool
}

// compared sets the children of a group who answered in both ranges against
// themselves. A child's answers in the two ranges lean together, so the
// error of the difference is counted by child, the two ranges of a child
// together: as how the estimate keeps up counts the error of one range, with
// each child's share of the difference in place of its share of the mean.
func compared(byChild map[string]*pair) comparison {
	var both []*pair
	var result comparison
	for _, user := range slices.Sorted(maps.Keys(byChild)) {
		p := byChild[user]
		if p.earlier.answers == 0 || p.later.answers == 0 {
			continue
		}
		both = append(both, p)
		result.earlier.answers += p.earlier.answers
		result.earlier.sum += p.earlier.sum
		result.later.answers += p.later.answers
		result.later.sum += p.later.sum
	}
	result.children = len(both)
	if result.children == 0 {
		return result
	}
	earlierMean := result.earlier.sum / float64(result.earlier.answers)
	laterMean := result.later.sum / float64(result.later.answers)
	result.difference = laterMean - earlierMean
	if result.children < 2 {
		return result
	}
	var squares float64
	for _, p := range both {
		share := (p.later.sum-float64(p.later.answers)*laterMean)/float64(result.later.answers) -
			(p.earlier.sum-float64(p.earlier.answers)*earlierMean)/float64(result.earlier.answers)
		squares += share * share
	}
	children := float64(result.children)
	result.standardError, result.hasError = math.Sqrt(children/(children-1)*squares), true
	return result
}

// laterAbout says how to read the table of the later set against the earlier,
// and names a range of answers the lines hold that neither range can take.
func (c *counts) laterAbout() string {
	about := "The same answers once more, of the children with answers in two ranges of their own: earlier, from the " +
		"first after the trial series to the " + ordinal(EarlierTo) + ", and later, from the " +
		ordinal(LaterFrom) + " to the " + ordinal(LaterTo) + " — where a child of the learners' bench " +
		"stops answering, so that the two can be laid side by side. Each child is set against itself: in each range " +
		"a right answer as 1 and a wrong one as 0, less the chance promised, on average over the answers there, and " +
		"the later less the earlier, with its standard error counted by child, a child's answers in both ranges " +
		"together. What leans the same way in both ranges falls out of the difference — the answers the weighing " +
		"leaves out, a model that misses the difficulty it is asked for —, and a child who stopped before its " +
		ordinal(LaterFrom) + " answer is left out rather than read against the children who went on. Above " +
		"zero, the later answers came out right more often against their promise than the earlier: the estimate " +
		"falls further behind as the answers pile up; below zero, it falls behind less, or runs ahead. " + everyHost +
		" takes a version's hosts together, so a child who answered the two ranges in different hosts is set against " +
		"itself there alone; a child the load tool or MCP Inspector handed a task to is left out, and so are their " +
		"calls. A range of fewer than " + strconv.Itoa(fewestAnswers) + " answers says " + tooFew +
		"; the numbers are to three places."
	if crossing := c.crossingEdges(); len(crossing) > 0 {
		about += " The lines name ranges of answers that cross where these begin or end, which neither takes: " +
			inWords(crossing) + "."
	}
	return about
}

// laterTable is, for every group and every version's hosts together, the
// children who answered in both ranges, their answers in each, how far each
// range came out from its promise, and the later less the earlier with its
// standard error.
func (c *counts) laterTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Children", "Answers earlier", "Came true less promised, earlier",
		"Answers later", "Came true less promised, later", "Later less earlier", "Standard error",
	}, named: 2}
	pairs := c.pairsByGroup()
	for _, g := range slices.SortedFunc(maps.Keys(pairs), c.compareRows) {
		result := compared(pairs[g])
		if result.children == 0 {
			continue
		}
		earlierMean, laterMean, difference, standardError := tooFew, tooFew, tooFew, tooFew
		if result.earlier.answers >= fewestAnswers {
			earlierMean = signedTo(result.earlier.sum/float64(result.earlier.answers), 3)
		}
		if result.later.answers >= fewestAnswers {
			laterMean = signedTo(result.later.sum/float64(result.later.answers), 3)
		}
		if result.earlier.answers >= fewestAnswers && result.later.answers >= fewestAnswers {
			difference, standardError = signedTo(result.difference, 3), oneChild
			if result.hasError {
				standardError = strconv.FormatFloat(result.standardError, 'f', 3, 64)
			}
		}
		t.add(g.version, g.host, number(result.children), number(result.earlier.answers), earlierMean,
			number(result.later.answers), laterMean, difference, standardError)
	}
	return t
}

// compareRows orders the rows of a table by version, in the order the versions
// first appear, then by host, with a version's hosts taken together last.
func (c *counts) compareRows(a, b group) int {
	together := func(g group) int {
		if g.host == everyHost {
			return 1
		}
		return 0
	}
	return cmp.Or(
		cmp.Compare(slices.Index(c.versions, a.version), slices.Index(c.versions, b.version)),
		cmp.Compare(together(a), together(b)),
		cmp.Compare(a.host, b.host),
	)
}
