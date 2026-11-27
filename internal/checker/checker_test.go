package checker

import (
	"strings"
	"testing"

	"github.com/vanmassey/sitka/internal/lexer"
	"github.com/vanmassey/sitka/internal/parser"
)

func checkSource(t *testing.T, source string) []Diagnostic {
	t.Helper()
	tokens, err := lexer.Lex(source)
	if err != nil {
		t.Fatal(err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	return Check(program)
}

func hasDiagnostic(diagnostics []Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func TestAcceptsValidProgramWithBuiltinsAndClosures(t *testing.T) {
	diagnostics := checkSource(t, `
		let base = 10
		fun makeAdder(amount) {
			return fun(value) { return value + amount + base }
		}
		let addFive = makeAdder(5)
		map([1, 2, 3], addFive)
	`)
	if len(diagnostics) != 0 {
		t.Fatalf("valid program produced diagnostics: %#v", diagnostics)
	}
}

func TestFindsUndefinedName(t *testing.T) {
	diagnostics := checkSource(t, "print missing")
	if !hasDiagnostic(diagnostics, "undefined-name") {
		t.Fatalf("expected undefined-name diagnostic, got %#v", diagnostics)
	}
	if diagnostics[0].Line != 1 {
		t.Fatalf("diagnostic line = %d, want 1", diagnostics[0].Line)
	}
}

func TestFindsDuplicateDeclarations(t *testing.T) {
	diagnostics := checkSource(t, "let value = 1\nlet value = 2")
	if !hasDiagnostic(diagnostics, "duplicate-name") {
		t.Fatalf("expected duplicate-name diagnostic, got %#v", diagnostics)
	}
}

func TestFindsFunctionArityMismatch(t *testing.T) {
	diagnostics := checkSource(t, "fun greet(name) { return name }\ngreet()")
	if !hasDiagnostic(diagnostics, "wrong-arity") {
		t.Fatalf("expected wrong-arity diagnostic, got %#v", diagnostics)
	}
}

func TestFindsControlFlowOutsideValidContext(t *testing.T) {
	diagnostics := checkSource(t, "break\ncontinue\nreturn 1")
	for _, code := range []string{"break-outside-loop", "continue-outside-loop", "return-outside-function"} {
		if !hasDiagnostic(diagnostics, code) {
			t.Errorf("missing %s diagnostic: %#v", code, diagnostics)
		}
	}
}

func TestLoopControlIsInvalidInsideNestedFunction(t *testing.T) {
	diagnostics := checkSource(t, `
		while true {
			fun invalid() {
				break
			}
		}
	`)
	if !hasDiagnostic(diagnostics, "break-outside-loop") {
		t.Fatalf("expected nested break diagnostic, got %#v", diagnostics)
	}
}

func TestAllowsLexicalShadowingButRejectsSameScopeDuplicates(t *testing.T) {
	diagnostics := checkSource(t, `
		let value = 1
		{
			let value = 2
			print value
		}
	`)
	if len(diagnostics) != 0 {
		t.Fatalf("shadowing produced diagnostics: %#v", diagnostics)
	}
}

func TestBuiltinsCanBeShadowed(t *testing.T) {
	diagnostics := checkSource(t, "let len = 42\nprint len")
	if len(diagnostics) != 0 {
		t.Fatalf("builtin shadowing produced diagnostics: %#v", diagnostics)
	}
}

func TestImportAliasIsAvailableToChecker(t *testing.T) {
	diagnostics := checkSource(t, `
		import "math.sitka" as math
		math.double(21)
	`)
	if len(diagnostics) != 0 {
		t.Fatalf("valid import produced diagnostics: %#v", diagnostics)
	}
}

func TestChecksStructDeclarationsAndConstructorArity(t *testing.T) {
	diagnostics := checkSource(t, `
		struct Point { x, y }
		let origin = Point(0)
	`)
	if !hasDiagnostic(diagnostics, "wrong-arity") {
		t.Fatalf("expected struct constructor arity diagnostic, got %#v", diagnostics)
	}
}

func TestFindsDuplicateStructFields(t *testing.T) {
	diagnostics := checkSource(t, "struct Point { x, x }")
	if !hasDiagnostic(diagnostics, "duplicate-field") {
		t.Fatalf("expected duplicate-field diagnostic, got %#v", diagnostics)
	}
}

func TestChecksTryCatchScopesAndThrownExpressions(t *testing.T) {
	diagnostics := checkSource(t, `
		try {
			throw missing
		} catch problem {
			print problem
		}
		print problem
	`)
	if !hasDiagnostic(diagnostics, "undefined-name") {
		t.Fatalf("expected undefined-name diagnostic, got %#v", diagnostics)
	}
}

func TestChecksRuntimeArgumentsAndNewStandardLibraryNames(t *testing.T) {
	diagnostics := checkSource(t, `
		print args[0]
		print csv_decode("a,b")
		print regex_match("a", "a")
		print time_unix(2026, 1, 1)
		print sha256("sitka")
		print pad_left("x", 3, "0")
	`)
	if len(diagnostics) != 0 {
		t.Fatalf("known runtime names produced diagnostics: %#v", diagnostics)
	}
}

func TestLoopAndFunctionContextsAreRestoredAfterTryCatch(t *testing.T) {
	diagnostics := checkSource(t, `
		fun find() {
			while true {
				try {
					break
				} catch problem {
					continue
				}
			}
			return 1
		}
		find()
	`)
	if len(diagnostics) != 0 {
		t.Fatalf("valid control flow produced diagnostics: %#v", diagnostics)
	}
}

func TestDiagnosticsAreStableAndReadable(t *testing.T) {
	diagnostics := checkSource(t, "print unknown\nbreak")
	if len(diagnostics) < 2 {
		t.Fatalf("got %d diagnostics, want at least 2", len(diagnostics))
	}
	seen := map[string]bool{}
	for _, diagnostic := range diagnostics {
		text := diagnostic.String()
		if !strings.Contains(text, diagnostic.Code) || !strings.Contains(text, string(diagnostic.Severity)) {
			t.Fatalf("diagnostic not rendered with code and severity: %q", text)
		}
		seen[diagnostic.Code] = true
	}
	if !seen["undefined-name"] || !seen["break-outside-loop"] {
		t.Fatalf("unexpected diagnostics: %#v", diagnostics)
	}
}
