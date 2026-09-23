package profile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// The refusals of Parse. A caller branches on them with errors.Is, because
// what a person is told differs: a file from a newer service is a reason to
// wait, and a broken one is not.
var (
	// ErrMalformed means the bytes are not a profile at all.
	ErrMalformed = errors.New("profile: malformed")
	// ErrNewer means the file was written by a later version of the service.
	// Nothing is touched: two revisions are live during every rollout, and an
	// older instance rewriting a newer file would quietly drop whatever it
	// did not understand.
	ErrNewer = errors.New("profile: written by a newer version")
	// ErrOlder means the file is of an earlier shape than this build reads.
	ErrOlder = errors.New("profile: written by an older version")
	// ErrInvalid means the file is a profile, but not one this service could
	// work with.
	ErrInvalid = errors.New("profile: invalid")
)

// indent is what makes the file readable to the parent who opens it and
// diffable in the revisions Drive keeps.
const indent = "  "

// Parse reads a profile and refuses one the service could not work with.
//
// A field this version does not know is read past and not written back:
// forward compatibility is the refusal of a newer file, not a bag of
// leftovers carried along for ever.
func Parse(raw []byte) (*Profile, error) { return parse(raw, migrations) }

// parse is Parse with the chain of migrations handed in, so that a test can
// walk a chain this shape of the file has never needed without reaching into
// a variable every other test is reading at the same time.
func parse(raw []byte, chain map[int]migration) (*Profile, error) {
	// The version is read first, out of a document that is otherwise left
	// alone: deciding whether this shape can be read at all comes before
	// reading it into this shape.
	var header struct {
		SchemaVersion *int `json:"schema_version"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	if header.SchemaVersion == nil {
		return nil, fmt.Errorf("%w: schema_version is missing", ErrMalformed)
	}
	switch version := *header.SchemaVersion; {
	case version > Version:
		return nil, fmt.Errorf("%w: version %d, this service reads %d", ErrNewer, version, Version)
	case version < Version:
		migrated, err := migrate(raw, version, chain)
		if err != nil {
			return nil, err
		}
		raw = migrated
	}

	var p Profile
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

// Marshal writes the profile as the file it is: indented, with the keys of
// every object in order, and with nothing escaped that a person would then
// have to decode by eye.
func Marshal(p *Profile) ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}

	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetIndent("", indent)
	// Without this an ampersand in a pseudonym or a less-than in the notes is
	// written as & and <. Nothing here is ever put in a web page,
	// and the file is meant to be read.
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(p); err != nil {
		return nil, fmt.Errorf("profile: write: %w", err)
	}
	return out.Bytes(), nil
}
