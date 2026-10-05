package report

import (
	"cmp"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// The answers a chance is weighed against are those to tasks the rule chose,
// given after the trial series and without the hint. The rule sets a task at
// the chance it promises; while the series runs, the estimate is still finding
// the child's place; and a hint changes the chance a task was set at. The
// model's own choices are left out with them, since a model may choose a task
// off the chance on purpose.

// fewestAnswers is how many answers a cell needs before its share or mean is
// worth reading: under it, one answer more or less moves it by several
// hundredths.
const fewestAnswers = 30

// The words a cell says in place of a number it cannot give: too few answers
// to read, or the answers of one child alone, whose error says nothing of
// children in general.
const (
	tooFew   = "too few"
	oneChild = "one child"
)

// The kinds of answer the weighing leaves out, each as the report names it
// after "the answers".
const (
	leftModel   = "to tasks the model chose"
	leftPerson  = "to tasks whose topic a person chose"
	leftUnknown = "to tasks whose chooser is not known"
	leftTrial   = "in the trial series"
	leftHint    = "given with the hint"
	leftOld     = "on lines that carry no chance or no range of answers"
)

// leftOutInOrder is the order the report names the kinds left out in.
var leftOutInOrder = []string{leftModel, leftPerson, leftUnknown, leftTrial, leftHint, leftOld}

// whyLeftOut is the kind of answer a line is that the weighing leaves out, or
// nothing for one it takes. A line can be of several kinds; it is named by the
// first of them it is asked about. The weighing is of the rule's own choices:
// a topic a person chose is set at the rule's point too, but which topics are
// chosen is the person's, and the tasks the rule chose are what it is judged by.
func whyLeftOut(l *line) string {
	switch {
	case l.Chance == nil:
		return leftOld
	case l.TutorMode == string(profile.TutorLLM):
		return leftModel
	case l.TutorMode == string(profile.TutorPerson):
		return leftPerson
	case l.TutorMode != string(profile.TutorRule):
		return leftUnknown
	case l.Trial != 0:
		return leftTrial
	case l.HintUsed:
		return leftHint
	case l.AnswersBucket == "":
		return leftOld
	}
	return ""
}

// chanceRange is a range of the chance a task was handed out at, by the
// highest chance it holds in hundredths, as the line writes a chance.
type chanceRange struct {
	upTo  int
	named string
}

// chanceRanges are the ranges a chance is counted in. The two inside the
// corridor are split at its middle, where the rule aims.
var chanceRanges = []chanceRange{
	{49, "under 0.50"},
	{59, "0.50-0.59"},
	{69, "0.60-0.69"},
	{77, "0.70-0.77"},
	{85, "0.78-0.85"},
	{100, "over 0.85"},
}

// rangeOfChance is the place in chanceRanges of the range a chance falls in.
// A chance past the last range is in the last one.
func rangeOfChance(chance float64) int {
	hundredths := int(math.Round(chance * 100))
	if at := slices.IndexFunc(chanceRanges, func(r chanceRange) bool { return hundredths <= r.upTo }); at >= 0 {
		return at
	}
	return len(chanceRanges) - 1
}

// promisedIn is a cell of the table of chances: a group, and a range of the
// chance a task was handed out at.
type promisedIn struct {
	group
	chance int
}

// cameTrue is how the answers of a cell of chances came out: how many there
// were, the chances they were promised, summed, how many were right, and the
// children who gave them.
type cameTrue struct {
	answers, right int
	promised       float64
	children       map[string]bool
}

// keptUpIn is a cell of the table of how the estimate keeps up: a group, and
// the range of the child's answers an answer fell in.
type keptUpIn struct {
	group
	answers string
}

// keptUp is how far the answers of a cell came out from what they were
// promised: how many there were, and the differences — a right answer as 1, a
// wrong one as 0, less the chance — summed, all of them and each child's.
type keptUp struct {
	answers int
	sum     float64
	byChild map[string]*ofChild
}

// ofChild is one child's answers in a cell: how many, and their differences
// summed.
type ofChild struct {
	answers int
	sum     float64
}

// mean is the average difference: above zero, the answers came out right more
// often than they were promised.
func (k *keptUp) mean() float64 { return k.sum / float64(k.answers) }

// standardError is how far the mean may be off by chance alone, counted by
// child: the answers of one child lean together, so a cell of a few children
// is as uncertain as a few answers, however many each gave. It needs two
// children, and says so with false when there are fewer.
func (k *keptUp) standardError() (float64, bool) {
	if len(k.byChild) < 2 {
		return 0, false
	}
	mean := k.mean()
	var squares float64
	for _, child := range k.byChild {
		off := child.sum - float64(child.answers)*mean
		squares += off * off
	}
	children := float64(len(k.byChild))
	return math.Sqrt(children/(children-1)*squares) / float64(k.answers), true
}

// answer weighs an answer against the chance its task was handed out at, by
// that chance and by how many answers the child had given, or counts it among
// those the two tables leave out.
func (c *counts) answer(l *line, g group) {
	if why := whyLeftOut(l); why != "" {
		c.answersLeftOut[why]++
		return
	}
	chance, right := *l.Chance, 0.0
	if l.Correct {
		right = 1
	}

	byChance := promisedIn{g, rangeOfChance(chance)}
	promised := c.promises[byChance]
	if promised == nil {
		promised = &cameTrue{children: map[string]bool{}}
		c.promises[byChance] = promised
	}
	promised.answers++
	promised.promised += chance
	if l.Correct {
		promised.right++
	}
	promised.children[l.User] = true

	byAnswers := keptUpIn{g, l.AnswersBucket}
	kept := c.keptUp[byAnswers]
	if kept == nil {
		kept = &keptUp{byChild: map[string]*ofChild{}}
		c.keptUp[byAnswers] = kept
	}
	child := kept.byChild[l.User]
	if child == nil {
		child = &ofChild{}
		kept.byChild[l.User] = child
	}
	kept.answers++
	kept.sum += right - chance
	child.answers++
	child.sum += right - chance
}

// promisesAbout says how to read the table of chances, what the two tables
// leave out, and which way what they leave out leans.
func (c *counts) promisesAbout() string {
	about := "Answers to tasks the rule chose, after the trial series and without the hint, by the chance of a " +
		"right answer each task was handed out at: how many there were, from how many children, the chance " +
		"promised on average, and the share that came out right. A cell of fewer than " +
		strconv.Itoa(fewestAnswers) + " answers says " + tooFew + "."
	var left []string
	for _, why := range leftOutInOrder {
		if n := c.answersLeftOut[why]; n > 0 {
			left = append(left, fmt.Sprintf("%s (%d)", why, n))
		}
	}
	if len(left) > 0 {
		about += " Left out of this table and the next are the answers " + inWords(left) + "."
	}
	unanswered := "tasks left unanswered"
	if c.skipped > 0 {
		unanswered += fmt.Sprintf(" — %d in these lines, of every kind —", c.skipped)
	}
	return about + " Answers given with the hint, and " + unanswered + " are mostly ones that looked too hard, " +
		"so the share right reads a little high by what they take away."
}

// promisesTable is, for each group, the answers by the chance their tasks were
// handed out at: how many, from how many children, the chance promised on
// average, and the share that came out right.
func (c *counts) promisesTable() *table {
	t := &table{columns: []string{"Instructions", "Host", "Chance", "Answers", "Children", "Promised", "Came true"}, named: 3}
	keys := slices.Collect(maps.Keys(c.promises))
	slices.SortFunc(keys, func(a, b promisedIn) int {
		return cmp.Or(c.compareGroups(a.group, b.group), cmp.Compare(a.chance, b.chance))
	})
	for _, key := range keys {
		cell := c.promises[key]
		promised, came := tooFew, tooFew
		if cell.answers >= fewestAnswers {
			promised = hundredths(cell.promised / float64(cell.answers))
			came = hundredths(float64(cell.right) / float64(cell.answers))
		}
		t.add(key.version, key.host, chanceRanges[key.chance].named, number(cell.answers), number(len(cell.children)),
			promised, came)
	}
	return t
}

// keptUpTable is, for each group, the answers by which of the child's answers
// each was: how many, from how many children, how far they came out from their
// promise on average, and the standard error of that mean, counted by child.
func (c *counts) keptUpTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Answer number", "Answers", "Children", "Came true less promised", "Standard error",
	}, named: 3}
	keys := slices.Collect(maps.Keys(c.keptUp))
	slices.SortFunc(keys, func(a, b keptUpIn) int {
		return cmp.Or(c.compareGroups(a.group, b.group), compareAnswerRanges(a.answers, b.answers))
	})
	for _, key := range keys {
		cell := c.keptUp[key]
		mean, standardError := tooFew, tooFew
		if cell.answers >= fewestAnswers {
			mean, standardError = signed(cell.mean()), oneChild
			if byChild, read := cell.standardError(); read {
				standardError = hundredths(byChild)
			}
		}
		t.add(key.version, key.host, key.answers, number(cell.answers), number(len(cell.byChild)), mean, standardError)
	}
	return t
}

