package picture

import "slices"

// numberLine is a line of evenly spaced ticks the card numbers itself, from
// one end to the other, with the marks the question names standing on ticks.
type numberLine struct {
	ends  []int
	marks []lineMark
}

// lineMark is a mark on a tick of a number line: where it stands, and its
// label, if it has one.
type lineMark struct {
	at    int
	label string
}

func readNumberLine(o *object) Picture {
	var n numberLine
	from, fromHeld := o.whole("from", lowestTick, highestTick, true)
	to, toHeld := o.whole("to", lowestTick, highestTick, true)
	step, stepHeld := 1, true
	if _, given := o.member("step"); given {
		step, stepHeld = o.whole("step", 1, highestTick-lowestTick, false)
	}
	if fromHeld {
		n.ends = append(n.ends, from)
	}
	if toHeld {
		n.ends = append(n.ends, to)
	}
	ticked := fromHeld && toHeld && stepHeld && ticksMade(o, from, to, step)

	var taken []int
	for _, item := range o.objects("marks", 0, maxTicks, "marks", false) {
		if item == nil {
			continue
		}
		mark := readLineMark(item)
		at, atHeld := item.whole("at", lowestTick, highestTick, true)
		item.close("a mark of a number line")
		switch {
		case !atHeld:
		case ticked && (at < from || at > to || (at-from)%step != 0):
			item.fault(item.at("at"), "must stand on a tick of the line")
		case slices.Contains(taken, at):
			item.fault(item.at("at"), "must stand on a tick no other mark stands on")
		default:
			taken = append(taken, at)
			mark.at = at
			n.marks = append(n.marks, mark)
		}
	}
	o.close("a number line")
	return n
}

// readLineMark reads the label of a mark, if it has one.
func readLineMark(item *object) lineMark {
	label := item.label("label", false)
	return lineMark{label: label}
}

// ticksMade says whether a line's ends and step make as many ticks as a line
// may have, and says what is wrong where they do not.
func ticksMade(o *object, from, to, step int) bool {
	ticks := (to-from)/step + 1
	switch {
	case to <= from:
		o.fault(o.at("to"), "must be greater than from")
	case (to-from)%step != 0:
		o.fault(o.at("step"), "must reach to from from in whole steps")
	case ticks < minTicks || ticks > maxTicks:
		o.fault(o.at("step"), "must make %d to %d ticks from from to to", minTicks, maxTicks)
	default:
		return true
	}
	return false
}

func (numberLine) Kind() Kind { return NumberLine }

func (n numberLine) Labels() []string {
	var labels []string
	for _, mark := range n.marks {
		labels = append(labels, mark.label)
	}
	return present(labels...)
}

// Shown are the line's two ends, since a line may end where the answer is, the
// tick each mark singles out, and the marks' labels. The other ticks are
// numbered alike, and single nothing out.
func (n numberLine) Shown() []Shown {
	var shown []Shown
	for _, end := range n.ends {
		shown = append(shown, number(end))
	}
	for _, mark := range n.marks {
		shown = append(shown, number(mark.at))
		shown = append(shown, texts(mark.label)...)
	}
	return shown
}
