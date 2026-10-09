package picture

// ring is places in a ring — stations on a loop line, seats round a table —
// which the card numbers from 1 in the order of travel, and the place to
// start at, if the question names one.
type ring struct {
	count int
	start *ringStart
}

// ringStart is the place a ring starts at, and its label, if it has one.
type ringStart struct {
	at    int
	label string
}

func readRing(o *object) Picture {
	var r ring
	r.count, _ = o.whole("count", minRingPlaces, maxRingPlaces, true)
	if start := o.child("start"); start != nil {
		r.start = readRingStart(start, r.count)
	}
	o.close("a ring")
	return r
}

// readRingStart reads where a ring starts: a place of the ring, when its
// count could be read, and a place of a ring of any size otherwise.
func readRingStart(start *object, count int) *ringStart {
	most := count
	if most == 0 {
		most = maxRingPlaces
	}
	at, atHeld := start.bounded("at", 1, most, true, "must be a place of the ring: a whole number from 1 to its count")
	label := start.label("label", false)
	start.close("the start of a ring")
	if !atHeld {
		return &ringStart{label: label}
	}
	return &ringStart{at: at, label: label}
}

func (ring) Kind() Kind { return Ring }

func (r ring) Labels() []string {
	if r.start == nil {
		return nil
	}
	return present(r.start.label)
}

// Shown are the ring's count, the number of its last place, which says how
// many places there are, and the place it starts at with its label. The other
// places are numbered alike, and single nothing out.
func (r ring) Shown() []Shown {
	var shown []Shown
	if r.count > 0 {
		shown = append(shown, number(r.count))
	}
	if r.start != nil {
		if r.start.at > 0 {
			shown = append(shown, number(r.start.at))
		}
		shown = append(shown, texts(r.start.label)...)
	}
	return shown
}
