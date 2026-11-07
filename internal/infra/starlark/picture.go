package starlark

import (
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"slices"
	"strconv"
	"strings"

	"go.starlark.net/starlark"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/picture"
)

// pictureOf is the description of a task's picture as a program is handed it,
// the second argument of a solve that takes two, and what notes the members
// the program reads of it. A task with no picture hands it None, and so does
// a description that is no object: the format's check refuses that one, and
// a program is not where a fault of the format is told.
//
// It is read from the description as it came, with no decimal mark, since a
// program runs before the lesson's language is known: a number written as a
// number reaches it as a number, and a label as a text. Reading it never
// fails, whatever the model wrote, so that a description out of the format
// costs the task a refusal rather than the service a failure.
func pictureOf(description json.RawMessage) (starlark.Value, *drawn) {
	members, isObject := picture.Members(description)
	if !isObject {
		return starlark.None, nil
	}
	drawing := &drawn{
		names:   slices.Sorted(maps.Keys(members)),
		members: make(map[string]starlark.Value, len(members)),
		taken:   map[string]bool{},
	}
	for name, member := range members {
		drawing.members[name] = valueOf(member)
	}
	return drawing, drawing
}

// valueOf is one value inside a picture, frozen: an object becomes a
// dictionary, a list holding an object a list, and any other list a tuple, so
// that a cell [2, 3] reads as (2, 3) and can stand in a set. An object's
// members come in the order of their names, since the order they are written
// in means nothing in JSON, and a program that needs another order sorts them.
func valueOf(value any) starlark.Value {
	switch typed := value.(type) {
	case bool:
		return starlark.Bool(typed)
	case string:
		return starlark.String(typed)
	case json.Number:
		return numberOf(typed)
	case map[string]any:
		dict := starlark.NewDict(len(typed))
		for _, name := range slices.Sorted(maps.Keys(typed)) {
			// A text never fails to be a key.
			_ = dict.SetKey(starlark.String(name), valueOf(typed[name]))
		}
		dict.Freeze()
		return dict
	case []any:
		items := make([]starlark.Value, 0, len(typed))
		holdsObject := false
		for _, item := range typed {
			_, isObject := item.(map[string]any)
			holdsObject = holdsObject || isObject
			items = append(items, valueOf(item))
		}
		if !holdsObject {
			return starlark.Tuple(items)
		}
		list := starlark.NewList(items)
		list.Freeze()
		return list
	default:
		return starlark.None
	}
}

// numberOf is a number as a program reads it: a whole one as an integer, 3.0
// included, as the format reads a whole number; any other as a float, and one
// too large for a float as an infinity rather than a failure.
func numberOf(number json.Number) starlark.Value {
	if whole, err := strconv.ParseInt(number.String(), 10, 64); err == nil {
		return starlark.MakeInt64(whole)
	}
	float, _ := strconv.ParseFloat(number.String(), 64)
	if float == math.Trunc(float) && float >= math.MinInt64 && float < math.MaxInt64 {
		return starlark.MakeInt64(int64(float))
	}
	return starlark.Float(float)
}

// drawn is the description's own object as a program reads it: a frozen
// dictionary that notes each member the program takes out of it, whether by
// its name, with get, or with items, values or dict(). Asking whether a member
// is there, or listing the names, takes nothing out of it and notes nothing.
// Its members come in the order of their names.
//
// The members are noted at the top of the description alone. What a program
// does with a member once it holds it is its own affair: a member is read when
// anything in it is.
type drawn struct {
	names   []string
	members map[string]starlark.Value
	taken   map[string]bool
}

var (
	_ starlark.IterableMapping = (*drawn)(nil)
	_ starlark.Sequence        = (*drawn)(nil)
	_ starlark.HasAttrs        = (*drawn)(nil)
	_ starlark.Container       = (*drawn)(nil)
)

// read are the members the program took out of the picture, sorted, and none
// for a run handed no picture.
func (d *drawn) read() []string {
	if d == nil {
		return nil
	}
	var read []string
	for _, name := range d.names {
		if d.taken[name] {
			read = append(read, name)
		}
	}
	return read
}

// take is a member, noted as read.
func (d *drawn) take(name string) starlark.Value {
	d.taken[name] = true
	return d.members[name]
}

