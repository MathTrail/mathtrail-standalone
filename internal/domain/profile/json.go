package profile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"
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

// NewerError is a file a later version of the service wrote: which version,
// and when the file says it was last written. A rollout is over in minutes, so
// the moment is what tells a rollout under way from a file edited by hand or
// a newer version withdrawn. It is zero when the file does not say, in a form
// this build reads. It is ErrNewer to errors.Is.
type NewerError struct {
	Version   int
	WrittenAt time.Time
}

// Error says which version wrote the file, and which this build reads.
func (e *NewerError) Error() string {
	return fmt.Sprintf("%s: version %d, this service reads %d", ErrNewer, e.Version, Version)
}

// Is makes a newer file ErrNewer to errors.Is.
func (e *NewerError) Is(target error) bool { return target == ErrNewer }

// writtenAt is the moment an updated_at says, or zero when it says none this
// build reads.
func writtenAt(raw json.RawMessage) time.Time {
	var at Time
	if len(raw) == 0 || json.Unmarshal(raw, &at) != nil {
		return time.Time{}
	}
	return at.Time
}

// parse is Parse with the chain of migrations handed in, so that a test can
// walk a chain this shape of the file has never needed without reaching into
// a variable every other test is reading at the same time.
func parse(raw []byte, chain map[int]migration) (*Profile, error) {
	// The version is read first, out of a document that is otherwise left
	// alone: deciding whether this shape can be read at all comes before
	// reading it into this shape.
	var header struct {
		SchemaVersion *int            `json:"schema_version"`
		UpdatedAt     json.RawMessage `json:"updated_at"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	if header.SchemaVersion == nil {
		return nil, fmt.Errorf("%w: schema_version is missing", ErrMalformed)
	}
	switch version := *header.SchemaVersion; {
	case version > Version:
		return nil, &NewerError{Version: version, WrittenAt: writtenAt(header.UpdatedAt)}
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
	// A file is read only with room in every count for one more, so that
	// whatever the next step adds can be written: a count at the most, set
	// there by hand, would otherwise make every write after it fail.
	if err := p.validate(mostCount - 1); err != nil {
		return nil, err
	}
	return &p, nil
}

// inOrder is a picture's description with the keys of every object in order,
// as the file writes the rest of itself, and its numbers as they were written;
// none stays none. The profile never reads the picture, and keeps it as it
// came but for the order of its keys.
func inOrder(picture json.RawMessage) (json.RawMessage, error) {
	if len(bytes.TrimSpace(picture)) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(picture))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("profile: read the picture: %w", err)
	}
	var out bytes.Buffer
	encoder := json.NewEncoder(&out)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, fmt.Errorf("profile: write the picture: %w", err)
	}
	return bytes.TrimRight(out.Bytes(), "\n"), nil
}

// MaxSize is the most a profile may be as a file. A store reads no more than
// this back, so none is ever written larger: a file the store could not read
// again would be taken for damage and rolled back. The caps keep a profile far
// below it, which leaves room for a file edited by hand.
const MaxSize = 1 << 20

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
	if out.Len() > MaxSize {
		return nil, fmt.Errorf("%w: the file would be %d bytes, and a profile is at most %d", ErrInvalid, out.Len(), MaxSize)
	}
	return out.Bytes(), nil
}
