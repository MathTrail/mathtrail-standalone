package starlark

import (
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"go.starlark.net/starlark"
)

// The calendar helpers answer one another: counting the days back to where a
// move began gives the move, and a week on is the same day of the week. A
// solver about dates is built from these two facts, and a helper that broke
// either would hand it a confident wrong answer.
func TestTheCalendarAnswersItself(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("the days between a date and one moved by n days are n", prop.ForAll(
		func(days []int) bool {
			start, end := days[0], days[1]
			from, apart := dateAfterFirst(start), starlark.MakeInt(end-start)
			moved := call(t, "add_days", addDays, from, apart)
			return same(t, call(t, "days_between", daysBetween, from, moved), apart)
		},
		gen.SliceOfN(2, gen.IntRange(0, 3_000_000)),
	))

	properties.Property("a week on is the same day of the week", prop.ForAll(
		func(start int) bool {
			from := dateAfterFirst(start)
			week, isDate := call(t, "add_days", addDays, from, starlark.MakeInt(7)).(starlark.Tuple)
			return isDate && same(t, call(t, "weekday", weekday, week...), call(t, "weekday", weekday, from...))
		},
		gen.IntRange(0, 3_000_000),
	))

	properties.TestingRun(t)
}

// dateAfterFirst is the date that many days after the first of January of the
// first year, as a solver writes a date.
func dateAfterFirst(days int) starlark.Tuple {
	when := time.Date(firstYear, time.January, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, days)
	return starlark.Tuple{starlark.MakeInt(when.Year()), starlark.MakeInt(int(when.Month())), starlark.MakeInt(when.Day())}
}

// call runs one calendar helper the way a solver calls it, and fails the test
// on a refusal: every date here is inside the calendar.
func call(
	t *testing.T, name string,
	helper func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error),
	args ...starlark.Value,
) starlark.Value {
	t.Helper()

	result, err := starlark.NewBuiltin(name, helper).CallInternal(&starlark.Thread{}, args, nil)
	if err != nil {
		t.Fatalf("%s%v: got error %v, want none", name, args, err)
	}
	return result
}

// same says whether two values are equal as the language compares them.
func same(t *testing.T, a, b starlark.Value) bool {
	t.Helper()

	equal, err := starlark.Equal(a, b)
	if err != nil {
		t.Fatalf("compare %v with %v: got error %v, want none", a, b, err)
	}
	return equal
}
