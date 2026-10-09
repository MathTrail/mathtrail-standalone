package picture

import "fmt"

// bars is quantities as bars drawn to one scale — of equal parts, some of them
// shaded, or of segments of their own sizes — with braces under some parts,
// and notes under them all.
type bars struct {
	bars  []bar
	notes []string
	// decimals is the mark the notes write their numbers with.
	decimals Decimals
}

// bar is one bar: its labels, and the parts it is cut into and shades, where
// it is cut into parts.
type bar struct {
	labels []string
	parts  int
	shaded *int
}

func readBars(o *object) Picture {
	b := bars{decimals: o.decimals}
	for _, item := range o.objects("bars", minBars, maxBars, "bars", true) {
		if item != nil {
			b.bars = append(b.bars, readBar(item))
		}
	}
	for i, item := range o.list("notes", 0, maxNotes, "notes", false) {
		note, isText := item.(string)
		if !isText || !o.decimals.isNote(note) {
			o.fault(fmt.Sprintf("%s.%d", o.at("notes"), i), "%s", noteRule)
			continue
		}
		b.notes = append(b.notes, note)
	}
	o.close("bars")
	return b
}

// readBar reads one bar.
func readBar(o *object) bar {
	var read bar
	label := o.label("label", false)
	read.labels = append(read.labels, label)
	partsGiven := holds(o.members["parts"])
	parts, cut := o.whole("parts", minParts, maxParts, false)
	segments := readSegments(o)
	read.labels = append(read.labels, segments...)
	// divisions is how many parts or segments the bar is cut into, 1 for a bar
	// cut into none, and 0 where that could not be read: then no brace can be
	// judged against it.
	divisions := 1
	switch {
	case partsGiven && segments != nil:
		o.fault(o.at("segments"), "come where parts are: a bar is cut into parts or into segments, not both")
		divisions = 0
	case cut:
		read.parts, divisions = parts, parts
	case partsGiven:
		divisions = 0
	case segments != nil:
		divisions = len(segments)
	}
	read.shaded = readShaded(o, partsGiven, cut, parts)
	o.whole("length", minLength, maxLength, false)
	value := o.label("value", false)
	read.labels = append(read.labels, value)
	read.labels = append(read.labels, readBraces(o, divisions)...)
	span := o.label("span", false)
	read.labels = append(read.labels, span)
	o.close("a bar")
	return read
}

// readSegments reads the segments a bar is cut into, each of a size of its
// own, and their labels: nil when the bar has none, and an empty label for a
// segment that has none.
func readSegments(o *object) []string {
	items := o.objects("segments", minParts, maxParts, "segments", false)
	if len(items) == 0 {
		return nil
	}
	labels := make([]string, 0, len(items))
	for _, segment := range items {
		if segment == nil {
			labels = append(labels, "")
			continue
		}
		segment.whole("size", minLength, maxLength, true)
		label := segment.label("label", false)
		segment.close("a segment of a bar")
		labels = append(labels, label)
	}
	return labels
}

// readShaded reads how many parts of a bar are shaded, which only a bar cut
// into parts has, and never more than it has. Where the parts were given and
// could not be read, the shaded parts cannot be judged.
func readShaded(o *object, partsGiven, cut bool, parts int) *int {
	if _, held := o.member("shaded"); !held {
		return nil
	}
	if !partsGiven {
		o.fault(o.at("shaded"), "comes with parts: only a bar cut into parts shades some of them")
		return nil
	}
	if !cut {
		return nil
	}
	shaded, held := o.bounded("shaded", 0, parts, false, "must be a whole number from 0 up to the bar's parts")
	if !held {
		return nil
	}
	return &shaded
}

// readBraces reads the braces under a bar and their labels: a brace spans the
// parts or segments between two of the edges between them, counted from 0 at
// the bar's start, and lies within the bar.
func readBraces(o *object, divisions int) []string {
	var labels []string
	for _, brace := range o.objects("braces", 0, maxBraces, "braces", false) {
		if brace == nil {
			continue
		}
		from, fromHeld := brace.whole("from", 0, maxParts, true)
		to, toHeld := brace.whole("to", 0, maxParts, true)
		label := brace.label("label", true)
		brace.close("a brace")
		if fromHeld && toHeld && divisions > 0 && (from >= to || to > divisions) {
			brace.fault(brace.path, "must lie within its bar: from an edge before to, and to no further than "+
				"the bar's last edge")
		}
		labels = append(labels, label)
	}
	return labels
}

func (bars) Kind() Kind { return Bars }

// Labels are the labels of the bars, of their segments and braces, their
// values and spans, and the labels the notes name.
func (b bars) Labels() []string {
	var labels []string
	for _, each := range b.bars {
		labels = append(labels, each.labels...)
	}
	return present(append(labels, capitalRuns(b.notes)...)...)
}

// Shown are every label of the bars, how many parts each is cut into and
// shades, and the labels and numbers of the notes. A bar's length, a
// segment's size and the edges a brace stands on only lay the picture out.
func (b bars) Shown() []Shown {
	var shown []Shown
	for _, each := range b.bars {
		shown = append(shown, texts(each.labels...)...)
		if each.parts > 0 {
			shown = append(shown, number(each.parts))
		}
		if each.shaded != nil {
			shown = append(shown, number(*each.shaded))
		}
	}
	for _, note := range b.notes {
		shown = append(shown, texts(noteTokens(note, b.decimals)...)...)
	}
	return shown
}
