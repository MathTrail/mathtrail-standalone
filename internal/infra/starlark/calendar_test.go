package starlark_test

import (
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// TestCalendar covers the five helpers a task about dates needs. The expected
// answers are the calendar's, not the code's: a leap year every four except
// every hundred except every four hundred, and days of the week taken from
// dates anybody can check.
func TestCalendar(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		expression string
		want       string
	}{
		{`is_leap(2024)`, `True`},
		{`is_leap(2023)`, `False`},
		{`is_leap(1900)`, `False`},
		{`is_leap(2000)`, `True`},

		{`days_in_month(2023, 2)`, `28`},
		{`days_in_month(2024, 2)`, `29`},
		{`days_in_month(2023, 4)`, `30`},
		{`days_in_month(2023, 12)`, `31`},
		{`days_in_month(2023, 1)`, `31`},

		// The first of January 2024 was a Monday, and Monday is nothing.
		{`weekday(2024, 1, 1)`, `0`},
		{`weekday(2024, 1, 7)`, `6`},
		// The fifth of May 2023 was a Friday.
		{`weekday(2023, 5, 5)`, `4`},
		{`weekday(2024, 2, 29)`, `3`},

		{`add_days((2024, 2, 28), 1)`, `(2024, 2, 29)`},
		{`add_days((2023, 2, 28), 1)`, `(2023, 3, 1)`},
		{`add_days((2023, 1, 1), -1)`, `(2022, 12, 31)`},
		{`add_days((2023, 1, 1), 0)`, `(2023, 1, 1)`},
		{`add_days((2023, 12, 31), 1)`, `(2024, 1, 1)`},
		{`add_days((2024, 1, 1), 366)`, `(2025, 1, 1)`},

		{`days_between((2023, 4, 28), (2023, 5, 5))`, `7`},
		{`days_between((2023, 5, 5), (2023, 4, 28))`, `-7`},
		{`days_between((2023, 1, 1), (2023, 1, 1))`, `0`},
		{`days_between((2023, 1, 1), (2024, 1, 1))`, `365`},
		{`days_between((2024, 1, 1), (2025, 1, 1))`, `366`},
		{`days_between((2023, 6, 25), (2023, 7, 5)) + 1`, `11`},

		// Every one of them by name as well, because a model that writes
		// Python writes by name and half a vocabulary that allowed it would
		// be a rule nobody could hold in mind.
		{`is_leap(year=2024)`, `True`},
		{`days_in_month(year=2024, month=2)`, `29`},
		{`weekday(year=2024, month=1, day=1)`, `0`},
		{`add_days(date=(2023, 1, 1), days=1)`, `(2023, 1, 2)`},
		{`days_between(a=(2023, 1, 1), b=(2023, 1, 8))`, `7`},
	} {
		t.Run(test.expression, func(t *testing.T) {
			t.Parallel()
			if got := repr(t, test.expression); got != test.want {
				t.Errorf("%s = %s, want %s", test.expression, got, test.want)
			}
		})
	}
}

// TestCalendarRefusals holds the calendar to refusing a date that never
// happened. A solver asking about the thirtieth of February has a mistake in
// it, and being told so is worth more than being answered about the first of
// March.
func TestCalendarRefusals(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name       string
		expression string
		contains   string
	}{
		{"a thirteenth month", `days_in_month(2023, 13)`, "a year has twelve"},
		{"a month before the first", `days_in_month(2023, 0)`, "a year has twelve"},
		// A thirteenth month is the January after, and a date that reads as
		// the wrong year is worse than one that is refused.
		{"a thirteenth month in a date", `weekday(2023, 13, 1)`, "a year has twelve"},
		{"a month before the first in a date", `add_days((2023, 0, 1), 1)`, "a year has twelve"},
		{"a day the month does not have", `weekday(2023, 2, 30)`, "has 28"},
		{"the twenty-ninth of a year that has none", `weekday(2023, 2, 29)`, "has 28"},
		{"a day before the first", `weekday(2023, 2, 0)`, "day 0"},
		{"a year before the calendar", `is_leap(0)`, "runs from 1 to 9999"},
		{"a year after it", `is_leap(10000)`, "runs from 1 to 9999"},
		{"a step out of the calendar", `add_days((9999, 12, 31), 1)`, "runs from 1 to 9999"},
		{"further than the calendar runs", `add_days((2023, 1, 1), 5000000)`, "further than this calendar"},
		{"a date of two numbers", `add_days((2023, 1), 1)`, "three whole numbers"},
		{"a date that is not a tuple", `days_between((2023, 1, 1), "yesterday")`, "three whole numbers"},
		{"a date of words", `days_between((2023, 1, 1), ("a", "b", "c"))`, "three whole numbers"},
		{"a year that is not a number", `is_leap("2024")`, "got string"},
		{"a month nobody named", `days_in_month(2023)`, "missing argument for month"},
		{"a day nobody named", `weekday(2023, 1)`, "missing argument for day"},
		{"nothing to add days to", `add_days((2023, 1, 1))`, "missing argument for days"},
		{"only one date to count between", `days_between((2023, 1, 1))`, "missing argument for b"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			result := run(t, "def solve(options):\n    return match(options, "+test.expression+")\n")
			if result.Status != solver.StatusError {
				t.Errorf("status: got %q (%s), want %q", result.Status, result.Message, solver.StatusError)
			}
			if !strings.Contains(result.Message, test.contains) {
				t.Errorf("message: got %q, want it to mention %q", result.Message, test.contains)
			}
		})
	}
}
