package starlark

import (
	"fmt"
	"time"

	"go.starlark.net/starlark"
)

// A date is a tuple of three whole numbers — the year, the month and the day —
// and not a type of its own. There is nothing to learn before writing one, a
// solver can take it apart with an index, and two of them compare the way any
// two tuples do.
//
// The calendar is the one a child is taught: the years below run from the
// first to the four-digit last, and a date that does not exist is refused
// rather than quietly moved to one that does. A solver that asks for the
// thirtieth of February has a bug in it, and saying so is more use than
// answering about the first of March.
const (
	firstYear = 1
	lastYear  = 9999

	// mostDays is more days than lie between the first year and the last, so
	// nothing within it can carry a date out of the calendar by arithmetic
	// alone.
	mostDays = 4_000_000

	hoursADay   = 24
	secondsADay = 60 * 60 * hoursADay
)

// isLeap says whether a year has a twenty-ninth of February.
func isLeap(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (
	starlark.Value, error,
) {
	name := builtin.Name()
	var year int
	if err := starlark.UnpackArgs(name, args, kwargs, "year", &year); err != nil {
		return nil, err
	}
	if err := withinCalendar(name, year); err != nil {
		return nil, err
	}
	return starlark.Bool(year%4 == 0 && (year%100 != 0 || year%400 == 0)), nil
}

// daysInMonth is how many days that month of that year has.
func daysInMonth(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (
	starlark.Value, error,
) {
	name := builtin.Name()
	var year, month int
	if err := starlark.UnpackArgs(name, args, kwargs, "year", &year, "month", &month); err != nil {
		return nil, err
	}
	if err := withinCalendar(name, year); err != nil {
		return nil, err
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("%s: month %d, and a year has twelve", name, month)
	}
	return starlark.MakeInt(lengthOfMonth(year, month)), nil
}

// weekday is the day of the week that date falls on, Monday being nothing and
// Sunday six. Monday first because that is the week a child is taught, and
// counting from nothing because a list is indexed from nothing.
func weekday(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (
	starlark.Value, error,
) {
	name := builtin.Name()
	var year, month, day int
	if err := starlark.UnpackArgs(name, args, kwargs, "year", &year, "month", &month, "day", &day); err != nil {
		return nil, err
	}
	when, err := dayOf(name, year, month, day)
	if err != nil {
		return nil, err
	}
	// The language of the standard library starts its week on Sunday.
	return starlark.MakeInt((int(when.Weekday()) + 6) % 7), nil
}

// addDays is the date that many days after this one, or before it when the
// number is negative.
func addDays(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (
	starlark.Value, error,
) {
	name := builtin.Name()
	var date starlark.Value
	var days int
	if err := starlark.UnpackArgs(name, args, kwargs, "date", &date, "days", &days); err != nil {
		return nil, err
	}
	when, err := dateOf(name, date)
	if err != nil {
		return nil, err
	}
	if days < -mostDays || days > mostDays {
		return nil, fmt.Errorf("%s: %d days, which is further than this calendar runs", name, days)
	}

	moved := when.AddDate(0, 0, days)
	if err := withinCalendar(name, moved.Year()); err != nil {
		return nil, err
	}
	return starlark.Tuple{
		starlark.MakeInt(moved.Year()),
		starlark.MakeInt(int(moved.Month())),
		starlark.MakeInt(moved.Day()),
	}, nil
}

// daysBetween is how many days it is from the first date to the second:
// positive when the second is the later one, negative when it is not.
func daysBetween(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (
	starlark.Value, error,
) {
	name := builtin.Name()
	var from, to starlark.Value
	if err := starlark.UnpackArgs(name, args, kwargs, "a", &from, "b", &to); err != nil {
		return nil, err
	}
	first, err := dateOf(name, from)
	if err != nil {
		return nil, err
	}
	second, err := dateOf(name, to)
	if err != nil {
		return nil, err
	}
	// Seconds rather than the difference of two times: the calendar is longer
	// than a span of nanoseconds can hold, and every date here stands at
	// midnight, so the division is exact.
	return starlark.MakeInt64((second.Unix() - first.Unix()) / secondsADay), nil
}

// dateOf reads a date out of the tuple a solver wrote it as.
func dateOf(name string, value starlark.Value) (time.Time, error) {
	tuple, isTuple := value.(starlark.Tuple)
	if !isTuple || len(tuple) != 3 {
		return time.Time{}, fmt.Errorf("%s: a date is three whole numbers in a tuple, the year, the month and the day", name)
	}
	parts := make([]int, len(tuple))
	for i, part := range tuple {
		number, err := starlark.AsInt32(part)
		if err != nil {
			return time.Time{}, fmt.Errorf("%s: a date is three whole numbers: %w", name, err)
		}
		parts[i] = number
	}
	return dayOf(name, parts[0], parts[1], parts[2])
}

// dayOf is a year, a month and a day as one day of the calendar, and it is
// where a date that does not exist is refused.
func dayOf(name string, year, month, day int) (time.Time, error) {
	if err := withinCalendar(name, year); err != nil {
		return time.Time{}, err
	}
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("%s: month %d, and a year has twelve", name, month)
	}
	if length := lengthOfMonth(year, month); day < 1 || day > length {
		return time.Time{}, fmt.Errorf("%s: day %d, and month %d of %d has %d", name, day, month, year, length)
	}
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// withinCalendar refuses a year this calendar does not run to.
func withinCalendar(name string, year int) error {
	if year < firstYear || year > lastYear {
		return fmt.Errorf("%s: year %d, and this calendar runs from %d to %d", name, year, firstYear, lastYear)
	}
	return nil
}

// lengthOfMonth is the day before the first of the month after, which is the
// last day of this one whether or not the year is a leap year.
func lengthOfMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