// compareAnswerRanges orders two ranges of the child's answers by the first
// answer each holds, read from its name as the service writes it — 6-20
// before 21-50, and 201+ last — so that the order follows the ranges the
// service counts in with no copy of them here; a range named otherwise comes
// after them, by name.
func compareAnswerRanges(a, b string) int {
	first := func(named string) int {
		if n, read := firstAnswerOf(named); read {
			return n
		}
		return math.MaxInt
	}
	return cmp.Or(cmp.Compare(first(a), first(b)), cmp.Compare(a, b))
}

// firstAnswerOf is the first of the child's answers a range holds, read from
// its name as the service writes it: 6 of 6-20, 201 of 201+.
func firstAnswerOf(named string) (int, bool) {
	digits := named[:len(named)-len(strings.TrimLeft(named, "0123456789"))]
	n, err := strconv.Atoi(digits)
	return n, err == nil
}

// lastAnswerOf is the last of the child's answers a range holds, read from
// its name: 20 of 6-20. A range open at its end, such as 201+, has none.
func lastAnswerOf(named string) (int, bool) {
	_, last, found := strings.Cut(named, "-")
	if !found {
		return 0, false
	}
	n, err := strconv.Atoi(last)
	return n, err == nil
}

// hundredths is a share or a chance to two places.
func hundredths(x float64) string { return strconv.FormatFloat(x, 'f', 2, 64) }

// signed is a difference to two places, with its sign: above zero is a child
// who did better than promised.
func signed(x float64) string { return signedTo(x, 2) }

// signedTo is a difference to so many places, with its sign; one that rounds
// to nothing has none.
func signedTo(x float64, places int) string {
	scale := math.Pow(10, float64(places))
	rounded := math.Round(x*scale) / scale
	written := strconv.FormatFloat(rounded, 'f', places, 64)
	switch {
	case rounded > 0:
		return "+" + written
	case rounded < 0:
		return written
	default:
		return strconv.FormatFloat(0, 'f', places, 64)
	}
}

// inWords joins phrases as a sentence lists them: commas between, and "and"
// before the last.
func inWords(phrases []string) string {
	if len(phrases) < 2 {
		return strings.Join(phrases, "")
	}
	return strings.Join(phrases[:len(phrases)-1], ", ") + " and " + phrases[len(phrases)-1]
}
