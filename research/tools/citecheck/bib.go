package main

import (
	"fmt"
	"strings"
)

// Entry is one entry of a BibTeX file: a type, a key and its fields in the
// order they were written.
type Entry struct {
	Type   string
	Key    string
	Fields []Field
}

// Field is one `name = {value}` of an entry.
type Field struct {
	Name  string
	Value string
}

// Get is the value of a field, or "" when the entry has none. Field names are
// matched without regard to case, as BibTeX matches them.
func (e Entry) Get(name string) string {
	for _, f := range e.Fields {
		if strings.EqualFold(f.Name, name) {
			return f.Value
		}
	}
	return ""
}

// ParseBib reads every entry of a BibTeX file: `@type{key,` followed by
// `name = {value}` fields separated by commas and a closing brace. A value is
// braced, possibly with braces inside, or a bare word such as a year. Lines
// that start with `%` between entries are comments. A key used twice, or a
// field given twice in one entry, is an error: either would make one citation
// mean two things.
func ParseBib(text string) ([]Entry, error) {
	p := &bibParser{text: text}
	var entries []Entry
	firstLine := map[string]int{}
	for {
		p.skipSpaceAndComments()
		if p.done() {
			return entries, nil
		}
		line := p.line()
		entry, err := p.entry()
		if err != nil {
			return nil, err
		}
		if entry.Type == "comment" {
			continue
		}
		key := strings.ToLower(entry.Key)
		if first, seen := firstLine[key]; seen {
			return nil, fmt.Errorf("line %d: key %s is already used at line %d", line, entry.Key, first)
		}
		firstLine[key] = line
		entries = append(entries, entry)
	}
}

// FormatEntry writes an entry the way refs.bib keeps them: one field a line,
// the equals signs aligned, every value braced.
func FormatEntry(e Entry) string {
	width := 0
	for _, f := range e.Fields {
		width = max(width, len(f.Name))
	}
	var b strings.Builder
	fmt.Fprintf(&b, "@%s{%s,\n", e.Type, e.Key)
	for _, f := range e.Fields {
		fmt.Fprintf(&b, "  %-*s = {%s},\n", width, f.Name, f.Value)
	}
	b.WriteString("}\n")
	return b.String()
}

type bibParser struct {
	text string
	pos  int
}

func (p *bibParser) done() bool { return p.pos >= len(p.text) }

func (p *bibParser) peek() byte {
	if p.done() {
		return 0
	}
	return p.text[p.pos]
}

func (p *bibParser) line() int { return 1 + strings.Count(p.text[:p.pos], "\n") }

func (p *bibParser) errorf(format string, args ...any) error {
	return fmt.Errorf("line %d: %s", p.line(), fmt.Sprintf(format, args...))
}

func (p *bibParser) skipSpace() {
	for !p.done() && strings.ContainsRune(" \t\r\n", rune(p.text[p.pos])) {
		p.pos++
	}
}

func (p *bibParser) skipSpaceAndComments() {
	for {
		p.skipSpace()
		if p.peek() != '%' {
			return
		}
		for !p.done() && p.text[p.pos] != '\n' {
			p.pos++
		}
	}
}

func (p *bibParser) entry() (Entry, error) {
	if p.peek() != '@' {
		return Entry{}, p.errorf("an entry starts with @")
	}
	p.pos++
	kind, err := p.word("an entry's type")
	if err != nil {
		return Entry{}, err
	}
	p.skipSpace()
	if p.peek() != '{' {
		return Entry{}, p.errorf("@%s is not followed by {", kind)
	}
	switch strings.ToLower(kind) {
	case "comment":
		_, err = p.value("a comment")
		return Entry{Type: "comment"}, err
	case "string", "preamble":
		return Entry{}, p.errorf("@%s is not used here: write the value out in the entry", kind)
	}
	p.pos++
	p.skipSpace()
	key, err := p.key()
	if err != nil {
		return Entry{}, err
	}
	fields, err := p.fields(key)
	if err != nil {
		return Entry{}, err
	}
	return Entry{Type: strings.ToLower(kind), Key: key, Fields: fields}, nil
}

// fields reads an entry's fields up to the brace that closes it.
func (p *bibParser) fields(key string) ([]Field, error) {
	var fields []Field
	seen := map[string]bool{}
	for {
		p.skipSpace()
		if p.peek() == '}' {
			p.pos++
			return fields, nil
		}
		field, err := p.field()
		if err != nil {
			return nil, err
		}
		if seen[field.Name] {
			return nil, p.errorf("%s has the field %s twice", key, field.Name)
		}
		seen[field.Name] = true
		fields = append(fields, field)
		p.skipSpace()
		switch p.peek() {
		case ',':
			p.pos++
		case '}':
		default:
			if p.done() {
				return nil, p.errorf("the entry %s never closes its brace", key)
			}
			return nil, p.errorf("%s: a field is followed by , or }", key)
		}
	}
}

// key reads an entry's key, up to the comma that ends it.
func (p *bibParser) key() (string, error) {
	start := p.pos
	for !p.done() && !strings.ContainsRune(", \t\r\n{}", rune(p.text[p.pos])) {
		p.pos++
	}
	key := p.text[start:p.pos]
	p.skipSpace()
	if key == "" || p.peek() != ',' {
		return "", p.errorf("an entry's key is one word followed by a comma")
	}
	p.pos++
	return key, nil
}

func (p *bibParser) field() (Field, error) {
	name, err := p.word("a field's name")
	if err != nil {
		return Field{}, err
	}
	p.skipSpace()
	if p.peek() != '=' {
		return Field{}, p.errorf("the field %s is not followed by =", name)
	}
	p.pos++
	p.skipSpace()
	value, err := p.value(name)
	if err != nil {
		return Field{}, err
	}
	return Field{Name: strings.ToLower(name), Value: value}, nil
}

// word reads letters, digits and the few other characters BibTeX allows in
// a type or a field name.
func (p *bibParser) word(what string) (string, error) {
	start := p.pos
	for !p.done() && isWordByte(p.text[p.pos]) {
		p.pos++
	}
	if p.pos == start {
		return "", p.errorf("%s is missing", what)
	}
	return p.text[start:p.pos], nil
}

func isWordByte(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-'
}

// value reads a braced value, keeping the braces inside it, a quoted one, or a
// bare word such as a year. A quoted value ends at a quote outside braces.
func (p *bibParser) value(name string) (string, error) {
	switch p.peek() {
	case '"':
		return p.quoted(name)
	case '{':
	default:
		return p.word("the value of " + name)
	}
	start, depth := p.pos+1, 0
	for !p.done() {
		switch p.text[p.pos] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				p.pos++
				return p.text[start : p.pos-1], nil
			}
		}
		p.pos++
	}
	return "", p.errorf("the value of %s never closes its brace", name)
}

func (p *bibParser) quoted(name string) (string, error) {
	start, depth := p.pos+1, 0
	for p.pos++; !p.done(); p.pos++ {
		switch p.text[p.pos] {
		case '{':
			depth++
		case '}':
			depth--
			if depth < 0 {
				return "", p.errorf("the value of %s closes a brace it never opened", name)
			}
		case '"':
			if depth == 0 {
				p.pos++
				return p.text[start : p.pos-1], nil
			}
		}
	}
	return "", p.errorf("the value of %s never closes its quote", name)
}
