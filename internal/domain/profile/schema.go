package profile

import _ "embed"

// schema describes the file for whoever is not this package: a person reading
// it in Drive, or a tool that has to make sense of an exported profile. It is
// not what this service validates against — the types above and Validate are —
// so a test holds the two to naming the same fields, which is the only way a
// description kept beside the code stays true to it.
//
//go:embed profile.schema.json
var schema []byte

// Schema is the JSON Schema of a profile file.
func Schema() []byte { return schema }
