package content

import (
	"maps"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

// gapsTopic is the topic the templates below are written for: one that has
// reference tasks to generalise a template from.
const gapsTopic = "counting.gaps"

// templateFile is where one template of a topic lives.
func templateFile(topic, name string) string {
	return templatesDir + "/" + topic + "/" + name + solverSuffix
}

// withTemplates puts these templates, by name, in place of a topic's real
// ones, so that a test is looking at the templates it wrote and nothing else.
// With none given, the topic is left without templates.
func withTemplates(src fstest.MapFS, topic string, programs map[string]string) {
	for name := range src {
		if strings.HasPrefix(name, templatesDir+"/"+topic+"/") {
			delete(src, name)
		}
	}
	for name, program := range programs {
		src[templateFile(topic, name)] = &fstest.MapFile{Data: []byte(program)}
	}
}

// sampleTemplate is a template with nothing wrong with it, generalised from
// the reference task named.
func sampleTemplate(source string) string {
	return "# For counting a row: a sample.\n# From reference task " + source + ".\n\n" +
		"def solve(options):\n    return match(options, 3)\n"
}

func TestATemplateIsReadWithTheTaskItCameFrom(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	withTemplates(src, gapsTopic, map[string]string{
		"rows":  sampleTemplate("gaps-12-d1-2"),
		"posts": sampleTemplate("gaps-12-d1-1"),
	})
	loaded, err := load(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	want := []Template{
		{Topic: gapsTopic, Name: "posts", Source: "gaps-12-d1-1", Program: sampleTemplate("gaps-12-d1-1")},
		{Topic: gapsTopic, Name: "rows", Source: "gaps-12-d1-2", Program: sampleTemplate("gaps-12-d1-2")},
	}
	if got := loaded.Templates(gapsTopic); !reflect.DeepEqual(got, want) {
		t.Errorf("Templates(%s) = %+v, want %+v in the order of their names", gapsTopic, got, want)
	}
	if got := loaded.Templates("counting.everything"); got != nil {
		t.Errorf("Templates(counting.everything) = %+v, want none for a topic nobody has", got)
	}
}

// A template saved where lines end the Windows way is read as the interpreter
// reads it, the same program naming the same task.
func TestATemplateWithWindowsLineEndingsIsReadTheSameWay(t *testing.T) {
	t.Parallel()

	program := strings.ReplaceAll(sampleTemplate("gaps-12-d1-1"), "\n", "\r\n")
	src := contentCopy(t)
	withTemplates(src, gapsTopic, map[string]string{"posts": program})
	loaded, err := load(src)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if got := loaded.Templates(gapsTopic); len(got) != 1 || got[0].Source != "gaps-12-d1-1" {
		t.Errorf("Templates(%s) = %+v, want the one template, from gaps-12-d1-1", gapsTopic, got)
	}
}

// A template the loader cannot vouch for stops the service: one it cannot
// place under a topic, one that is not a program by its name, and one that
// does not say which reference task it came from — or names one it could not
// have come from, since the bench proves a template on that task.
func TestABrokenTemplateStopsTheService(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		spoil func(fstest.MapFS)
		want  string
	}{
		{"a template beside the topics", func(src fstest.MapFS) {
			src[templatesDir+"/posts.star"] = &fstest.MapFile{Data: []byte(sampleTemplate("gaps-12-d1-1"))}
		}, "posts.star: a template lives in the directory of its topic, not beside it"},
		{"a file named after a topic", func(src fstest.MapFS) {
			src[templatesDir+"/fractions.parts"] = &fstest.MapFile{Data: []byte(sampleTemplate("gaps-12-d1-1"))}
		}, "fractions.parts: a template lives in the directory of its topic, not beside it"},
		{"a directory for a topic nobody has", func(src fstest.MapFS) {
			src[templateFile("counting.everything", "posts")] = &fstest.MapFile{Data: []byte(sampleTemplate("gaps-12-d1-1"))}
		}, "counting.everything: this is not a topic in the catalog"},
		{"a file that is not a program", func(src fstest.MapFS) {
			src[templatesDir+"/"+gapsTopic+"/notes.txt"] = &fstest.MapFile{Data: []byte("a note\n")}
		}, gapsTopic + "/notes.txt: a template is a file named in lowercase words joined by dashes"},
		{"a name that is not lowercase words", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{"Posts": sampleTemplate("gaps-12-d1-1")})
		}, gapsTopic + "/Posts.star: a template is a file named in lowercase words joined by dashes"},
		{"a directory named like a template", func(src fstest.MapFS) {
			src[templatesDir+"/"+gapsTopic+"/old.star/posts.star"] = &fstest.MapFile{Data: []byte(sampleTemplate("gaps-12-d1-1"))}
		}, gapsTopic + "/old.star: a template is a file"},
		{"an empty template", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{"posts": "\n   \n"})
		}, gapsTopic + "/posts.star: the template is empty"},
		{"a template that names no task", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{"posts": "def solve(options):\n    return match(options, 3)\n"})
		}, gapsTopic + "/posts.star: the template does not name the reference task it came from"},
		{"a template whose line names nothing", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{
				"posts": strings.Replace(sampleTemplate("gaps-12-d1-1"), "task gaps-12-d1-1.", "task .", 1),
			})
		}, gapsTopic + "/posts.star: the template does not name the reference task it came from"},
		{"a template that names two tasks", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{
				"posts": "# From reference task gaps-12-d1-2.\n" + sampleTemplate("gaps-12-d1-1"),
			})
		}, gapsTopic + "/posts.star: the template names 2 reference tasks it came from, and it comes from one"},
		{"a template that names a task nobody has", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{"posts": sampleTemplate("gaps-99-d9-9")})
		}, gapsTopic + `/posts.star: no reference task is called "gaps-99-d9-9"`},
		{"a template that names a task of another topic", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{"posts": sampleTemplate("ord-12-d1-1")})
		}, gapsTopic + `/posts.star: reference task "ord-12-d1-1" is a task of logic.ordering, not of counting.gaps`},
		{"three templates for one topic", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{
				"posts": sampleTemplate("gaps-12-d1-1"), "rows": sampleTemplate("gaps-12-d1-2"),
				"trees": sampleTemplate("gaps-12-d1-3"),
			})
		}, gapsTopic + ": 3 templates, and a topic has at most 2"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			test.spoil(src)
			wantProblem(t, src, test.want)
		})
	}
}

