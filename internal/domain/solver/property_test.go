package solver_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/solver"
)

// The properties here name what the two runs are for. Worked examples show
// that one particular honest program passes and one particular lazy program
// fails; what has to hold is that every honest program passes and no lazy one
// does, whatever the task says and wherever its answer sits.

// follower is the program the contract asks for: it works the answer out —
// here by carrying the text it is looking for — and lets the options say which
// letter that is. Shifting the labels moves its answer with them.
type follower struct{ text string }

func (f follower) Run(_ context.Context, _ string, options solver.Options) (solver.Result, error) {
	return solver.Result{Status: solver.StatusOK, Letters: options.Match(f.text)}, nil
}

// hardCoded is the laziest failure there is: the letter written into the
// program. Shifting the labels leaves it exactly where it was.
type hardCoded struct{ letter string }

func (h hardCoded) Run(_ context.Context, _ string, _ solver.Options) (solver.Result, error) {
	return solver.Result{Status: solver.StatusOK, Letters: []string{h.letter}}, nil
}

// genOptions builds five option texts no two of which mean the same thing,
// which is what a well-formed task offers, and writes them the way tasks write
// them: plainly, with a decimal point, padded, negative, and with a unit —
// which is not a number at all and is compared as words.
//
// One number per option carries both how far it sits from the one before it
// and how it is written, so the whole generator stays a single slice the
// library can shrink as one. The gaps are what is generated rather than the
// numbers, so the five stay distinct through any shrinking; the renderings
// keep them distinct too, because a negative is never a positive and a text
// carrying a unit is never a number.
func genOptions() gopter.Gen {
	const forms = 5
	return gen.SliceOfN(solver.Count, gen.IntRange(0, 40*forms-1)).Map(func(seeds []int) solver.Options {
		var options solver.Options
		running := 0
		for place, seed := range seeds {
			running += seed/forms + 1
			options[place] = written(running, seed%forms)
		}
		return options
	})
}

// written is one number as an option might carry it.
func written(number, form int) string {
	text := strconv.Itoa(number)
	switch form {
	case 1:
		return text + ".0"
	case 2:
		return "  " + text + "  "
	case 3:
		return "-" + text
	case 4:
		return text + " apples"
	default:
		return text
	}
}

func genPlace() gopter.Gen { return gen.IntRange(0, solver.Count-1) }

func TestTheTwoRunsHoldTheirProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("a program that follows the option is agreed with", prop.ForAll(
		func(options solver.Options, place int) bool {
			agreement, err := solver.Verdict(
				context.Background(), follower{text: options[place]}, "irrelevant", options)
			return err == nil && agreement.Letter == solver.Letter(place)
		},
		genOptions(), genPlace(),
	))

	properties.Property("a program that writes the letter out is never agreed with", prop.ForAll(
		func(options solver.Options, place int) bool {
			agreement, err := solver.Verdict(
				context.Background(), hardCoded{letter: solver.Letter(place)}, "irrelevant", options)
			return err == nil && !agreement.Agreed() && agreement.Explain() != ""
		},
		genOptions(), genPlace(),
	))

	properties.Property("the shift leaves no option under its own letter", prop.ForAll(
		func(options solver.Options) bool {
			rotated := options.Rotated()
			for place := range solver.Count {
				if rotated[place] == options[place] {
					return false
				}
			}
			return true
		},
		genOptions(),
	))

	properties.Property("shifting as many times as there are options is no shift at all", prop.ForAll(
		func(options solver.Options) bool {
			back := options
			for range solver.Count {
				back = back.Rotated()
			}
			return back == options
		},
		genOptions(),
	))

	properties.Property("a letter and its text move together", prop.ForAll(
		func(options solver.Options, place int) bool {
			moved := solver.Relabelled(solver.Letter(place))
			return options.Rotated()[solver.Place(moved)] == options[place]
		},
		genOptions(), genPlace(),
	))

	properties.Property("distinct options are matched one at a time", prop.ForAll(
		func(options solver.Options, place int) bool {
			matched := options.Match(options[place])
			return len(matched) == 1 && matched[0] == solver.Letter(place)
		},
		genOptions(), genPlace(),
	))

	properties.TestingRun(t)
}
