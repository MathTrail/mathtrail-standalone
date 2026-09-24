package content

import (
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strings"
)

// Template is a sample solver for one topic: the solver of one of its
// reference tasks, generalised, so that a model writing a task of that topic
// starts from a search that is known to work rather than from nothing.
//
// It keeps the numbers of the task it came from. That is what lets it still
// run and still prove that task's answer, which is how a template is known to
// be a correct program rather than a plausible one; the places a model puts
// its own numbers are the capitals at its top.
type Template struct {
	// Topic is the topic it is a sample for.
	Topic string
	// Name tells it from the other template of its topic.
	Name string
	// Source is the reference task whose solver it was generalised from.
	Source string
	// Program is the Starlark source, exactly as the model is shown it.
	Program string
}

const (
	// templatesDir holds one directory per topic, each with the templates of
	// that topic. It is not the directory of the reference tasks' own solvers,
	// which lives among the reference tasks.
	templatesDir = "solvers"
	// templatesPerTopic is the most templates a topic has: the search its
	// tasks usually need and, where the topic has another of a different kind,
	// that one as well.
	templatesPerTopic = 2
)

// Every template names the reference task it came from on a line of its own,
// as in "# From reference task ord-34-d2-2.". The line may end the way Windows
// ends lines, because the interpreter reads such a program as it reads any
// other, and a line whose id is missing is a line that names no task.
var sourceLine = regexp.MustCompile(`(?m)^# From reference task (\S+)\.\r?$`)

// Templates returns the solver templates of one topic, in the order a package
// shows them. A topic with none — one whose reference tasks are still to be
// written, so that nothing exists to generalise from — returns none.
func (c *Content) Templates(topic string) []Template { return slices.Clone(c.templates[topic]) }

// loadTemplates reads the solver templates, one directory per topic, and
// refuses what could not be what it claims to be: a template in a directory
// no topic has, one that says nothing, one that does not name the reference
// task it came from or names a task of another topic, and a topic with more
// templates than it may have.
func loadTemplates(src fs.FS, topics map[string]Topic, examples []Example) (map[string][]Template, error) {
	entries, err := fs.ReadDir(src, templatesDir)
	if err != nil {
		return nil, fmt.Errorf("content: read %s: %w", templatesDir, err)
	}

	topicOf := make(map[string]string, len(examples))
	for i := range examples {
		topicOf[examples[i].ID] = examples[i].Topic
	}

	p := &problems{file: templatesDir}
	templates := make(map[string][]Template, len(entries))
	for _, entry := range entries {
		if _, known := topics[entry.Name()]; !entry.IsDir() || !known {
			p.addf("%s: %s", entry.Name(), misplaced(entry))
			continue
		}
		ofTopic, err := loadTopicTemplates(src, p, entry.Name(), topicOf)
		if err != nil {
			return nil, err
		}
		templates[entry.Name()] = ofTopic
	}
	return templates, p.err()
}

// misplaced says what is wrong with something beside the topics' directories:
// a template lying loose, or a directory for a topic the catalog does not
// have, whose templates no package would ever show.
func misplaced(entry fs.DirEntry) string {
	if !entry.IsDir() {
		return "a template lives in the directory of its topic, not beside it"
	}
	return "this is not a topic in the catalog, so no template here can be shown"
}

// loadTopicTemplates reads the templates of one topic, in the order of their
// names, which is the order a package shows them in.
func loadTopicTemplates(src fs.FS, p *problems, topic string, topicOf map[string]string) ([]Template, error) {
	dir := templatesDir + "/" + topic
	entries, err := fs.ReadDir(src, dir)
	if err != nil {
		return nil, fmt.Errorf("content: read %s: %w", dir, err)
	}

	var templates []Template
	for _, entry := range entries {
		where := topic + "/" + entry.Name()
		name, named := strings.CutSuffix(entry.Name(), solverSuffix)
		if entry.IsDir() || !named || !exampleIDPattern.MatchString(name) {
			p.addf("%s: a template is a file named in lowercase words joined by dashes, ending in %s",
				where, solverSuffix)
			continue
		}
		program, err := fs.ReadFile(src, dir+"/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("content: read %s/%s: %w", dir, entry.Name(), err)
		}
		// Kept even when its source is at fault, so that a topic with too many
		// templates is told so in the same pass as whatever else is wrong with
		// them; a fault anywhere means the content is not used at all.
		templates = append(templates, Template{
			Topic: topic, Name: name, Program: string(program),
			Source: sourceOf(p, where, topic, string(program), topicOf),
		})
	}
	if len(templates) > templatesPerTopic {
		p.addf("%s: %d templates, and a topic has at most %d", topic, len(templates), templatesPerTopic)
	}
	return templates, nil
}

// sourceOf is the reference task a template names as the one it came from,
// once it is shown to be a task of the template's own topic: a pouring
// template generalised from a calendar task would teach the wrong search, and
// one naming a task nobody has could not be proved on it.
func sourceOf(p *problems, where, topic, program string, topicOf map[string]string) string {
	if strings.TrimSpace(program) == "" {
		p.addf("%s: the template is empty", where)
		return ""
	}

	named := sourceLine.FindAllStringSubmatch(program, -1)
	switch {
	case len(named) == 0:
		p.addf("%s: the template does not name the reference task it came from, on a line of its own: "+
			"# From reference task <id>.", where)
		return ""
	case len(named) > 1:
		p.addf("%s: the template names %d reference tasks it came from, and it comes from one", where, len(named))
		return ""
	}

	source := named[0][1]
	switch sourceTopic, known := topicOf[source]; {
	case !known:
		p.addf("%s: no reference task is called %q", where, source)
	case sourceTopic != topic:
		p.addf("%s: reference task %q is a task of %s, not of %s", where, source, sourceTopic, topic)
	}
	return source
}

// versioned are the templates as the version of the instructions counts them:
// each under its own path, so that a template moved to another topic or
// renamed changes the version as surely as one rewritten. They come in no
// particular order, since the version sorts what it hashes by name.
func versioned(templates map[string][]Template) []instructionFile {
	var files []instructionFile
	for topic, ofTopic := range templates {
		for _, template := range ofTopic {
			files = append(files, instructionFile{
				name: templatesDir + "/" + topic + "/" + template.Name + solverSuffix,
				text: template.Program,
			})
		}
	}
	return files
}

// templatePrograms are the programs of a topic's templates, as a package
// carries them: a list that is empty rather than null for a topic with none.
func (c *Content) templatePrograms(topic string) []string {
	programs := make([]string, 0, templatesPerTopic)
	for _, template := range c.templates[topic] {
		programs = append(programs, template.Program)
	}
	return programs
}
