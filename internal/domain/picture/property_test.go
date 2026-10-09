package picture_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// The properties here name what reading a description holds for any picture,
// not only for the ones written out above. Every description within the
// limits is read without a problem; breaking one of its numbers past its limit
// is one problem, at that number's member; reading the same description twice
// says the same things; and no problem repeats what a member held.

// description is a picture as JSON values, before it is written out.
type description = map[string]any

// genLabel is a label of every form a label takes.
func genLabel() gopter.Gen {
	return gen.OneConstOf("A", "B", "AB", "XYZQW", "?", "7", "12", "-3", "007", "2.5", "99999")
}

// genTime is a time of day, written as the format writes it.
func genTime() gopter.Gen {
	return gopter.CombineGens(gen.IntRange(0, 23), gen.IntRange(0, 59)).Map(func(values []any) string {
		hour, _ := values[0].(int)
		minute, _ := values[1].(int)
		return fmt.Sprintf("%d:%02d", hour, minute)
	})
}

// genLabels are between two numbers of labels.
func genLabels(least, most int) gopter.Gen {
	return gen.IntRange(least, most).FlatMap(func(value any) gopter.Gen {
		count, _ := value.(int)
		return gen.SliceOfN(count, genLabel())
	}, reflect.TypeOf([]string{}))
}

func genClock() gopter.Gen {
	return gopter.CombineGens(genTime(), genLabel()).Map(func(values []any) description {
		return description{"kind": "clock", "time": values[0], "hour_label": values[1]}
	})
}

func genTable() gopter.Gen {
	return gopter.CombineGens(gen.IntRange(1, 8), gen.IntRange(1, 6), genLabel(), genTime()).
		Map(func(values []any) description {
			rows, _ := values[0].(int)
			width, _ := values[1].(int)
			cells := []any{values[2], values[3], "…", ""}
			table := make([]any, rows)
			for i := range table {
				row := make([]any, width)
				for j := range row {
					row[j] = cells[(i+j)%len(cells)]
				}
				table[i] = row
			}
			return description{"kind": "table", "rows": table}
		})
}

func genNumberLine() gopter.Gen {
	return gopter.CombineGens(gen.IntRange(-50, 50), gen.IntRange(1, 15), gen.IntRange(1, 9), genLabel()).
		Map(func(values []any) description {
			from, _ := values[0].(int)
			gaps, _ := values[1].(int)
			step, _ := values[2].(int)
			return description{
				"kind": "number_line", "from": from, "to": from + gaps*step, "step": step,
				"marks": []any{description{"at": from + step, "label": values[3]}, description{"at": from}},
			}
		})
}

func genRow() gopter.Gen {
	return gopter.CombineGens(gen.IntRange(2, 12), genLabel(), gen.Bool(), gen.IntRange(1, 2)).
		Map(func(values []any) description {
			count, _ := values[0].(int)
			cut, _ := values[2].(bool)
			items := make([]any, count)
			for i := range items {
				items[i] = description{"label": values[1], "mark": "square"}
			}
			if cut && count >= 3 {
				items[1] = description{"skip": true}
			}
			return description{"kind": "row", "items": items, "gaps": values[1], "copies": values[3]}
		})
}

func genRing() gopter.Gen {
	return gen.IntRange(3, 24).FlatMap(func(value any) gopter.Gen {
		count, _ := value.(int)
		return gen.IntRange(1, count).Map(func(at int) description {
			return description{"kind": "ring", "count": count, "start": description{"at": at, "label": "S"}}
		})
	}, reflect.TypeOf(description{}))
}

func genGrid() gopter.Gen {
	rows, cols := gen.IntRange(1, 8), gen.IntRange(1, 8)
	return gopter.CombineGens(rows, cols, genLabel()).Map(func(values []any) description {
		rows, _ := values[0].(int)
		cols, _ := values[1].(int)
		rowNames, colNames := make([]any, rows), make([]any, cols)
		for i := range rowNames {
			rowNames[i] = string(rune('A' + i))
		}
		for i := range colNames {
			colNames[i] = fmt.Sprint(i + 1)
		}
		last := fmt.Sprintf("%c%d", 'A'+rows-1, cols)
		return description{
			"kind": "grid", "rows": rowNames, "cols": colNames, "filled": []any{"A1"},
			"marks": description{last: values[2]},
		}
	})
}

func genBars() gopter.Gen {
	return gopter.CombineGens(gen.IntRange(1, 12), gen.IntRange(1, 60), genLabel()).Map(func(values []any) description {
		parts, _ := values[0].(int)
		length, _ := values[1].(int)
		return description{"kind": "bars", "bars": []any{
			description{"label": values[2], "parts": parts, "shaded": parts / 2, "length": length,
				"braces": []any{description{"from": 0, "to": parts, "label": "?"}}},
			description{"segments": []any{description{"size": length, "label": values[2]}}, "value": values[2]},
		}, "notes": []any{"A + B = 12"}}
	})
}

func genVenn() gopter.Gen {
	count, both := genLabel(), genLabel()
	return gopter.CombineGens(count, both).Map(func(values []any) description {
		return description{"kind": "venn", "sets": []any{
			description{"label": "A", "count": values[0]}, description{"label": "B"},
		}, "both": values[1], "neither": "?"}
	})
}

func genBalance() gopter.Gen {
	left, right := genLabels(0, 4), genLabels(0, 4)
	return gopter.CombineGens(left, right).Map(func(values []any) description {
		return description{"kind": "balance", "left": values[0], "right": values[1]}
	})
}

