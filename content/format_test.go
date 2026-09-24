package content_test

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"

	"go.starlark.net/syntax"

	"github.com/MathTrail/mathtrail-standalone/internal/infra/starlark"
)

// conversions are the letters Starlark fills after a %. It has no flags, no
// width and no precision, so anything else stops the program where it
// formats.
const conversions = "sdioxXeEfFgGcr"

// scalarBuiltins are the built-ins whose result a program formats and which
// always give one value, never a tuple.
var scalarBuiltins = []string{"len", "str", "int", "sum", "abs"}

// A message a program builds with % is only as good as its format: a stray
// %, as in "a rise of 10% on %d", or a format that asks for more values than
// it is given, stops the program with an error about the format in place of
// the message, and exactly when the message was needed. A template is copied
// by the model, mistakes and all. Every string the shipped programs format
// with % has to be one Starlark can fill, whether it is written where it is
// used or kept in a constant at the top of the program.
func TestEveryFormatInTheProgramsIsOneStarlarkCanFill(t *testing.T) {
	t.Parallel()

	embedded := loaded(t)
	programs := map[string]string{}
	examples := embedded.Examples()
	for i := range examples {
		programs["examples/solvers/"+examples[i].ID+".star"] = examples[i].Solver
	}
	for _, topic := range embedded.Topics() {
		for _, template := range embedded.Templates(topic.ID) {
			programs["solvers/"+topic.ID+"/"+template.Name+".star"] = template.Program
		}
	}

	formats := 0
	for _, name := range slices.Sorted(maps.Keys(programs)) {
		for _, use := range formatsIn(t, name, programs[name]) {
			formats++
			if problem := formatProblem(use); problem != "" {
				t.Errorf("%s:%d: format %q %s, want one Starlark can fill with the values it is given",
					name, use.line, use.text, problem)
			}
		}
	}
	if formats == 0 {
		t.Fatal("no program formats a string with %, so nothing here was tested")
	}
}

// The check finds each kind of format Starlark refuses, and none that it
// fills: a check that could never fail would pass every program above.
func TestAFormatStarlarkCannotFillIsFound(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		use     formatUse
		refused bool
	}{
		{"a stray percent sign", formatUse{text: "a rise of 10% on %d", values: 1}, true},
		{"fewer values than conversions", formatUse{text: "%d moves from %d", values: 1}, true},
		{"more values than conversions", formatUse{text: "%d moves", values: 2}, true},
		{"a key never closed", formatUse{text: "%(share d", values: -1}, true},
		{"a percent sign at the end", formatUse{text: "a rise of 10%", values: 0}, true},
		{"a letter that is no conversion", formatUse{text: "%q", values: 1}, true},
		{"a percent sign written twice", formatUse{text: "%d%%", values: 1}, false},
		{"conversions named by keys", formatUse{text: "%(share)d%(sign)%", values: -1}, false},
		{"values known only when it runs", formatUse{text: "%d from the row of %d", values: -1}, false},
		{"as many values as conversions", formatUse{text: "%d : %d", values: 2}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			problem := formatProblem(test.use)
			if refused := problem != ""; refused != test.refused {
				t.Errorf("formatProblem(%q, %d values) = %q, want refused %v", test.use.text, test.use.values, problem, test.refused)
			}
		})
	}
}

// formatUse is a string a program formats with %: its text, the line it is
// used on, and how many values it is given, or -1 when that shows only when
// the program runs.
type formatUse struct {
	text   string
	line   int32
	values int
}

// formatsIn is every string a program formats with %: a literal written
// where it is used, or one kept in a constant, a top-level name in capitals.
// An augmented %= is not looked at; no shipped program formats that way.
func formatsIn(t *testing.T, name, program string) []formatUse {
	t.Helper()

	file, err := starlark.Parse(name, program)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	constants := constantStrings(file)
	var found []formatUse
	syntax.Walk(file, func(node syntax.Node) bool {
		binary, isBinary := node.(*syntax.BinaryExpr)
		if !isBinary || binary.Op != syntax.PERCENT {
			return true
		}
		if text, isString := stringOf(unwrapped(binary.X), constants); isString {
			found = append(found, formatUse{text: text, line: binary.OpPos.Line, values: valuesIn(unwrapped(binary.Y))})
		}
		return true
	})
	return found
}

