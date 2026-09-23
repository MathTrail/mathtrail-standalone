package profile_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/profile"
)

// withField puts one key into the JSON of a fixture, replacing what is there.
// It is how a test says "the same profile, except that" without writing a
// profile out by hand.
func withField(t *testing.T, student, key string, value any) []byte {
	t.Helper()

	var document map[string]json.RawMessage
	if err := json.Unmarshal(readFixture(t, student), &document); err != nil {
		t.Fatalf("parse the fixture: %v", err)
	}
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %v: %v", value, err)
	}
	document[key] = raw

	out, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal the document: %v", err)
	}
	return out
}

// A field this version does not know is read past and not carried along. The
// refusal of a newer file is what protects us from losing something; keeping
// leftovers would only mean writing back what we never understood.
func TestUnknownFieldsAreDroppedRatherThanKept(t *testing.T) {
	t.Parallel()

	raw := withField(t, "dima", "invented_by_a_later_version", map[string]any{"a": 1})
	p, err := profile.Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v, want an unknown field to be read past", err)
	}

	written, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	if strings.Contains(string(written), "invented_by_a_later_version") {
		t.Error("the unknown field was written back")
	}
	if !bytes.Equal(written, readFixture(t, "dima")) {
		t.Error("reading a file with an unknown field changed the rest of it")
	}
}

// Two revisions of the service are live during every rollout. An older one
// must not rewrite what a newer one wrote, because it would drop whatever it
// did not understand — silently, and in the parent's own Drive.
func TestAFileFromAnotherVersionIsRefused(t *testing.T) {
	t.Parallel()

	newer, err := profile.Parse(withField(t, "dima", "schema_version", profile.Version+1))
	if !errors.Is(err, profile.ErrNewer) {
		t.Errorf("Parse() error = %v, want %v", err, profile.ErrNewer)
	}
	if newer != nil {
		t.Error("Parse() returned a profile alongside a refusal")
	}

	older, err := profile.Parse(withField(t, "dima", "schema_version", profile.Version-1))
	if !errors.Is(err, profile.ErrOlder) {
		t.Errorf("Parse() error = %v, want %v", err, profile.ErrOlder)
	}
	if older != nil {
		t.Error("Parse() returned a profile alongside a refusal")
	}
}

