package mcpserver

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/content"
)

// childInTheChat are the wordings that have the model talk with the child, or
// the child talk in the chat. The adult runs the lesson and types every
// message, and the child answers on the card: whatever the child would say,
// the adult says, and the model talks with the adult. "Show the child's
// progress" names whose progress it is, and is not among them.
var childInTheChat = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(tell|ask|show|greet|praise)\s+the\s+child([^'’\w]|$)`),
	regexp.MustCompile(`(?i)\b(talk|speak|say)\s+to\s+the\s+child\b`),
	regexp.MustCompile(`(?i)\bwith\s+the\s+child\b`),
	regexp.MustCompile(`(?i)\bthe\s+child\s+(asks|says|wants|types|writes)\b`),
	regexp.MustCompile(`(?i)\bthe\s+child\s+(answers|gives)\s+in\s+the\s+chat\b`),
	regexp.MustCompile(`(?i)\bchild['’]s\s+(next\s+)?message\b`),
}

// No words the model reads have it talk with the child, or have the child talk
// in the chat: neither the server instructions nor any text of this package's
// own or of the domain's, which between them hold every tool's description and
// arguments and the words of every result, the rationale a package carries
// and the reasons a refusal gives among them.
func TestNoWordsForTheModelPutTheChildInTheChat(t *testing.T) {
	t.Parallel()

	for _, said := range append(instructionLines(t), packageWordings(t)...) {
		for _, pattern := range childInTheChat {
			if found := pattern.FindString(said.words); found != "" {
				t.Errorf("%s puts the child in the chat with %q: %q", said.where, found, said.words)
			}
		}
	}
}

// wording is words the model may read, and where they are written.
type wording struct {
	where string
	words string
}

// instructionLines are the server instructions, line by line.
func instructionLines(t *testing.T) []wording {
	t.Helper()

	loaded, err := content.Load()
	if err != nil {
		t.Fatalf("content.Load() error = %v", err)
	}
	var lines []wording
	for i, line := range strings.Split(loaded.ServerInstructions(), "\n") {
		lines = append(lines, wording{where: fmt.Sprintf("the server instructions, line %d,", i+1), words: line})
	}
	return lines
}

// packageWordings are the strings of the package's own source and of the
// domain's, their tests aside: every string written in them, struct tags among
// them, with the strings joined by + read as one.
func packageWordings(t *testing.T) []wording {
	t.Helper()

	var files []string
	for _, pattern := range []string{"*.go", "../../domain/*/*.go"} {
		found, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("list the files of %s: %v", pattern, err)
		}
		files = append(files, found...)
	}
	var wordings []wording
	fset := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			found, isString := literalOf(t, node)
			if isString {
				wordings = append(wordings, wording{where: fset.Position(node.Pos()).String(), words: found})
			}
			return !isString
		})
	}
	return wordings
}

// literalOf is the string a node writes: one literal, or literals joined by +.
// Anything else is not one, and is looked into for the strings it holds.
func literalOf(t *testing.T, node ast.Node) (string, bool) {
	t.Helper()

	switch node := node.(type) {
	case *ast.BasicLit:
		if node.Kind != token.STRING {
			return "", false
		}
		unquoted, err := strconv.Unquote(node.Value)
		if err != nil {
			t.Fatalf("read %s: %v", node.Value, err)
		}
		return unquoted, true
	case *ast.BinaryExpr:
		if node.Op != token.ADD {
			return "", false
		}
		left, leftIsString := literalOf(t, node.X)
		right, rightIsString := literalOf(t, node.Y)
		return left + right, leftIsString && rightIsString
	}
	return "", false
}
