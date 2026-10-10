package picture

import (
	"maps"
	"slices"
	"strconv"
)

// calendar is the page of a month: the weekday its first day falls on, how
// many days it has, the day its weeks start on, and the days it marks with a
// label. The card writes the weekdays' names in its own language.
type calendar struct {
	noWords
	days  int
	marks []dayMark
}

// dayMark is a label set on a day of the month.
type dayMark struct {
	day   int
	label string
}

func readCalendar(o *object) Picture {
	var c calendar
	o.whole("first", firstWeekday, lastWeekday, true)
	c.days, _ = o.whole("days", fewestDays, mostDays, true)
	o.word("week_starts", "monday", "sunday")
	c.marks = readDayMarks(o, c.days)
	o.close("a calendar")
	return c
}

// readDayMarks reads the labels a calendar sets on its days, keyed by the day,
// each a day of the month where the month's days could be read. A day is
// written one way only, so that no day takes two marks: 1, never 01.
func readDayMarks(o *object, days int) []dayMark {
	entries := o.keyed("marks")
	most := days
	if most == 0 {
		most = mostDays
	}
	var marks []dayMark
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		path := o.at("marks") + ".*"
		day, err := strconv.Atoi(key)
		label, isText := entries[key].(string)
		switch {
		case err != nil || key != strconv.Itoa(day) || day < 1 || day > most:
			o.fault(path, "must be a day of the month: a whole number from 1 to its days, with no nought before it")
		case !isText || !o.decimals.isLabel(label):
			o.fault(path, "%s", labelRule)
		default:
			marks = append(marks, dayMark{day: day, label: label})
		}
	}
	slices.SortFunc(marks, func(a, b dayMark) int { return a.day - b.day })
	return marks
}

func (calendar) Kind() Kind { return Calendar }

func (c calendar) Labels() []string {
	var labels []string
	for _, mark := range c.marks {
		labels = append(labels, mark.label)
	}
	return present(labels...)
}

// Shown are the month's last day, which says how many days it has, the days
// it marks and their labels. The other days are numbered alike, and single
// nothing out, and nor does the weekday the month starts on.
func (c calendar) Shown() []Shown {
	var shown []Shown
	if c.days > 0 {
		shown = append(shown, number(c.days))
	}
	for _, mark := range c.marks {
		shown = append(shown, number(mark.day))
		shown = append(shown, texts(mark.label)...)
	}
	return shown
}
