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

// The answers in a topic after it was shown as mastered that a mastery is
// read within: the first few, which a child truly past the topic may still
// slip on, and those to the tenth, which the share taken back is read by.
const (
	TakenBackEarly = 5
	TakenBackBy    = 10
)

// fewestMasteries is how many masteries a share taken back needs before it is
// worth reading, settled each way — taken back, or followed to the end of the
// answers it is read by.
const fewestMasteries = 30

// Followed is a topic a child was shown as mastered, followed through the
// child's answers in that topic after it: how many answers followed it, and
// at which of them it was taken back, none while it was not.
type Followed struct {
	Answers, TakenBackAt int
	wrongInARow          int
}

// Answer follows the mastery through one more answer in its topic, and says
// whether that answer took it back, as the service takes a mastery back: at
// the second wrong answer in a row. An answer is wrong as the service counts
// it, one with the hint or "I don't know" among them, and a right answer, with
// the hint or without, ends the run.
func (f *Followed) Answer(correct bool) bool {
	f.Answers++
	if correct {
		f.wrongInARow = 0
		return false
	}
	f.wrongInARow++
	if f.wrongInARow < profile.MasteryLostAfter {
		return false
	}
	f.TakenBackAt = f.Answers
	return true
}

// shownMastery is a mastery the lines followed, and the child it was shown to.
type shownMastery struct {
	child string
	Followed
}

// following is a mastery the lines are following, and the group it was shown
// in.
type following struct {
	group
	shownMastery
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
		if followed := open[key]; followed != nil && followed.Answer(l.Correct) {
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

// heldOn is, at each answer in a topic after it was shown as mastered, up to
// the tenth, how many masteries were followed to that answer and how many it
// took back.
type heldOn [TakenBackBy]struct{ followed, takenBack int }

// heldOnOf counts the masteries at each answer: a mastery at each answer it
// was followed to, and at the one it was taken back at.
func heldOnOf(masteries []shownMastery) heldOn {
	var on heldOn
	for _, m := range masteries {
		on.add(m)
	}
	return on
}

// add counts a mastery in.
func (on *heldOn) add(m shownMastery) {
	for at := range min(m.Answers, TakenBackBy) {
		on[at].followed++
	}
	if m.TakenBackAt > 0 && m.TakenBackAt <= TakenBackBy {
		on[m.TakenBackAt-1].takenBack++
	}
}

// less is the counts without another's among them.
func (on *heldOn) less(other *heldOn) heldOn {
	without := *on
	for at := range without {
		without[at].followed -= other[at].followed
		without[at].takenBack -= other[at].takenBack
	}
	return without
}

// held is the chance a mastery is still held at the tenth answer in its
// topic, as the masteries counted tell it: over each answer, the share of the
// masteries followed to it that it did not take back, multiplied together. A
// mastery the lines stopped following early counts for the answers it was
// followed through, rather than for none or for all.
func (on *heldOn) held() float64 {
	held := 1.0
	for _, at := range on {
		if at.followed > 0 {
			held *= 1 - float64(at.takenBack)/float64(at.followed)
		}
	}
	return held
}

// settledBy is how many masteries are settled by so many answers in their
// topic: taken back by then, or followed to it.
func settledBy(masteries []shownMastery, answers int) int {
	settled := 0
	for _, m := range masteries {
		if (m.TakenBackAt > 0 && m.TakenBackAt <= answers) || m.Answers >= answers {
			settled++
		}
	}
	return settled
}

// takenBackError is the standard error of the share taken back by the tenth
// answer, counted by child: the share worked out again without each child in
// turn, from every child's counts less that child's own. One child's
// masteries lean together as its answers do. It needs two children, and says
// so with false when there are fewer.
func takenBackError(masteries []shownMastery) (float64, bool) {
	all, byChild := heldOnOf(masteries), map[string]*heldOn{}
	for _, m := range masteries {
		own := byChild[m.child]
		if own == nil {
			own = &heldOn{}
			byChild[m.child] = own
		}
		own.add(m)
	}
	if len(byChild) < 2 {
		return 0, false
	}
	shares := make([]float64, 0, len(byChild))
	for _, child := range slices.Sorted(maps.Keys(byChild)) {
		without := all.less(byChild[child])
		shares = append(shares, 1-without.held())
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
	n := float64(len(shares))
	return math.Sqrt((n - 1) / n * squares), true
}

// takenBackWhen is how many masteries were taken back by the fifth answer in
// their topic, from the sixth to the tenth and after the tenth, and how many
// were not taken back.
func takenBackWhen(masteries []shownMastery) (early, middle, late, held int) {
	for _, m := range masteries {
		switch {
		case m.TakenBackAt == 0:
			held++
		case m.TakenBackAt <= TakenBackEarly:
			early++
		case m.TakenBackAt <= TakenBackBy:
			middle++
		default:
			late++
		}
	}
	return early, middle, late, held
}

// masteriesByGroup are the masteries shown, by group, by every version's hosts
// together and by every version together.
func (c *counts) masteriesByGroup() map[group][]shownMastery {
	byGroup := map[group][]shownMastery{}
	for g, masteries := range c.masteries {
		for _, row := range rowsOf(g) {
			byGroup[row] = append(byGroup[row], masteries...)
		}
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
		"while they followed it. A mastery is settled by the " + ordinal(TakenBackBy) + " answer once it is taken " +
		"back by then or followed to it. The share taken back by the " + ordinal(TakenBackBy) + " answer counts a " +
		"mastery followed for fewer answers for those it was followed for; its standard error is counted by child, " +
		"the share worked out again without each child in turn. A child truly past a topic still slips twice in a " +
		"row now and then, so the share is read beside the learners' bench, never alone. A mastery counts under the " +
		"version of the answer that showed it, and is followed whatever the version of the answers after it. The " +
		"share and its standard error are to three places, and a share that fewer than " +
		strconv.Itoa(fewestMasteries) + " settled masteries stand behind says " + tooFew + ". " + everyHost +
		" takes a version's hosts together, and " + everyVersion + ", last, every version's; the load tool and MCP " +
		"Inspector are left out, as in the table before."
}

// masteriesTable is, for every group, every version's hosts together and every
// version together, the masteries shown, the children shown them, when they
// were taken back, those settled by the tenth answer, and the share taken back
// by then with its standard error.
func (c *counts) masteriesTable() *table {
	t := &table{columns: []string{
		"Instructions", "Host", "Shown", "Children", "Taken back by the " + ordinal(TakenBackEarly) + " answer",
		"From the " + ordinal(TakenBackEarly+1) + " to the " + ordinal(TakenBackBy),
		"After the " + ordinal(TakenBackBy), "Not taken back", "Settled by the " + ordinal(TakenBackBy),
		"Share taken back by the " + ordinal(TakenBackBy), "Standard error",
	}, named: 2}
	byGroup := c.masteriesByGroup()
	for _, g := range slices.SortedFunc(maps.Keys(byGroup), c.compareRows) {
		masteries := byGroup[g]
		children := map[string]bool{}
		for _, m := range masteries {
			children[m.child] = true
		}
		early, middle, late, held := takenBackWhen(masteries)
		settled := settledBy(masteries, TakenBackBy)
		share, standardError := tooFew, tooFew
		if settled >= fewestMasteries {
			on := heldOnOf(masteries)
			share, standardError = thousandths(1-on.held()), oneChild
			if byChild, read := takenBackError(masteries); read {
				standardError = thousandths(byChild)
			}
		}
		t.add(g.version, g.host, number(len(masteries)), number(len(children)), number(early), number(middle),
			number(late), number(held), number(settled), share, standardError)
	}
	return t
}