func genContainers() gopter.Gen {
	return gen.IntRange(1, 20).FlatMap(func(value any) gopter.Gen {
		capacity, _ := value.(int)
		return gen.IntRange(0, capacity).Map(func(amount int) description {
			return description{"kind": "containers", "items": []any{
				description{"capacity": capacity, "amount": amount, "label": "A"},
				description{"capacity": 20, "amount": 0},
			}}
		})
	}, reflect.TypeOf(description{}))
}

func genPiles() gopter.Gen {
	return gen.IntRange(3, 40).FlatMap(func(value any) gopter.Gen {
		count, _ := value.(int)
		return gopter.CombineGens(gen.IntRange(2, count-1), gen.IntRange(1, 40)).Map(func(values []any) description {
			return description{"kind": "piles", "piles": []any{
				description{"label": "A", "count": count, "shown": values[0], "group": values[1], "fill": "dark"},
				description{"skip": true},
				description{"value": "?", "shape": "square", "boxed": true},
			}, "across": true}
		})
	}, reflect.TypeOf(description{}))
}

func genCalendar() gopter.Gen {
	return gopter.CombineGens(gen.IntRange(1, 7), gen.IntRange(28, 31), gen.Bool()).Map(func(values []any) description {
		days, _ := values[1].(int)
		sunday, _ := values[2].(bool)
		starts := "monday"
		if sunday {
			starts = "sunday"
		}
		return description{
			"kind": "calendar", "first": values[0], "days": days, "week_starts": starts,
			"marks": description{"1": "A", fmt.Sprint(days): "?"},
		}
	})
}

// genDescription is a description of any kind, within every limit.
func genDescription() gopter.Gen {
	return gen.OneGenOf(genClock(), genTable(), genNumberLine(), genRow(), genRing(), genGrid(), genBars(),
		genVenn(), genBalance(), genContainers(), genPiles(), genCalendar())
}

// written is a description written out as JSON.
func written(value description) json.RawMessage {
	raw, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return raw
}

// wholes are the paths of every whole number a description holds, from the
// picture down.
func wholes(value any, path string) []string {
	var found []string
	switch held := value.(type) {
	case description:
		for name, member := range held {
			found = append(found, wholes(member, path+"."+name)...)
		}
	case []any:
		for i, item := range held {
			found = append(found, wholes(item, fmt.Sprintf("%s.%d", path, i))...)
		}
	case int:
		found = append(found, path)
	}
	slices.Sort(found)
	return found
}

// withWhole is a description with the whole number at a path set to another.
func withWhole(value any, path []string, to int) any {
	switch held := value.(type) {
	case description:
		copied := description{}
		for name, member := range held {
			copied[name] = member
		}
		copied[path[0]] = withWhole(held[path[0]], path[1:], to)
		return copied
	case []any:
		copied := slices.Clone(held)
		var at int
		_, _ = fmt.Sscan(path[0], &at)
		copied[at] = withWhole(held[at], path[1:], to)
		return copied
	default:
		return to
	}
}

// marked is a description whose every label is a text no label may be, and
// which therefore must not be repeated in any problem.
func marked(value any) any {
	switch held := value.(type) {
	case description:
		copied := description{}
		for name, member := range held {
			copied[name] = marked(member)
		}
		return copied
	case []any:
		copied := make([]any, len(held))
		for i, item := range held {
			copied[i] = marked(item)
		}
		return copied
	case []string:
		copied := make([]any, len(held))
		for i, item := range held {
			copied[i] = marked(item)
		}
		return copied
	case string:
		if slices.Contains(picture.Kinds(), picture.Kind(held)) || held == "square" || held == "dark" ||
			held == "monday" || held == "sunday" {
			return held
		}
		return "QXQXQXQ"
	default:
		return held
	}
}

func TestReadingADescriptionHoldsItsProperties(t *testing.T) {
	t.Parallel()

	properties := gopter.NewProperties(nil)

	properties.Property("every description within the limits is accepted", prop.ForAll(
		func(value description) bool {
			read, problems := picture.Parse(written(value), picture.Point)
			return read != nil && len(problems) == 0
		},
		genDescription(),
	))

	properties.Property("a whole number past every limit is one problem, at its member", prop.ForAll(
		func(value description, pick int) bool {
			paths := wholes(value, "picture")
			if len(paths) == 0 {
				return true
			}
			path := paths[pick%len(paths)]
			broken, _ := withWhole(value, strings.Split(path, ".")[1:], 1_000_000).(description)
			_, problems := picture.Parse(written(broken), picture.Point)
			return len(problems) == 1 && problems[0].Path == path
		},
		genDescription(), gen.IntRange(0, 100),
	))

	properties.Property("reading a description twice says the same things", prop.ForAll(
		func(value description) bool {
			raw := written(value)
			first, firstProblems := picture.Parse(raw, picture.Point)
			second, secondProblems := picture.Parse(raw, picture.Point)
			return reflect.DeepEqual(firstProblems, secondProblems) &&
				reflect.DeepEqual(first.Labels(), second.Labels()) && reflect.DeepEqual(first.Shown(), second.Shown())
		},
		genDescription(),
	))

	properties.Property("no problem repeats what a member held", prop.ForAll(
		func(value description) bool {
			broken, _ := marked(value).(description)
			raw := written(broken)
			if !strings.Contains(string(raw), "QXQXQXQ") {
				return true
			}
			_, problems := picture.Parse(raw, picture.Point)
			for _, problem := range problems {
				if strings.Contains(problem.Path+" "+problem.Rule, "QXQXQXQ") {
					return false
				}
			}
			return len(problems) > 0
		},
		genDescription(),
	))

	properties.TestingRun(t)
}
