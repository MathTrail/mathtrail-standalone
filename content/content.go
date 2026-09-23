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
// Load, and what it hands out shares no memory with what it holds: a caller may
// sort, append to or rewrite its copy, down to an option inside a reference
// task, and the next caller is handed the content as it was built.
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
func (c *Content) Topics() []Topic {
	topics := make([]Topic, 0, len(c.topics))
	for _, topic := range c.topics {
		topics = append(topics, topic.clone())
	}
	return topics
}

// Topic returns the topic with this id, and reports whether the catalog has one.
func (c *Content) Topic(id string) (Topic, bool) {
	topic, ok := c.topicByID[id]
	if !ok {
		return Topic{}, false
	}
	return topic.clone(), true
}

// Traps returns the whole trap catalog, in catalog order.
func (c *Content) Traps() []Trap { return slices.Clone(c.traps) }

// Trap returns the trap with this id, and reports whether the catalog has one.
func (c *Content) Trap(id string) (Trap, bool) {
	trap, ok := c.trapByID[id]
	return trap, ok
}

// HasTopic reports whether the topic catalog has this id.
func (c *Content) HasTopic(id string) bool {
	_, ok := c.topicByID[id]
	return ok
}

// HasTrap reports whether the trap catalog has this id.
func (c *Content) HasTrap(id string) bool {
	_, ok := c.trapByID[id]
	return ok
}

// HasSkill reports whether the skill catalog has this id.
func (c *Content) HasSkill(id string) bool {
	_, ok := c.skillByID[id]
	return ok
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
func (c *Content) Examples() []Example {
	examples := make([]Example, 0, len(c.examples))
	for i := range c.examples {
		examples = append(examples, c.examples[i].clone())
	}
	return examples
}

// ExampleCount is how many reference tasks the binary carries, for a caller
// that wants the number rather than the tasks.
func (c *Content) ExampleCount() int { return len(c.examples) }

// TopicsAt lists the topics a child of this grade can be given, in catalog
// order. A grade the levels do not cover has no topics: the caller is told
// nothing rather than something wrong.
func (c *Content) TopicsAt(grade int) []string {
	level, known := LevelOf(grade)
	if !known {
		return nil
	}

	ids := make([]string, 0, len(c.topics))
	for _, topic := range c.topics {
		if topic.HasLevel(level) {
			ids = append(ids, topic.ID)
		}
	}
	return ids
}

// TrapIDs lists the whole trap catalog in catalog order. It is the order two
// traps are told apart by when nothing else separates them, which is what
// keeps a choice made over them from depending on the order a map happened to
// be walked in.
func (c *Content) TrapIDs() []string {
	ids := make([]string, 0, len(c.traps))
	for _, trap := range c.traps {
		ids = append(ids, trap.ID)
	}
	return ids
}

// ExampleTraps lists the traps of the reference tasks of one topic at one
// grade, the most frequent first and ties broken by catalog order.
//
// The reference tasks already carry a trap on every wrong option, so what a
// topic's usual mistakes are is a count over them rather than a list anybody
// has to keep up to date beside the catalog.
func (c *Content) ExampleTraps(topic string, grade int) []string {
	level, known := LevelOf(grade)
	if !known {
		return nil
	}

	seen := map[string]int{}
	for i := range c.examples {
		task := &c.examples[i]
		if task.Topic != topic || task.GradeLevel != level {
			continue
		}
		for _, distractor := range task.Distractors {
			seen[distractor.Trap]++
		}
	}

	ids := make([]string, 0, len(seen))
	for _, trap := range c.traps { // catalog order, which settles every tie
		if seen[trap.ID] > 0 {
			ids = append(ids, trap.ID)
		}
	}
	slices.SortStableFunc(ids, func(a, b string) int { return seen[b] - seen[a] })
	return ids
}

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
