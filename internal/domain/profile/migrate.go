package profile

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// A migration turns a file of one shape into a file of the next one. It works
// on the document rather than on the types of this package, because the types
// describe the shape it is being migrated to and not the one it is in.
//
// Each step does one version. The chain is walked in memory and the file is
// written back in the current shape, so a profile is migrated by being used.
type migration func(document map[string]json.RawMessage) error

// migrations is the chain, keyed by the version each step reads.
//
// It is empty because there has only ever been one shape. The walk above it
// exists all the same: the first time a field is removed or comes to mean
// something else, what is needed is a step here and a fixture beside it, not a
// decision about how migrating should work taken on the day it is needed.
var migrations = map[int]migration{}

// migrate walks a document up to the current version.
func migrate(raw []byte, from int, chain map[int]migration) ([]byte, error) {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformed, err)
	}

	for version := from; version < Version; version++ {
		step, known := chain[version]
		if !known {
			return nil, fmt.Errorf("%w: version %d, and this build knows no way up from it", ErrOlder, version)
		}
		if err := step(document); err != nil {
			return nil, fmt.Errorf("%w: from version %d: %w", ErrOlder, version, err)
		}
		document["schema_version"] = json.RawMessage(strconv.Itoa(version + 1))
	}

	migrated, err := json.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMalformed, err)
	}
	return migrated, nil
}
