package starlark

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"go.starlark.net/resolve"
	"go.starlark.net/syntax"
)

// conversions are the letters Starlark fills after a %. It has no flags, no
// width and no precision, so anything else stops the program where it formats.
const conversions = "sdioxXeEfFgGcr"

// scalarBuiltins are the built-ins whose result a program formats and which
// always give one value, never a tuple.
var scalarBuiltins = []string{"len", "str", "int", "sum", "abs"}

// UnfillableFormat says what stops Starlark filling a string a program formats
// with %, anywhere in it and whether or not a run would reach it, or nothing
// when it can fill every one the source shows. A program others start from has
// to be clean throughout: a stray % in a fail() written for the case that
// should never happen breaks that message exactly when it is needed.
//
// The source is read exactly as a run reads it, so that a local that shadows a
// constant, or a function of the program's own named after a built-in, is told
// apart here just as a run tells it apart. The error is for a program a run
// would refuse before it began.
//
// A run is never refused over this. A stray % the run does not reach stops
// nothing, and a program that proves its answer is not turned away for it; a
// run is told only about the format it stopped at.
//
// Only what the source shows is read: a string written where it is formatted,
// or one a name holds throughout the program. A format built while the program
// runs is left to the run.
func UnfillableFormat(name, source string) (string, error) {
	// Reading needs the vocabulary's names alone, and a name does not depend on
	// the budget its helper charges.
	declared, err := vocabulary(1)
	if err != nil {
		return "", err
	}
	// Read under the name a run gives every program, which a refusal leaves out
	// of its sentence: the error names the program once, by its own name.
	file, _, refusal := compile(programName, source, declared.Has)
	if refusal != "" {
		return "", fmt.Errorf("starlark: read %s: %s", name, refusal)
	}
	for _, use := range formatsIn(file) {
		if problem := formatProblem(use); problem != "" {
			return use.described(problem), nil
		}
	}
	return "", nil
}

// unfillableFormatAt says what stops Starlark filling the string the program
// formats with the % at a position, or nothing when there is no format there
// or it can be filled. A run stops at the position of the operator it could
// not carry out, so this is how a stopped run is told apart from any other
// failure without reading the interpreter's own words, which it may reword.
func unfillableFormatAt(file *syntax.File, at syntax.Position) string {
	for _, use := range formatsIn(file) {
		if use.at.Line != at.Line || use.at.Col != at.Col {
			continue
		}
		if problem := formatProblem(use); problem != "" {
			return use.described(problem)
		}
	}
	return ""
}

// formatUse is a string a program formats with %: its text, where the % is,
// and how many values it is given, or -1 when that shows only when the
// program runs.
type formatUse struct {
	text   string
	at     syntax.Position
	values int
}

// described is one sentence about a problem of this format, saying where it is
// without quoting the string: it is the program's own text, and may hold an
// option's.
func (use formatUse) described(problem string) string {
	return fmt.Sprintf("line %d: a string formatted with %% %s", use.at.Line, problem)
}

// formatsIn is every string a program formats with %: a literal written where
// it is used, or one a name holds throughout the program. An augmented %= is
// left to the run: the string it formats is whatever its name holds by then.
// The program has to be one the interpreter has bound the names of.
func formatsIn(file *syntax.File) []formatUse {
	constants := constantStrings(file)
	var found []formatUse
	syntax.Walk(file, func(node syntax.Node) bool {
		binary, isBinary := node.(*syntax.BinaryExpr)
		if !isBinary || binary.Op != syntax.PERCENT {
			return true
		}
		if text, isString := stringOf(unwrapped(binary.X), constants); isString {
			found = append(found, formatUse{text: text, at: binary.OpPos, values: valuesIn(unwrapped(binary.Y))})
		}
		return true
	})
	return found
}

// constantStrings are the strings a program keeps in names at its top level,
// as in FROM_ONE = "%d from the row of %d": a name bound there once, by a
// plain =, to a string written out. A name bound more than once — assigned
// again, made longer with +=, a loop's variable, a function — holds different
// things at different points of a run, so it holds no one string. Which
// bindings are the top-level name is the interpreter's call: a local of the
// same name, inside a function, is another name.
func constantStrings(file *syntax.File) map[string]string {
	texts, bindings := map[string]string{}, map[string]int{}
	syntax.Walk(file, func(node syntax.Node) bool {
		switch node := node.(type) {
		case *syntax.DefStmt:
			countGlobal(node.Name, bindings)
		case *syntax.ForStmt:
			countGlobal(node.Vars, bindings)
		case *syntax.AssignStmt:
			countGlobal(node.LHS, bindings)
			if name, isName := unwrapped(node.LHS).(*syntax.Ident); isName && node.Op == syntax.EQ && global(name) {
				if text, isString := stringOf(unwrapped(node.RHS), nil); isString {
					texts[name.Name] = text
				}
			}
		}
		return true
	})

	constants := map[string]string{}
	for name, text := range texts {
		if bindings[name] == 1 {
			constants[name] = text
		}
	}
	return constants
}

