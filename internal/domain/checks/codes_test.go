package checks_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/MathTrail/mathtrail-standalone/internal/domain/checks"
)

// The list is how anything outside the package counts the checks, so a code
// declared and left out of it is a check nobody counts.
func TestTheListOfCodesHoldsEveryCodeDeclaredOnce(t *testing.T) {
	t.Parallel()

	declared := declaredCodes(t)
	listed := map[checks.Code]int{}
	for _, code := range checks.Codes() {
		listed[code]++
	}
	for _, code := range declared {
		if listed[code] != 1 {
			t.Errorf("Codes() lists %q %d times, want once", code, listed[code])
		}
	}
	if got := checks.Codes(); len(got) != len(declared) {
		t.Errorf("Codes() = %v, want the %d codes the package declares: %v", got, len(declared), declared)
	}
}

// declaredCodes are the constants of type Code in the package's own source,
// read as written rather than through the list they are held to.
func declaredCodes(t *testing.T) []checks.Code {
	t.Helper()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list the package's files: %v", err)
	}
	var codes []checks.Code
	fset := token.NewFileSet()
	for _, name := range files {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range file.Decls {
			codes = append(codes, codesIn(t, decl)...)
		}
	}
	if len(codes) == 0 {
		t.Fatal("the package declares no code, want the constants of type Code")
	}
	return codes
}

// codesIn are the codes one declaration holds.
func codesIn(t *testing.T, decl ast.Decl) []checks.Code {
	t.Helper()

	general, ok := decl.(*ast.GenDecl)
	if !ok || general.Tok != token.CONST {
		return nil
	}
	var codes []checks.Code
	for _, spec := range general.Specs {
		if value, ok := spec.(*ast.ValueSpec); ok && isCode(value.Type) {
			codes = append(codes, valuesOf(t, value)...)
		}
	}
	return codes
}

// isCode says whether a constant is declared of type Code.
func isCode(kind ast.Expr) bool {
	name, ok := kind.(*ast.Ident)
	return ok && name.Name == "Code"
}

// valuesOf are the codes of one line of constants, each a string literal.
func valuesOf(t *testing.T, value *ast.ValueSpec) []checks.Code {
	t.Helper()

	codes := make([]checks.Code, 0, len(value.Values))
	for i, expr := range value.Values {
		literal, ok := expr.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			t.Fatalf("%s is no string literal, want every code written out", value.Names[i].Name)
		}
		code, err := strconv.Unquote(literal.Value)
		if err != nil {
			t.Fatalf("unquote %s: %v", value.Names[i].Name, err)
		}
		codes = append(codes, checks.Code(code))
	}
	return codes
}
