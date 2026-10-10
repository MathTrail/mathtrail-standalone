package picture

// piles is the piles of a game — stones, coins, counters — each a row of
// counters of one shape and fill, cut short where a count is long, with its
// label before it and a value after it.
type piles struct {
	noWords
	piles []pile
}

// pile is one pile, or a skip: its label, its value, and how many counters it
// has where it says.
type pile struct {
	label, value string
	count        *int
	skip         bool
}

func readPiles(o *object) Picture {
	var p piles
	for _, item := range o.objects("piles", minPiles, maxPiles, "piles", true) {
		p.piles = append(p.piles, readPile(item))
	}
	skips := make([]bool, len(p.piles))
	for i, each := range p.piles {
		skips[i] = each.skip
	}
	skipsPlaced(o, "piles", skips)
	o.flag("box")
	o.flag("across")
	o.close("piles")
	return p
}

// readPile reads one pile: a skip alone, or a pile with its members. A pile
// that shows only some of its counters needs a count to cut short, and shows
// at least one counter on each side of the cut.
func readPile(item *object) pile {
	if item == nil {
		return pile{}
	}
	skip := item.flag("skip")
	if skip {
		item.alone("is a skip, which holds nothing else")
	}
	var read pile
	read.label = item.label("label", false)
	read.value = item.label("value", false)
	countGiven := holds(item.members["count"])
	if count, held := item.whole("count", 0, maxPileCount, false); held {
		read.count = &count
	}
	readShown(item, countGiven, read.count)
	item.whole("group", minGroup, maxGroup, false)
	item.word("fill", "dark", "light")
	item.word("shape", "dot", "square")
	item.flag("boxed")
	item.close("a pile")
	if skip {
		return pile{skip: true}
	}
	return read
}

// readShown reads how many counters a pile cut short shows. Where the count
// was given and could not be read, what it shows cannot be judged.
func readShown(item *object, countGiven bool, count *int) {
	_, held := item.member("shown")
	switch {
	case !held:
	case !countGiven:
		item.fault(item.at("shown"), "comes with count: only a pile of a stated size is cut short")
	case count == nil:
	default:
		item.bounded("shown", fewestShown, *count-1, false,
			"must be a whole number from 2 to one less than the pile's count")
	}
}

func (piles) Kind() Kind { return Piles }

func (p piles) Labels() []string {
	var labels []string
	for _, each := range p.piles {
		labels = append(labels, each.label, each.value)
	}
	return present(labels...)
}

// Shown are each pile's label, its value and how many counters it has. How
// many it shows of them, and in what groups, only lay the picture out.
func (p piles) Shown() []Shown {
	var shown []Shown
	for _, each := range p.piles {
		shown = append(shown, texts(each.label, each.value)...)
		if each.count != nil {
			shown = append(shown, number(*each.count))
		}
	}
	return shown
}
