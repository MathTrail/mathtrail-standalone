package checks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// Decode reads what the model handed in — the brief, the task and the
// self-check, each as the JSON it arrived as — and reports everything about
// the format it cannot accept: a part that is missing or is not an object, a
// member the format does not have, a value of the wrong type.
//
// A member the format does not have is refused rather than read past, because
// a misspelt field is one the model believes it filled in. Every such member
// is named, not only the first, so that one more attempt can fix them all.
func Decode(brief, task, selfCheck json.RawMessage) (Draft, []Problem) {
	var (
		draft    Draft
		problems []Problem
	)
	draft.Brief, problems = decodePart[profile.Brief]("brief", brief, problems)
	draft.Task, problems = decodePart[Task]("task", task, problems)
	draft.SelfCheck, problems = decodePart[SelfCheck]("self_check", selfCheck, problems)
	return draft, distinct(problems)
}

// decodePart reads one part of a submission, or reports why it cannot. A part
// with members of the wrong type is still returned, with those members left
// empty, so that what else is wrong with it is found in the same pass.
func decodePart[T any](name string, raw json.RawMessage, problems []Problem) (*T, []Problem) {
	if absent(raw) {
		return nil, append(problems, structural("%s is missing", name))
	}
	var members map[string]json.RawMessage
	if err := json.Unmarshal(raw, &members); err != nil {
		if _, broken := errors.AsType[*json.SyntaxError](err); broken {
			return nil, append(problems, structural("%s is not valid JSON", name))
		}
		return nil, append(problems, structural("%s must be a JSON object", name))
	}

	format := reflect.TypeFor[T]()
	problems = append(problems, strays(raw, format, name)...)

	part := new(T)
	if err := json.Unmarshal(raw, part); err != nil {
		problems = append(problems, mistyped(name, format, err))
	}
	return part, problems
}

// absent is a part that was left out, or sent as nothing.
func absent(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

// strays names every member of raw that the type it is read into has no field
// for, at any depth. A member of an entry keyed by an option's letter is named
// through "*", so that the refusal does not quote the letter. A value that is
// not the shape its type expects has no members to judge: the typed read says
// what is wrong with it, in its own words.
func strays(raw json.RawMessage, format reflect.Type, path string) []Problem {
	for format.Kind() == reflect.Pointer {
		format = format.Elem()
	}
	switch format.Kind() {
	case reflect.Struct:
		return strayMembers(raw, format, path)
	case reflect.Map:
		return strayEntries(raw, format, path)
	case reflect.Slice:
		return strayItems(raw, format, path)
	default:
		return nil
	}
}

// strayMembers judges an object read into a struct: every member is a field of
// it, and every field is judged in turn.
func strayMembers(raw json.RawMessage, format reflect.Type, path string) []Problem {
	var members map[string]json.RawMessage
	if json.Unmarshal(raw, &members) != nil {
		return nil
	}
	fields := jsonFields(format)
	var found []Problem
	for _, name := range slices.Sorted(maps.Keys(members)) {
		field, known := fields[name]
		if !known {
			found = append(found, structural("%s has a member the format does not have: %q", path, name))
			continue
		}
		found = append(found, strays(members[name], field.Type, path+"."+name)...)
	}
	return found
}

// strayEntries judges an object read into a map: its keys are the model's to
// choose, and only what each entry holds is judged.
func strayEntries(raw json.RawMessage, format reflect.Type, path string) []Problem {
	var entries map[string]json.RawMessage
	if json.Unmarshal(raw, &entries) != nil {
		return nil
	}
	var found []Problem
	for _, key := range slices.Sorted(maps.Keys(entries)) {
		found = append(found, strays(entries[key], format.Elem(), path+".*")...)
	}
	return found
}

// strayItems judges a list: each of its items in turn, named by its place.
func strayItems(raw json.RawMessage, format reflect.Type, path string) []Problem {
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) != nil {
		return nil
	}
	var found []Problem
	for i, item := range items {
		found = append(found, strays(item, format.Elem(), fmt.Sprintf("%s.%d", path, i))...)
	}
	return found
}

// jsonFields is a struct's fields by the names they are written under.
func jsonFields(format reflect.Type) map[string]reflect.StructField {
	fields := make(map[string]reflect.StructField, format.NumField())
	for field := range format.Fields() {
		name, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		switch name {
		case "-":
			continue
		case "":
			name = field.Name
		}
		fields[name] = field
	}
	return fields
}

// mistyped turns the first value of the wrong type the JSON library met into
// a refusal that names the field and what it should hold. The library stops
// reporting after the first; the members it could read are read regardless.
func mistyped(name string, format reflect.Type, err error) Problem {
	wrong, isType := errors.AsType[*json.UnmarshalTypeError](err)
	if !isType {
		return structural("%s could not be read", name)
	}
	where := name
	if wrong.Field != "" {
		where = pathThrough(format, name, wrong.Field)
	}
	return structural("%s must be %s", where, kindOf(wrong.Type))
}

// pathThrough rewrites a path the JSON library reported so that it names the
// field without the letter of the option the value sat under.
func pathThrough(format reflect.Type, name, field string) string {
	path := []string{name}
	for segment := range strings.SplitSeq(field, ".") {
		for format != nil && format.Kind() == reflect.Pointer {
			format = format.Elem()
		}
		switch {
		case format == nil:
			path = append(path, segment)
		case format.Kind() == reflect.Map:
			path = append(path, "*")
			format = format.Elem()
		case format.Kind() == reflect.Slice:
			path = append(path, segment)
			format = format.Elem()
		case format.Kind() == reflect.Struct:
			path = append(path, segment)
			field, known := jsonFields(format)[segment]
			format = nil
			if known {
				format = field.Type
			}
		default:
			path = append(path, segment)
			format = nil
		}
	}
	return strings.Join(path, ".")
}

// kindOf says what a field holds, in the words of the format rather than of Go.
func kindOf(format reflect.Type) string {
	for format.Kind() == reflect.Pointer {
		format = format.Elem()
	}
	switch format.Kind() {
	case reflect.String:
		return "a string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "a whole number"
	case reflect.Slice, reflect.Array:
		return "a list"
	case reflect.Map, reflect.Struct:
		return "an object"
	default:
		return "of another type"
	}
}

// distinct drops a problem already reported. The same stray member in four
// explanations is one thing to fix, not four. What it was given is left as it
// was.
func distinct(problems []Problem) []Problem {
	seen := make(map[Problem]bool, len(problems))
	kept := make([]Problem, 0, len(problems))
	for _, problem := range problems {
		if !seen[problem] {
			seen[problem] = true
			kept = append(kept, problem)
		}
	}
	return kept
}
