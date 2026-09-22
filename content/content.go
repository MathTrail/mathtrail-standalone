// Package content holds everything the service ships that is not code: the
// closed catalogs of topics, traps and skills, the reference tasks the model is
// shown, the JSON schemas of what it has to hand back, and the instructions it
// works from.
//
// All of it is compiled into the binary, so a running service cannot drift from
// the content it was built with, and Load checks the whole set once at startup.
// A reference task naming a topic nobody has, or one with four options instead
// of five, is then a process that refuses to start rather than a lesson that
// breaks halfway through.
package content

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

//go:embed catalogs examples instructions schemas
var files embed.FS

// Content is the checked content of the binary. Nothing in it changes after
// Load; the slices it hands out are copies, so a caller that sorts or appends
// to one leaves the next caller's copy alone, and the entries inside them are
// to be read and not edited.
type Content struct {
	topics   []Topic
	traps    []Trap
	skills   []Skill
	examples []Example

	topicByID map[string]Topic
	trapByID  map[string]Trap
	skillByID map[string]Skill

	schemas             map[string][]byte
	instructions        map[string]string
	instructionsVersion string
}

// Load parses and checks everything embedded in the binary. Faults within one
// file are reported together, so that a broken catalog takes one run to fix
// rather than one run per fault.
func Load() (*Content, error) { return load(files) }

// load reads one set of content. Everything reaches the checks through this
// parameter, so the checks can be shown content the binary does not carry,
// which is the only way to prove that a fault would in fact stop the service.
func load(src fs.FS) (*Content, error) {
	topics, err := loadTopics(src)
	if err != nil {
		return nil, err
	}
	traps, err := loadTraps(src)
	if err != nil {
		return nil, err
	}
	skills, err := loadSkills(src)
	if err != nil {
		return nil, err
	}

	c := &Content{
		topics:    topics,
		traps:     traps,
		skills:    skills,
		topicByID: index(topics, func(t Topic) string { return t.ID }),
		trapByID:  index(traps, func(t Trap) string { return t.ID }),
		skillByID: index(skills, func(s Skill) string { return s.ID }),
	}

	if c.examples, err = loadExamples(src, c.topicByID, c.trapByID); err != nil {
		return nil, err
	}
	if c.schemas, err = loadSchemas(src); err != nil {
		return nil, err
	}

	instructions, err := loadInstructions(src)
	if err != nil {
		return nil, err
	}
	c.instructions = make(map[string]string, len(instructions))
	for _, file := range instructions {
		c.instructions[file.name] = file.text
	}
	c.instructionsVersion = instructionsVersion(instructions)

	return c, nil
}

// Topics returns the whole topic catalog, in catalog order.
func (c *Content) Topics() []Topic { return slices.Clone(c.topics) }

// Topic returns the topic with this id, and reports whether the catalog has one.
func (c *Content) Topic(id string) (Topic, bool) {
	topic, ok := c.topicByID[id]
	return topic, ok
}

// Traps returns the whole trap catalog, in catalog order.
func (c *Content) Traps() []Trap { return slices.Clone(c.traps) }

// Trap returns the trap with this id, and reports whether the catalog has one.
func (c *Content) Trap(id string) (Trap, bool) {
	trap, ok := c.trapByID[id]
	return trap, ok
}

// Skills returns the whole skill catalog, in catalog order.
func (c *Content) Skills() []Skill { return slices.Clone(c.skills) }

// Skill returns the skill with this id, and reports whether the catalog has one.
func (c *Content) Skill(id string) (Skill, bool) {
	skill, ok := c.skillByID[id]
	return skill, ok
}

// Examples returns every reference task, ordered by file and then by position
// within it.
func (c *Content) Examples() []Example { return slices.Clone(c.examples) }

// Schema returns the JSON schema of one of the formats the model works to,
// named as its file is: brief.json, task.json or self_check.json.
func (c *Content) Schema(name string) ([]byte, bool) {
	schema, ok := c.schemas[name]
	return bytes.Clone(schema), ok
}

// Instruction returns one file of instructions for the model, named as the file
// is.
func (c *Content) Instruction(name string) (string, bool) {
	text, ok := c.instructions[name]
	return text, ok
}

// InstructionsVersion identifies the instructions this binary carries. Every log
// line about a generated task holds it, so that months later it is still clear
// which wording produced which task.
func (c *Content) InstructionsVersion() string { return c.instructionsVersion }

// index keys entries by their id, for the lookups the checks and the package
// builder do far more often than they iterate.
func index[T any](entries []T, id func(T) string) map[string]T {
	byID := make(map[string]T, len(entries))
	for _, entry := range entries {
		byID[id(entry)] = entry
	}
	return byID
}

// decode reads one embedded file and refuses any field the type does not
// declare: a misspelt key in the content must stop the service, not be quietly
// dropped on the way in.
func decode(src fs.FS, name string, target any) error {
	raw, err := fs.ReadFile(src, name)
	if err != nil {
		return fmt.Errorf("content: read %s: %w", name, err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("content: parse %s: %w", name, err)
	}
	if dec.More() {
		return fmt.Errorf("content: parse %s: more than one JSON value in the file", name)
	}
	return nil
}

// problems collects everything wrong with one file, so that whoever fixes the
// content is shown every fault at once.
type problems struct {
	file string
	list []string
}

// addf records one fault, in the words the person fixing it needs.
func (p *problems) addf(format string, args ...any) {
	p.list = append(p.list, fmt.Sprintf(format, args...))
}

// err reports everything collected as a single error, or nil when the file is
// sound.
func (p *problems) err() error {
	if len(p.list) == 0 {
		return nil
	}
	return fmt.Errorf("content: %s:\n  - %s", p.file, strings.Join(p.list, "\n  - "))
}