// constantStrings are the strings a program keeps in constants, as in
// FROM_ONE = "%d from the row of %d". Only names in capitals count: a lower
// case name may be a local that shadows it.
func constantStrings(file *syntax.File) map[string]string {
	constants := map[string]string{}
	for _, statement := range file.Stmts {
		assign, isAssign := statement.(*syntax.AssignStmt)
		if !isAssign || assign.Op != syntax.EQ {
			continue
		}
		ident, isIdent := assign.LHS.(*syntax.Ident)
		if !isIdent || strings.ToUpper(ident.Name) != ident.Name || !strings.ContainsFunc(ident.Name, unicode.IsLetter) {
			continue
		}
		if text, isString := stringOf(unwrapped(assign.RHS), nil); isString {
			constants[ident.Name] = text
		}
	}
	return constants
}

// stringOf is the text of a string literal, or of a constant holding one.
func stringOf(expr syntax.Expr, constants map[string]string) (string, bool) {
	switch expr := expr.(type) {
	case *syntax.Literal:
		text, isString := expr.Value.(string)
		return text, isString && expr.Token == syntax.STRING
	case *syntax.Ident:
		text, isConstant := constants[expr.Name]
		return text, isConstant
	}
	return "", false
}

// unwrapped is an expression without the brackets around it.
func unwrapped(expr syntax.Expr) syntax.Expr {
	for {
		paren, isParen := expr.(*syntax.ParenExpr)
		if !isParen {
			return expr
		}
		expr = paren.X
	}
}

// valuesIn is how many values the right of a % gives the format, where that
// is certain from the text alone: the length of a tuple written out, and one
// for a literal or a built-in that never gives a tuple. Anything else — a
// name, an index, a call, a sum — may hold a tuple, and is -1.
func valuesIn(expr syntax.Expr) int {
	switch expr := expr.(type) {
	case *syntax.TupleExpr:
		return len(expr.List)
	case *syntax.Literal:
		return 1
	case *syntax.CallExpr:
		if callee, isIdent := expr.Fn.(*syntax.Ident); isIdent && slices.Contains(scalarBuiltins, callee.Name) {
			return 1
		}
	}
	return -1
}

// formatProblem says what stops Starlark filling a format with the values it
// is given, or nothing when it can: every % has to begin %%, a conversion, or
// a key in brackets and then a conversion, and where the number of values is
// known there have to be as many as there are conversions.
func formatProblem(use formatUse) string {
	positional, keyed := 0, false
	for i := 0; i < len(use.text); i++ {
		if use.text[i] != '%' {
			continue
		}
		end, kind, problem := conversionAt(use.text, i)
		if problem != "" {
			return problem
		}
		switch kind {
		case positionalConversion:
			positional++
		case keyedConversion:
			keyed = true
		}
		i = end
	}
	if !keyed && use.values >= 0 && use.values != positional {
		return fmt.Sprintf("asks for %d values and is given %d", positional, use.values)
	}
	return ""
}

// The kinds of thing a % begins in a format.
const (
	percentSign          = "a percent sign"
	positionalConversion = "a conversion of the next value"
	keyedConversion      = "a conversion of a value named by a key"
)

// conversionAt reads what the % at start begins: where it ends, which kind
// it is, and what stops Starlark filling it, if anything does. After a key a
// second % is a conversion like any other, and Starlark writes it as a sign.
func conversionAt(text string, start int) (end int, kind, problem string) {
	i := start + 1
	if i < len(text) && text[i] == '%' {
		return i, percentSign, ""
	}
	kind = positionalConversion
	if i < len(text) && text[i] == '(' {
		closing := strings.IndexByte(text[i:], ')')
		if closing < 0 {
			return i, "", "opens a key in brackets and never closes it"
		}
		i += closing + 1
		kind = keyedConversion
	}
	if i == len(text) {
		return i, "", "ends in the middle of a %"
	}
	letter, size := utf8.DecodeRuneInString(text[i:])
	if !strings.ContainsRune(conversions, letter) && (kind != keyedConversion || letter != '%') {
		return i, "", fmt.Sprintf("has %q, which Starlark cannot fill; a %% sign is written %%%%", text[start:i+size])
	}
	return i, kind, ""
}
