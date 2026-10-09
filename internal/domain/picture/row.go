package picture

import "strconv"

// row is objects in a row — posts, trees, children in a queue — with a stretch
// cut short where a skip stands, a label under every gap, and a label of the
// whole length.
type row struct {
	items []rowItem
	gaps  string
	span  string
}

// rowItem is one object of a row, or a skip: the label over it and the one
// under it.
type rowItem struct {
	label, below string
	skip         bool
}

func readRow(o *object) Picture {
	var r row
	items := o.objects("items", minRowItems, maxRowItems, "items", true)
	for _, item := range items {
		r.items = append(r.items, readRowItem(item))
	}
	skipsPlaced(o, "items", skipsOf(r.items))
	o.flag("line")
	r.gaps = o.label("gaps", false)
	r.span = o.label("span", false)
	o.whole("copies", 1, maxCopies, false)
	o.close("a row")
	return r
}

// readRowItem reads one item of a row: a skip alone, or an object with its
// labels and its mark. An item that is no object reads as an object with
// nothing on it.
func readRowItem(item *object) rowItem {
	if item == nil {
		return rowItem{}
	}
	skip := item.flag("skip")
	if skip {
		item.alone("is a skip, which holds nothing else")
	}
	read := rowItem{label: item.label("label", false), below: item.label("below", false)}
	item.word("mark", "dot", "ring", "square")
	item.close("an item of a row")
	if skip {
		return rowItem{skip: true}
	}
	return read
}

// skipsOf are which items of a row are skips.
func skipsOf(items []rowItem) []bool {
	skips := make([]bool, len(items))
	for i, item := range items {
		skips[i] = item.skip
	}
	return skips
}

// skipsPlaced says where a skip stands that no skip may: a skip stands for a
// stretch cut short, so it has something on either side of it and is not cut
// short twice running.
func skipsPlaced(o *object, name string, skips []bool) {
	for i, skip := range skips {
		path := o.at(name) + "." + strconv.Itoa(i)
		switch {
		case !skip:
		case i == 0 || i == len(skips)-1:
			o.fault(path, "is a skip, which is never first or last")
		case skips[i-1]:
			o.fault(path, "is a skip next to another skip")
		}
	}
}

func (row) Kind() Kind { return Row }

func (r row) Labels() []string {
	labels := []string{r.gaps, r.span}
	for _, item := range r.items {
		labels = append(labels, item.label, item.below)
	}
	return present(labels...)
}

// Shown are the row's labels, but where the labels over its items count them
// — 1, 2, 3 and on, a skip between them allowed — only the last of them,
// which says how many there are: the others number the places alike.
func (r row) Shown() []Shown {
	counting, last := r.counts(), r.last()
	var shown []string
	for i, item := range r.items {
		if !counting || i == last {
			shown = append(shown, item.label)
		}
		shown = append(shown, item.below)
	}
	return texts(append(shown, r.gaps, r.span)...)
}

// counts says whether the labels over the items count them: every item that
// is not a skip has a label in digits, the first is 1, each next is one more
// than the one before, and a skip lets the count jump ahead.
func (r row) counts() bool {
	previous, afterSkip, seen := 0, false, false
	for _, item := range r.items {
		if item.skip {
			afterSkip = true
			continue
		}
		value, err := strconv.Atoi(item.label)
		if err != nil || !digits(item.label) {
			return false
		}
		switch {
		case !seen && value != 1:
			return false
		case seen && !afterSkip && value != previous+1:
			return false
		case seen && afterSkip && value <= previous:
			return false
		}
		previous, afterSkip, seen = value, false, true
	}
	return seen
}

// last is the place of the last item that is not a skip.
func (r row) last() int {
	for i := len(r.items) - 1; i >= 0; i-- {
		if !r.items[i].skip {
			return i
		}
	}
	return -1
}
