package drivetest

import (
	"errors"
	"fmt"
	"strings"
)

// matcher is a search of Drive's, as a question about one file.
type matcher func(*File) bool

// token is one word of a search: a name, a sign, or a quoted value.
type token struct {
	text   string
	quoted bool
}

// parseQuery reads a search in Drive's language, as far as the service speaks
// it: terms joined by and, each one either
//
//	appProperties has { key='…' and value='…' }
//	trashed = true | false
//
// Anything else is refused, so that a search the stand-in cannot answer the
// way Drive would fails the test that made it instead of finding the wrong
// files.
func parseQuery(q string) (matcher, error) {
	tokens, err := tokenize(q)
	if err != nil {
		return nil, err
	}
	read := &parser{tokens: tokens}
	var terms []matcher
	for {
		term, err := read.term()
		if err != nil {
			return nil, err
		}
		terms = append(terms, term)
		if read.done() {
			break
		}
		if err := read.expect("and"); err != nil {
			return nil, err
		}
	}
	return func(file *File) bool {
		for _, term := range terms {
			if !term(file) {
				return false
			}
		}
		return true
	}, nil
}

// tokenize splits a search into its words. A quoted value keeps what a
// backslash escapes — a quote, or a backslash — as itself.
func tokenize(q string) ([]token, error) {
	var tokens []token
	for i := 0; i < len(q); {
		c := q[i]
		switch {
		case c == ' ':
			i++
		case c == '{' || c == '}' || c == '=':
			tokens = append(tokens, token{text: string(c)})
			i++
		case c == '\'':
			value, next, err := quotedValue(q, i+1)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, token{text: value, quoted: true})
			i = next
		case isLetter(c):
			start := i
			for i < len(q) && isLetter(q[i]) {
				i++
			}
			tokens = append(tokens, token{text: q[start:i]})
		default:
			return nil, fmt.Errorf("the stand-in reads no %q in a search", c)
		}
	}
	return tokens, nil
}

// quotedValue reads a quoted value from just past its opening quote, and
// answers the value and where the search goes on after its closing quote.
func quotedValue(q string, from int) (value string, next int, err error) {
	var unescaped strings.Builder
	for i := from; i < len(q); i++ {
		switch q[i] {
		case '\'':
			return unescaped.String(), i + 1, nil
		case '\\':
			i++
			if i == len(q) {
				return "", 0, errors.New("a search ends inside an escape")
			}
		}
		unescaped.WriteByte(q[i])
	}
	return "", 0, errors.New("a quoted value is never closed")
}

// isLetter reports whether a byte is an ASCII letter.
func isLetter(c byte) bool { return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' }

// parser reads the words of a search one after another.
type parser struct {
	tokens []token
	next   int
}

// done reports whether every word has been read.
func (p *parser) done() bool { return p.next == len(p.tokens) }

// term reads one term of a search.
func (p *parser) term() (matcher, error) {
	switch {
	case p.accept("appProperties"):
		return p.property()
	case p.accept("trashed"):
		if err := p.expect("="); err != nil {
			return nil, err
		}
		switch {
		case p.accept("true"):
			return func(file *File) bool { return file.Trashed }, nil
		case p.accept("false"):
			return func(file *File) bool { return !file.Trashed }, nil
		}
		return nil, errors.New("trashed is true or false")
	}
	if p.done() {
		return nil, errors.New("a search ends where a term was due")
	}
	return nil, fmt.Errorf("the stand-in does not search by %q", p.tokens[p.next].text)
}

// property reads the rest of a search by one of the service's own properties.
func (p *parser) property() (matcher, error) {
	key, err := p.valueAfter("has", "{", "key", "=")
	if err != nil {
		return nil, err
	}
	value, err := p.valueAfter("and", "value", "=")
	if err != nil {
		return nil, err
	}
	if err := p.expect("}"); err != nil {
		return nil, err
	}
	return func(file *File) bool {
		held, has := file.AppProperties[key]
		return has && held == value
	}, nil
}

// valueAfter reads the words given, and then a quoted value.
func (p *parser) valueAfter(words ...string) (string, error) {
	if err := p.expect(words...); err != nil {
		return "", err
	}
	return p.value()
}

// accept reads the next word when it is the one given.
func (p *parser) accept(word string) bool {
	if p.done() || p.tokens[p.next].quoted || p.tokens[p.next].text != word {
		return false
	}
	p.next++
	return true
}

// expect reads the words given, in order, and refuses a search that says
// anything else there.
func (p *parser) expect(words ...string) error {
	for _, word := range words {
		if !p.accept(word) {
			return fmt.Errorf("a search goes on with %q here", word)
		}
	}
	return nil
}

// value reads a quoted value.
func (p *parser) value() (string, error) {
	if p.done() || !p.tokens[p.next].quoted {
		return "", errors.New("a quoted value is due here")
	}
	p.next++
	return p.tokens[p.next-1].text, nil
}