func TestWhatIsNotAProfileAtAll(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		raw  string
	}{
		{name: "nothing", raw: ""},
		{name: "not json", raw: "{"},
		{name: "an array", raw: "[]"},
		{name: "an empty object", raw: "{}"},
		{name: "a version that is not a number", raw: `{"schema_version":"1"}`},
		{name: "a moment that is not one", raw: `{"schema_version":1,"created_at":"yesterday"}`},
		{name: "a day that is not one", raw: `{"schema_version":1,"daily":{"date":"2026-13-45"}}`},
		{name: "a day that is a number", raw: `{"schema_version":1,"daily":{"date":5}}`},
		{name: "a moment that is a number", raw: `{"schema_version":1,"created_at":5}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p, err := profile.Parse([]byte(tc.raw))
			if err == nil {
				t.Fatal("Parse() error = nil, want a refusal")
			}
			if !errors.Is(err, profile.ErrMalformed) && !errors.Is(err, profile.ErrInvalid) {
				t.Errorf("Parse() error = %v, want one a caller can branch on", err)
			}
			if p != nil {
				t.Error("Parse() returned a profile alongside a refusal")
			}
		})
	}
}

// The file is opened by the parent and its revisions are shown by Drive, so it
// is written to be read: indented, one key per line, and nothing escaped into
// a number a person would have to decode.
func TestTheFileIsWrittenToBeRead(t *testing.T) {
	t.Parallel()

	p := parseFixture(t, "dima")
	p.Student.Pseudonym = "Tom & Jerry <3"
	p.Student.Notes = "Likes 5 > 3 puzzles."

	written, err := profile.Marshal(p)
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}
	text := string(written)

	for _, want := range []string{`"Tom & Jerry <3"`, `"Likes 5 > 3 puzzles."`} {
		if !strings.Contains(text, want) {
			t.Errorf("the file does not carry %s as written; escaping turned it into something else:\n%s", want, text)
		}
	}
	// Not one escape anywhere in the file: the values above carry none of
	// their own, so a \u here could only have been put in by the writer.
	if strings.Contains(text, `\u`) {
		t.Errorf("the file escapes a character a person would then have to decode:\n%s", text)
	}
	if !strings.HasPrefix(text, "{\n  \"app_version\"") {
		t.Errorf("the file does not start indented and in order:\n%.40s", text)
	}
	if !strings.HasSuffix(text, "}\n") {
		t.Error("the file does not end with a newline")
	}
}

// Every object in the file has its keys in order. It is what makes one
// revision in Drive diff against the next by the fields that changed rather
// than by the order they happened to be written in.
func TestEveryObjectIsWrittenInOrder(t *testing.T) {
	t.Parallel()

	for _, student := range students {
		t.Run(student, func(t *testing.T) {
			t.Parallel()

			checkSorted(t, "", readFixture(t, student))
		})
	}
}

// checkSorted walks a JSON document and fails on the first object whose keys
// are out of order.
func checkSorted(t *testing.T, path string, raw []byte) {
	t.Helper()

	// The keys are read in the order they are written, which a map would lose.
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	token, err := decoder.Token()
	if err != nil {
		t.Fatalf("%s: read: %v", path, err)
	}

	switch delimiter, isDelimiter := token.(json.Delim); {
	case !isDelimiter:
		return // a number, a string, a boolean or null: nothing to order
	case delimiter == '[':
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil {
			t.Fatalf("%s: read the array: %v", path, err)
		}
		for i, item := range items {
			checkSorted(t, path+"["+string(rune('0'+i%10))+"]", item)
		}
	case delimiter == '{':
		checkSortedObject(t, path, decoder)
	}
}

// checkSortedObject walks the keys of one object, in the order they were
// written.
func checkSortedObject(t *testing.T, path string, decoder *json.Decoder) {
	t.Helper()

	previous := ""
	for decoder.More() {
		key, err := decoder.Token()
		if err != nil {
			t.Fatalf("%s: read a key: %v", path, err)
		}
		name, isString := key.(string)
		if !isString {
			t.Fatalf("%s: a key that is not a string", path)
		}
		if name < previous {
			t.Errorf("%s: %q comes after %q, and the keys are written in order", path, name, previous)
		}
		previous = name

		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			t.Fatalf("%s.%s: read the value: %v", path, name, err)
		}
		checkSorted(t, path+"."+name, value)
	}
}

// A profile arrives from the parent's own Drive, where it can be edited by
// hand, restored from an older revision or replaced by anything at all.
// Whatever it says, reading it ends in a profile or a refusal — never in a
// panic that takes the request down.
func FuzzParse(f *testing.F) {
	f.Add(string(readFixture(f2t(f), "dima")))
	f.Add(`{"schema_version":1}`)
	f.Add(`{"schema_version":1,"recent":[{}]}`)
	f.Add(`{"schema_version":1,"topics":{"":{}}}`)
	f.Add(`{"schema_version":99}`)
	f.Add("{")
	f.Add("")

	f.Fuzz(func(t *testing.T, document string) {
		p, err := profile.Parse([]byte(document))
		if err == nil {
			if p == nil {
				t.Fatal("Parse() returned nothing and no error")
			}
			// Anything that parses is a profile this service could work with,
			// which means it can be written back out again.
			if _, wrote := profile.Marshal(p); wrote != nil {
				t.Errorf("a profile that parsed could not be written: %v", wrote)
			}
			return
		}
		switch {
		case errors.Is(err, profile.ErrMalformed), errors.Is(err, profile.ErrInvalid),
			errors.Is(err, profile.ErrNewer), errors.Is(err, profile.ErrOlder):
		default:
			t.Errorf("Parse() error = %v, want one a caller can branch on", err)
		}
		if p != nil {
			t.Error("Parse() returned a profile alongside a refusal")
		}
	})
}

// f2t lets the seed corpus reuse a helper written for a test. A fuzz target's
// seeds are read once, before any fuzzing, and a failure there is a failure of
// the test itself.
func f2t(f *testing.F) *testing.T {
	f.Helper()
	return &testing.T{}
}
