package content

import (
	"testing"
	"testing/fstest"
)

func TestAMissingSchemaStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	delete(src, schemasDir+"/task.json")
	wantProblem(t, src, "task.json: the schema is missing")
}

// A schema nobody knows about is a format somebody meant to introduce and
// wired up nowhere, which is worth saying out loud rather than ignoring.
func TestAnUnknownSchemaStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[schemasDir+"/profile.json"] = &fstest.MapFile{Data: []byte(`{"type":"object"}`)}
	wantProblem(t, src, "profile.json: the schemas are")
}

func TestASchemaThatDescribesNothingStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[schemasDir+"/task.json"] = &fstest.MapFile{Data: []byte(
		`{"$schema":"https://json-schema.org/draft/2020-12/schema","title":"Task","type":"object"}`)}
	wantProblem(t, src, `task.json: a schema says "properties"`)
}

func TestASchemaThatIsNotADocumentStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	src[schemasDir+"/task.json"] = &fstest.MapFile{Data: []byte(`["a list of rules"]`)}
	wantProblem(t, src, "content: parse "+schemasDir+"/task.json")
}
