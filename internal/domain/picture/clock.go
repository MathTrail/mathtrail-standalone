package picture

// clock is a clock face: the time its hands show, and the labels at their tips
// where the question calls the hands so.
type clock struct {
	noWords
	time        *Time
	hourLabel   string
	minuteLabel string
}

func readClock(o *object) Picture {
	var c clock
	if read, isTime := o.time("time", true); isTime {
		c.time = &read
	}
	c.hourLabel = o.label("hour_label", false)
	c.minuteLabel = o.label("minute_label", false)
	o.close("a clock")
	return c
}

func (clock) Kind() Kind { return Clock }

func (c clock) Labels() []string { return present(c.hourLabel, c.minuteLabel) }

// Shown are the labels at the tips of the hands and the time the hands show,
// which a child reads off the face as a twelve-hour face shows it.
func (c clock) Shown() []Shown {
	shown := texts(c.hourLabel, c.minuteLabel)
	if c.time != nil {
		face := c.time.OnTheFace()
		shown = append(shown, Shown{Face: &face})
	}
	return shown
}
