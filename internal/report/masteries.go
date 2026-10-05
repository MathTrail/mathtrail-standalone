package report

import (
	"cmp"
	"maps"
	"math"
	"slices"
	"strconv"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// A topic a child was shown as mastered is followed through the child's
// answers in that topic after it, and is taken back as the service takes it
// back: at the second wrong answer in a row there. No line says that a topic
// was taken back, and the count of topics mastered a line carries also falls
// when the child's level moves a topic's tasks up, so the answers themselves
// are what is followed.

// The answers in a topic a mastery is read within: the first few, which a
// child truly past the topic may still slip on, and those to the tenth, which
// the share taken back is read by.
const (
	takenBackEarly = 5
	takenBackBy    = 10
)

// fewestMasteries is how many masteries a share taken back needs before it is
// worth reading, settled each way — taken back, or followed to the end of the
// answers it is read by.
const fewestMasteries = 30

// shownMastery is a topic a child was shown as mastered, as the lines follow
// it: the child, how many of the child's answers in the topic the lines
// followed after it, and at which of them it was taken back, none when it was
// not taken back while the lines followed it.
type shownMastery struct {
	child       string
	followed    int
	takenBackAt int
}

// following is a mastery the lines are following: the group it was shown in,
// the mastery so far, and the run of wrong answers in its topic.
type following struct {
	group
	shownMastery
	wrongInARow int
}

// answer follows the mastery through one more answer in its topic, and says
// whether that answer took it back. Every answer counts as the service counts
// it: one with the hint, or "I don't know", which is a wrong answer, among
// them.
func (f *following) answer(correct bool) bool {
	f.followed++
	if correct {
		f.wrongInARow = 0
		return false
	}
	f.wrongInARow++
	if f.wrongInARow < profile.MasteryLostAfter {
		return false
	}
	f.takenBackAt = f.followed
	return true
}

// shownIn is the answer a mastery was shown on: the call it came in, the
// child and the topic.
type shownIn struct{ call, user, topic string }

// childTopic is a child's topic, which one mastery at a time is followed in.
type childTopic struct{ user, topic string }

// callsThatShowed are the answers a topic was shown as mastered on: the call
// of every line of a topic mastered, with its child and topic. A mastery is
// read with the answer that earned it, which shares its call, so that no tie
// of their moments needs settling.
func callsThatShowed(lines []line) map[shownIn]bool {
	shown := map[shownIn]bool{}
	for i := range lines {
		if l := &lines[i]; l.Message == eventTopicMastered && !l.repeat && l.call() != "" && l.User != "" && l.Topic != "" {
			shown[shownIn{l.call(), l.User, l.Topic}] = true
		}
	}
	return shown
}

// answersInTimeOrder are the lines of answers a mastery can be followed
// through, in the order they were given: the log hands them over newest first.
// A line the log repeated is followed once, and one with no moment, no child
// or no topic cannot be placed and is not followed.
func answersInTimeOrder(lines []line) []*line {
	var answers []*line
	for i := range lines {
		if l := &lines[i]; l.Message == eventAnswerRecorded && !l.repeat && l.User != "" && l.Topic != "" && !l.Time.IsZero() {
			answers = append(answers, l)
		}
	}
	slices.SortStableFunc(answers, func(a, b *line) int { return a.Time.Compare(b.Time) })
	return answers
}

// shownMasteries are the masteries the children were shown, by the group of
// the answer each was shown on, followed through the answers in their topics.
// The answer a mastery was shown on is not one of those it is followed
// through. A topic shown as mastered again, at a higher level, ends the
// mastery before it, which was not taken back; so does the end of the lines.
// A host that hands a task to no child a family has, and a child such a host
// handed a task to, are left out.
func shownMasteries(lines []line, hosts map[string]string, testing map[string]bool) map[group][]shownMastery {
	shown := callsThatShowed(lines)
	open := map[childTopic]*following{}
	done := map[group][]shownMastery{}
	end := func(key childTopic) {
		done[open[key].group] = append(done[open[key].group], open[key].shownMastery)
		delete(open, key)
	}
	for _, l := range answersInTimeOrder(lines) {
		key := childTopic{l.User, l.Topic}
		if followed := open[key]; followed != nil && followed.answer(l.Correct) {
			end(key)
		}
		if !shown[shownIn{l.call(), l.User, l.Topic}] {
			continue
		}
		if open[key] != nil {
			end(key)
		}
		host := hostOf(hosts, l.call())
		if slices.Contains(notChildren, host) || testing[l.User] {
			continue
		}
		open[key] = &following{group: group{l.InstructionsVersion, host}, shownMastery: shownMastery{child: l.User}}
	}
	for _, key := range slices.SortedFunc(maps.Keys(open), compareChildTopics) {
		end(key)
	}
	return done
}

// compareChildTopics orders a child's topics by child and then by topic.
func compareChildTopics(a, b childTopic) int {
	return cmp.Or(cmp.Compare(a.user, b.user), cmp.Compare(a.topic, b.topic))
}

// testingUsers are the children a host that hands a task to no child a family
// has handed a task to: whatever else such a child did is no child's either.
func testingUsers(lines []line) map[string]bool {
	testing := map[string]bool{}
	for i := range lines {
		if l := &lines[i]; l.Message == eventTaskAccepted && l.User != "" && slices.Contains(notChildren, l.Host) {
			testing[l.User] = true
		}
	}
	return testing
}

// heldThrough is the chance a mastery is still held after so many answers in
// its topic, as the masteries followed tell it: over each answer up to that
// one, the share of the masteries followed to it that it did not take back,
// multiplied together. A mastery the lines stopped following early counts for
// the answers it was followed through, rather than for none or for all.
func heldThrough(masteries []shownMastery, answers int) float64 {
	held := 1.0
	for at := 1; at <= answers; at++ {
		followed, takenBack := 0, 0
		for _, m := range masteries {
			if m.followed >= at {
				followed++
			}
			if m.takenBackAt == at {
				takenBack++
			}
		}
		if followed > 0 {
			held *= 1 - float64(takenBack)/float64(followed)
		}
	}
	return held
}

// settledBy is how many masteries are settled by so many answers in their
// topic: taken back by then, or followed to it.
func settledBy(masteries []shownMastery, answers int) int {
	settled := 0
	for _, m := range masteries {
		if (m.takenBackAt > 0 && m.takenBackAt <= answers) || m.followed >= answers {
			settled++
		}
	}
	return settled
}

// takenBackError is the standard error of the share taken back by the tenth
// answer, counted by child: the share worked out again without each child in
// turn. One child's masteries lean together as its answers do. It needs two
// children, and says so with false when there are fewer.
func takenBackError(masteries []shownMastery) (float64, bool) {
	var children []string
	for _, m := range masteries {
		children = append(children, m.child)
	}
	slices.Sort(children)
	children = slices.Compact(children)
	if len(children) < 2 {
		return 0, false
	}
	shares := make([]float64, 0, len(children))
	for _, child := range children {
		without := slices.DeleteFunc(slices.Clone(masteries), func(m shownMastery) bool { return m.child == child })
		shares = append(shares, 1-heldThrough(without, takenBackBy))
	}
	var mean float64
	for _, share := range shares {
		mean += share
	}
	mean /= float64(len(shares))
	var squares float64
	for _, share := range shares {
		squares += (share - mean) * (share - mean)
	}
	n := float64(len(children))
	return math.Sqrt((n - 1) / n * squares), true
}

// masteriesByGroup are the masteries shown, by group and by every version's
// hosts together.
func (c *counts) masteriesByGroup() map[group][]shownMastery {
	byGroup := map[group][]shownMastery{}
	for _, g := range slices.SortedFunc(maps.Keys(c.masteries), c.compareRows) {
		together := group{version: g.version, host: everyHost}
		byGroup[g] = append(byGroup[g], c.masteries[g]...)
		byGroup[together] = append(byGroup[together], c.masteries[g]...)
	}
	return byGroup
}

// masteriesAbout says how to read the table of masteries taken back.
func (c *counts) masteriesAbout() string {
	return "Every topic a child was shown as mastered — a line of a topic mastered, read with the answer that earned " +
		"it —, followed through the child's answers in that topic after it, whoever chose the task, those with the " +
		"hint and \"I don't know\" among them. A mastery is taken back at the second wrong answer in a row there, " +
		"as the service takes it back, and a right answer between ends the run; one the lines stop following " +
		"first — at their end, or at the topic shown as mastered again, at a higher level — was not taken back " +
		"while they followed it. The share taken back by the " + strconv.Itoa(takenBackBy) + "th answer counts a " +
		"mastery followed for fewer answers for those it was followed for; its standard error is counted by child, " +
		"the share worked out again without each child in turn. A child truly past a topic still slips twice in a " +
		"row now and then, so the share is read beside the learners' bench, never alone. A share that fewer than " +
		strconv.Itoa(fewestMasteries) + " masteries settle — taken back by the " + strconv.Itoa(takenBackBy) +
		"th answer, or followed to it — says " + tooFew + ". " + everyHost + " takes a version's hosts together, " +
		"and the load tool and MCP Inspector are left out, as in the table before."
}

// masteriesTable is, for every group and every version's hosts together, the
// masteries shown, the children shown them, when they were taken back, and
// the share taken back by the tenth answer with its standard error.
func (c *counts) masteriesTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Shown", "Children", "Taken back by the " + strconv.Itoa(takenBackEarly) + "th answer",
		"From the " + strconv.Itoa(takenBackEarly+1) + "th to the " + strconv.Itoa(takenBackBy) + "th",
		"After the " + strconv.Itoa(takenBackBy) + "th", "Not taken back",
		"Share taken back by the " + strconv.Itoa(takenBackBy) + "th", "Standard error",
	}, named: 2}
	byGroup := c.masteriesByGroup()
	for _, g := range slices.SortedFunc(maps.Keys(byGroup), c.compareRows) {
		masteries := byGroup[g]
		children := map[string]bool{}
		var early, middle, late, held int
		for _, m := range masteries {
			children[m.child] = true
			switch {
			case m.takenBackAt == 0:
				held++
			case m.takenBackAt <= takenBackEarly:
				early++
			case m.takenBackAt <= takenBackBy:
				middle++
			default:
				late++
			}
		}
		share, standardError := tooFew, tooFew
		if settledBy(masteries, takenBackBy) >= fewestMasteries {
			share, standardError = hundredths(1-heldThrough(masteries, takenBackBy)), oneChild
			if byChild, read := takenBackError(masteries); read {
				standardError = hundredths(byChild)
			}
		}
		t.add(g.version, g.host, number(len(masteries)), number(len(children)), number(early), number(middle),
			number(late), number(held), share, standardError)
	}
	return t
}
