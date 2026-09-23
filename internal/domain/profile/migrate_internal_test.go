package profile

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// The chain is tested with a step that does not exist in the product, because
// there has only ever been one shape of the file. What is being tested is the
// walk: that an older file is carried up one version at a time, that a version
// nobody can read is refused rather than guessed at, and that a step which
// fails stops the read instead of leaving the file half migrated.
//
// The first real migration replaces this step with its own, and brings a file
// of the old shape with it.

// olderProfile is a file of the shape before this one: the same profile with
// the pseudonym under a name it no longer has.
func olderProfile(t *testing.T) []byte {
	t.Helper()

	raw, err := Marshal(sampleProfile(t))
	if err != nil {
		t.Fatalf("Marshal() error = %v, want nil", err)
	}

	var document map[string]json.RawMessage
	if read := json.Unmarshal(raw, &document); read != nil {
		t.Fatalf("read the document: %v", read)
	}
	var student map[string]json.RawMessage
	if read := json.Unmarshal(document["student"], &student); read != nil {
		t.Fatalf("read the student: %v", read)
	}
	student["nickname"] = student["pseudonym"]
	delete(student, "pseudonym")

	document["student"], _ = json.Marshal(student)
	document["schema_version"] = json.RawMessage("0")
	older, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("write the document: %v", err)
	}
	return older
}

// renameNickname is what a migration looks like: one field, one version, and
// no knowledge of any other step.
func renameNickname(document map[string]json.RawMessage) error {
	var student map[string]json.RawMessage
	if err := json.Unmarshal(document["student"], &student); err != nil {
		return err
	}
	student["pseudonym"] = student["nickname"]
	delete(student, "nickname")

	rewritten, err := json.Marshal(student)
	if err != nil {
		return err
	}
	document["student"] = rewritten
	return nil
}

func TestAnOlderFileIsCarriedUp(t *testing.T) {
	t.Parallel()

	p, err := parse(olderProfile(t), map[int]migration{0: renameNickname})
	if err != nil {
		t.Fatalf("Parse() error = %v, want an older file to be migrated", err)
	}
	if p.Student.Pseudonym != "Otter" {
		t.Errorf("pseudonym = %q, want the migration to have moved it", p.Student.Pseudonym)
	}
	// The next write stores the shape it was migrated to, so a file is
	// migrated by being used rather than by anybody going looking for it.
	if p.SchemaVersion != Version {
		t.Errorf("schema_version = %d, want %d", p.SchemaVersion, Version)
	}
}

func TestAVersionWithNoWayUpIsRefused(t *testing.T) {
	t.Parallel()

	p, err := parse(olderProfile(t), map[int]migration{})
	if !errors.Is(err, ErrOlder) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrOlder)
	}
	if p != nil {
		t.Error("Parse() returned a profile alongside a refusal")
	}
}

func TestAMigrationThatFailsStopsTheRead(t *testing.T) {
	t.Parallel()

	_, err := parse(olderProfile(t), map[int]migration{
		0: func(map[string]json.RawMessage) error { return errors.New("the field is not there") },
	})
	if !errors.Is(err, ErrOlder) {
		t.Fatalf("Parse() error = %v, want %v", err, ErrOlder)
	}
	if !strings.Contains(err.Error(), "the field is not there") {
		t.Errorf("Parse() said %q, want it to carry what the step said", err)
	}
}

// The chain the product ships is empty, and that is the whole statement: there
// has been one shape of this file. A step added without a fixture beside it
// fails here.
func TestTheProductShipsTheChainItHas(t *testing.T) {
	t.Parallel()

	if len(migrations) != Version-1 {
		t.Errorf("the chain holds %d steps for %d versions of the file; every version but the first needs one",
			len(migrations), Version)
	}
}

// A step that leaves the document in a state nobody can write out stops the
// read, rather than producing a file that cannot be read back.
func TestAMigrationThatBreaksTheDocumentStopsTheRead(t *testing.T) {
	t.Parallel()

	_, err := parse(olderProfile(t), map[int]migration{
		0: func(document map[string]json.RawMessage) error {
			document["student"] = json.RawMessage(`{"pseudonym":`)
			return nil
		},
	})
	if !errors.Is(err, ErrMalformed) {
		t.Fatalf("parse() error = %v, want %v", err, ErrMalformed)
	}
}