// countGlobal counts each top-level name an assignment, a loop or a function
// binds: the name itself, or every name of a tuple or a list it unpacks into.
// An element or a field assigned to binds no name.
func countGlobal(target syntax.Expr, bindings map[string]int) {
	switch target := target.(type) {
	case *syntax.Ident:
		if global(target) {
			bindings[target.Name]++
		}
	case *syntax.ParenExpr:
		countGlobal(target.X, bindings)
	case *syntax.TupleExpr:
		for _, element := range target.List {
			countGlobal(element, bindings)
		}
	case *syntax.ListExpr:
		for _, element := range target.List {
			countGlobal(element, bindings)
		}
	}
}

// global says whether the interpreter has bound a name at the top level of the
// program, rather than inside a function.
func global(name *syntax.Ident) bool {
	bound, isBound := name.Binding.(*resolve.Binding)
	return isBound && bound.Scope == resolve.Global
}

// stringOf is the text of a string literal, or of the string a name holds
// throughout the program. A name counts only where it is the top-level one: a
// local of the same name holds a string of its own, which the source does not
// show.
func stringOf(expr syntax.Expr, constants map[string]string) (string, bool) {
	switch expr := expr.(type) {
	case *syntax.Literal:
		text, isString := expr.Value.(string)
		return text, isString && expr.Token == syntax.STRING
	case *syntax.Ident:
		if !global(expr) {
			return "", false
		}
		text, isConstant := constants[expr.Name]
		return text, isConstant
	}
	return "", false
}

// builtIn says whether a name calls the built-in it is named after, as the
// interpreter has bound it, rather than a function the program binds to that
// name itself.
func builtIn(name *syntax.Ident) bool {
	bound, isBound := name.Binding.(*resolve.Binding)
	return isBound && (bound.Scope == resolve.Predeclared || bound.Scope == resolve.Universal)
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
// for a literal or a call of a built-in that never gives a tuple. Anything
// else — a name, an index, a call, a sum — may hold a tuple, and is -1.
func valuesIn(expr syntax.Expr) int {
	switch expr := expr.(type) {
	case *syntax.TupleExpr:
		return len(expr.List)
	case *syntax.Literal:
		return 1
	case *syntax.CallExpr:
		if callee, isIdent := expr.Fn.(*syntax.Ident); isIdent && slices.Contains(scalarBuiltins, callee.Name) && builtIn(callee) {
			return 1
		}
	}
	return -1
}

// formatProblem says what stops Starlark filling a format with the values it
// is given, or nothing when it can: every % has to begin %%, a conversion, or
// a key in brackets and then a conversion, and where the number of values is
// known there have to be as many as there are conversions and no key at all.
//
// A format with a key is filled from a dictionary, and a conversion of the
// next value is then handed the dictionary itself. That works once, as the
// first conversion of all, and only for one that writes a value out whole —
// %s or %r; anywhere else no value can fill it.
func formatProblem(use formatUse) string {
	converted, positional, keyed, wholeFirst := 0, 0, false, true
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
			wholeFirst = wholeFirst && converted == 0 && strings.ContainsRune("sr", rune(use.text[end]))
			positional++
			converted++
		case keyedConversion:
			keyed = true
			converted++
		}
		i = end
	}
	switch {
	case keyed && use.values >= 0:
		// Values that can be counted are a tuple, a number or a string, and a
		// key can be looked up only in a dictionary.
		return "names a value by a key, and is given no dictionary to find it in"
	case keyed && positional > 0 && !wholeFirst:
		return "names values by keys, and asks for a value by its place too, which a dictionary cannot give"
	case !keyed && use.values >= 0 && use.values != positional:
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

// conversionAt reads what the % at start begins: where it ends, which kind it
// is, and what stops Starlark filling it, if anything does. After a key a
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
		// A key is the program's own words, and stays out of the sentence.
		shown := text[start : i+size]
		if kind == keyedConversion {
			shown = "%(…)" + text[i:i+size]
		}
		return i, "", fmt.Sprintf("has %q, which Starlark cannot fill; a %% sign is written %%%%", shown)
	}
	return i, kind, ""
}