// A template or a topic's directory that is there and cannot be read stops the
// service, rather than leaving the topic with fewer templates than it has.
func TestATemplateThatCannotBeReadStopsTheService(t *testing.T) {
	t.Parallel()

	for _, name := range []string{templateFile(gapsTopic, "posts"), templatesDir + "/" + gapsTopic} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			src := contentCopy(t)
			withTemplates(src, gapsTopic, map[string]string{"posts": sampleTemplate("gaps-12-d1-1")})
			wantProblem(t, unreadable{MapFS: src, name: name}, "content: read "+name)
		})
	}
}

// The templates are part of the content: a binary without their directory is
// one whose model would write every solver from nothing.
func TestContentWithoutItsTemplatesStopsTheService(t *testing.T) {
	t.Parallel()

	src := contentCopy(t)
	for name := range src {
		if strings.HasPrefix(name, templatesDir+"/") {
			delete(src, name)
		}
	}
	wantProblem(t, src, "content: read "+templatesDir)
}

// A template is part of what the model is told, so the version every log line
// carries follows it: rewritten, renamed or moved to another topic, it is a
// different version.
func TestAChangedTemplateChangesTheInstructionsVersion(t *testing.T) {
	t.Parallel()

	written := map[string]string{"posts": sampleTemplate("gaps-12-d1-1")}
	base := contentCopy(t)
	withTemplates(base, gapsTopic, written)
	before, err := load(base)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	for _, test := range []struct {
		name   string
		change func(fstest.MapFS)
	}{
		{"a word of it is rewritten", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{"posts": strings.Replace(written["posts"], "a sample", "an example", 1)})
		}},
		{"it is renamed", func(src fstest.MapFS) {
			withTemplates(src, gapsTopic, map[string]string{"fence-posts": written["posts"]})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			src := maps.Clone(base)
			test.change(src)
			after, err := load(src)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			if after.InstructionsVersion() == before.InstructionsVersion() {
				t.Errorf("version = %q both times, want a different one once the template changed",
					after.InstructionsVersion())
			}
		})
	}
}

// Two topics may each have a template of one name, even of one text, and the
// version still tells them apart: a template is counted under its whole path,
// topic included.
func TestTheSameTemplateUnderAnotherTopicIsAnotherVersion(t *testing.T) {
	t.Parallel()

	same := Template{Name: "search", Program: sampleTemplate("gaps-12-d1-1")}
	one := instructionsVersion(versioned(map[string][]Template{gapsTopic: {same}}))
	other := instructionsVersion(versioned(map[string][]Template{"logic.ordering": {same}}))
	if one == other {
		t.Errorf("version = %q under both topics, want two different ones", one)
	}
}