// String writes the picture out without reading it: the interpreter writes a
// value into the message of an error, and an error is no reading.
func (d *drawn) String() string {
	parts := make([]string, 0, len(d.names))
	for _, name := range d.names {
		parts = append(parts, fmt.Sprintf("%q: %s", name, d.members[name].String()))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// Type is a name of its own rather than a dictionary's: the interpreter takes
// two values whose types share a name to be of one kind, and would compare a
// dictionary of the program's with the picture as two dictionaries, which
// fails inside the interpreter rather than in the program.
func (*drawn) Type() string { return "picture" }

// Freeze has nothing to do: the picture was frozen when it was built.
func (*drawn) Freeze() {}

// Truth is whether the picture has any member.
func (d *drawn) Truth() starlark.Bool { return len(d.names) > 0 }

// Hash refuses, as a dictionary's does.
func (*drawn) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: dict") }

// Len is how many members the picture has, which the capped built-ins need to
// bound a walk over it.
func (d *drawn) Len() int { return len(d.names) }

// Get takes a member out by its name.
func (d *drawn) Get(key starlark.Value) (starlark.Value, bool, error) {
	name, isText := key.(starlark.String)
	if !isText {
		return nil, false, nil
	}
	if _, held := d.members[string(name)]; !held {
		return nil, false, nil
	}
	return d.take(string(name)), true, nil
}

// Has says whether a member is there, and takes nothing out.
func (d *drawn) Has(key starlark.Value) (bool, error) {
	name, isText := key.(starlark.String)
	if !isText {
		return false, nil
	}
	_, held := d.members[string(name)]
	return held, nil
}

// Iterate walks the names of the members, and takes none of them out.
func (d *drawn) Iterate() starlark.Iterator { return d.keys().Iterate() }

// Items takes every member out.
func (d *drawn) Items() []starlark.Tuple {
	items := make([]starlark.Tuple, 0, len(d.names))
	for _, name := range d.names {
		items = append(items, starlark.Tuple{starlark.String(name), d.take(name)})
	}
	return items
}

// keys are the names of the members.
func (d *drawn) keys() starlark.Tuple {
	names := make(starlark.Tuple, 0, len(d.names))
	for _, name := range d.names {
		names = append(names, starlark.String(name))
	}
	return names
}

// Attr is the few methods a dictionary of the picture has: get, items, keys
// and values. None of them edits it.
func (d *drawn) Attr(name string) (starlark.Value, error) {
	switch name {
	case "get":
		return starlark.NewBuiltin(name, d.get), nil
	case "items":
		return starlark.NewBuiltin(name, d.itemsOf), nil
	case "keys":
		return starlark.NewBuiltin(name, d.keysOf), nil
	case "values":
		return starlark.NewBuiltin(name, d.valuesOf), nil
	}
	return nil, nil
}

// AttrNames are the methods Attr has.
func (*drawn) AttrNames() []string { return []string{"get", "items", "keys", "values"} }

// get is a member by its name, or the default when the picture has none of
// that name, which takes nothing out.
func (d *drawn) get(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key starlark.Value
	var fallback starlark.Value = starlark.None
	if err := starlark.UnpackPositionalArgs(builtin.Name(), args, kwargs, 1, &key, &fallback); err != nil {
		return nil, err
	}
	member, found, err := d.Get(key)
	if err != nil || !found {
		return fallback, err
	}
	return member, nil
}

// itemsOf is items(): every member with its name, each taken out.
func (d *drawn) itemsOf(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(builtin.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	items := d.Items()
	values := make([]starlark.Value, len(items))
	for i, item := range items {
		values[i] = item
	}
	return starlark.NewList(values), nil
}

// keysOf is keys(): the names, with nothing taken out.
func (d *drawn) keysOf(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(builtin.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	return starlark.NewList(d.keys()), nil
}

// valuesOf is values(): every member, each taken out.
func (d *drawn) valuesOf(_ *starlark.Thread, builtin *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(builtin.Name(), args, kwargs, 0); err != nil {
		return nil, err
	}
	values := make([]starlark.Value, 0, len(d.names))
	for _, name := range d.names {
		values = append(values, d.take(name))
	}
	return starlark.NewList(values), nil
}
