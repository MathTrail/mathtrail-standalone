package content

import (
	"encoding/json"
	"fmt"
	"io/fs"
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
	found := make([]string, 0, len(entries))
	for _, entry := range entries {
		found = append(found, entry.Name())
	}
	if expected := slices.Sorted(slices.Values(schemaFiles)); !slices.Equal(slices.Sorted(slices.Values(found)), expected) {
		p.addf("the schemas are %v, and this directory holds %v", schemaFiles, found)
		return nil, p.err()
	}

	schemas := make(map[string][]byte, len(schemaFiles))
	for _, name := range schemaFiles {
		file := schemasDir + "/" + name
		var document map[string]json.RawMessage
		if err := decode(src, file, &document); err != nil {
			return nil, err
		}
		for _, key := range []string{"$schema", "title", "type", "properties"} {
			if _, ok := document[key]; !ok {
				p.addf("%s: a schema says %q, and this one does not", name, key)
			}
		}
		raw, err := fs.ReadFile(src, file)
		if err != nil {
			return nil, fmt.Errorf("content: read %s: %w", file, err)
		}
		schemas[name] = raw
	}
	return schemas, p.err()
}
