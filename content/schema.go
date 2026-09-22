package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"slices"
)

const schemasDir = "schemas"

// schemaFiles are the three formats the exchange with the model rests on, in
// the order that exchange happens: what the model is asked for, what it hands
// back, and the pass it makes over its own work. The list is here so that a
// schema added to the directory and forgotten everywhere else stops the service
// instead of going unnoticed.
var schemaFiles = []string{"brief.json", "task.json", "self_check.json"}

// loadSchemas reads the JSON schemas and refuses a set that is not the expected
// three, or one that is not a schema a reader could follow.
func loadSchemas(src fs.FS) (map[string][]byte, error) {
	entries, err := fs.ReadDir(src, schemasDir)
	if err != nil {
		return nil, fmt.Errorf("content: read %s: %w", schemasDir, err)
	}

	p := &problems{file: schemasDir}
	checkSchemaSet(p, entries)
	// Reading a set that is already wrong would only add the failure of an
	// absent file to a list that already says what is absent.
	if err := p.err(); err != nil {
		return nil, err
	}

	schemas := make(map[string][]byte, len(schemaFiles))
	for _, name := range schemaFiles {
		file := schemasDir + "/" + name
		raw, err := fs.ReadFile(src, file)
		if err != nil {
			return nil, fmt.Errorf("content: read %s: %w", file, err)
		}
		if err := checkSchemaDocument(p, file, raw); err != nil {
			return nil, err
		}
		schemas[name] = raw
	}
	return schemas, p.err()
}

// checkSchemaSet checks that the directory holds the expected schemas and
// nothing besides, naming whichever file is the surprise and whichever is gone.
func checkSchemaSet(p *problems, entries []fs.DirEntry) {
	found := make(map[string]bool, len(entries))
	for _, entry := range entries {
		found[entry.Name()] = true
		if !slices.Contains(schemaFiles, entry.Name()) {
			p.addf("%s: the schemas are %v, and this is not one of them", entry.Name(), schemaFiles)
		}
	}
	for _, name := range schemaFiles {
		if !found[name] {
			p.addf("%s: the schema is missing", name)
		}
	}
}

// checkSchemaDocument checks that a file is a schema a reader could follow. The
// error is for a file that is not a JSON document at all; anything else it finds
// is a fault on the list.
func checkSchemaDocument(p *problems, file string, raw []byte) error {
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		return fmt.Errorf("content: parse %s: %w", file, err)
	}
	// All three describe an object with fields. A schema built from a root
	// combinator instead would carry no properties of its own, and this list
	// would have to grow to let it through.
	for _, key := range []string{"$schema", "title", "type", "properties"} {
		if _, ok := document[key]; !ok {
			p.addf("%s: a schema says %q, and this one does not", path.Base(file), key)
		}
	}
	return nil
}
