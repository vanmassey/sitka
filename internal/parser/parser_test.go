package parser

import (
	"testing"

	"github.com/vanmassey/sitka/internal/ast"
	"github.com/vanmassey/sitka/internal/lexer"
)

func TestRespectsOperatorPrecedence(t *testing.T) {
	tokens, err := lexer.Lex("2 + 3 * 4")
	if err != nil {
		t.Fatal(err)
	}
	program, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("got %d statements, want 1", len(program.Statements))
	}
	statement, ok := program.Statements[0].(ast.ExpressionStmt)
	if !ok {
		t.Fatalf("got %T, want ExpressionStmt", program.Statements[0])
	}
	addition, ok := statement.Expression.(ast.Binary)
	if !ok || addition.Operator != "+" {
		t.Fatalf("got %#v, want top-level addition", statement.Expression)
	}
	if _, ok := addition.Right.(ast.Binary); !ok {
		t.Fatalf("right operand is %T, want multiplication", addition.Right)
	}
}

func TestRejectsUnclosedParentheses(t *testing.T) {
	tokens, err := lexer.Lex("(1 + 2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(tokens); err == nil {
		t.Fatal("Parse accepted unclosed parentheses")
	}
}

func TestParsesFunctionAndControlFlow(t *testing.T) {
	source := `
		fun sum(values) {
			let total = 0
			for value in values {
				if value > 0 {
					total = total + value
				}
			}
			return total
		}
	`
	tokens, err := lexer.Lex(source)
	if err != nil {
		t.Fatal(err)
	}
	program, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("got %d statements, want 1 function declaration", len(program.Statements))
	}
	if _, ok := program.Statements[0].(ast.FunctionDecl); !ok {
		t.Fatalf("got %T, want FunctionDecl", program.Statements[0])
	}
}

func TestRejectsInvalidAssignmentTarget(t *testing.T) {
	tokens, err := lexer.Lex("(1 + 2) = 3")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Parse(tokens); err == nil {
		t.Fatal("Parse accepted an invalid assignment target")
	}
}

func TestParsesImportWithAlias(t *testing.T) {
	tokens, err := lexer.Lex(`import "math.sitka" as math`)
	if err != nil {
		t.Fatal(err)
	}
	program, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("got %d statements, want 1", len(program.Statements))
	}
	imported, ok := program.Statements[0].(ast.ImportStmt)
	if !ok {
		t.Fatalf("got %T, want ImportStmt", program.Statements[0])
	}
	if imported.Path != "math.sitka" || imported.Alias != "math" {
		t.Fatalf("unexpected import node: %#v", imported)
	}
}

func TestParsesStructDeclaration(t *testing.T) {
	tokens, err := lexer.Lex("struct Point { x, y }")
	if err != nil {
		t.Fatal(err)
	}
	program, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("got %d statements, want 1", len(program.Statements))
	}
	structure, ok := program.Statements[0].(ast.StructDecl)
	if !ok {
		t.Fatalf("got %T, want StructDecl", program.Statements[0])
	}
	if structure.Name != "Point" || len(structure.Fields) != 2 ||
		structure.Fields[0] != "x" || structure.Fields[1] != "y" {
		t.Fatalf("unexpected struct declaration: %#v", structure)
	}
}

func TestParsesTryCatchAndThrow(t *testing.T) {
	tokens, err := lexer.Lex(`
		try {
			throw "failed"
		}
		catch problem {
			print problem
		}
	`)
	if err != nil {
		t.Fatal(err)
	}
	program, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("got %d statements, want 1", len(program.Statements))
	}
	statement, ok := program.Statements[0].(ast.TryStmt)
	if !ok {
		t.Fatalf("got %T, want TryStmt", program.Statements[0])
	}
	if statement.ErrorName != "problem" || len(statement.Body.Statements) != 1 ||
		len(statement.Catch.Statements) != 1 {
		t.Fatalf("unexpected try statement: %#v", statement)
	}
	if _, ok := statement.Body.Statements[0].(ast.ThrowStmt); !ok {
		t.Fatalf("got %T, want ThrowStmt", statement.Body.Statements[0])
	}
}

func TestParserRejectsIncompleteDeclarations(t *testing.T) {
	for _, source := range []string{
		"fun missingBody(value)",
		"import \"module.sitka\" missing",
		"try { print 1 }",
		"try { print 1 } catch { print 2 }",
		"throw",
		"let value =",
	} {
		t.Run(source, func(t *testing.T) {
			tokens, err := lexer.Lex(source)
			if err != nil {
				return
			}
			if _, err := Parse(tokens); err == nil {
				t.Fatalf("Parse(%q) unexpectedly succeeded", source)
			}
		})
	}
}

func TestParsesPropertyAndIndexAssignmentTargets(t *testing.T) {
	tokens, err := lexer.Lex(`
		let point = {"x": 1}
		point.x = 2
		point["y"] = 3
	`)
	if err != nil {
		t.Fatal(err)
	}
	program, err := Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Statements) != 3 {
		t.Fatalf("got %d statements, want 3", len(program.Statements))
	}
	property, ok := program.Statements[1].(ast.ExpressionStmt)
	if !ok {
		t.Fatalf("property assignment statement = %T", program.Statements[1])
	}
	if _, ok := property.Expression.(ast.Set); !ok {
		t.Fatalf("property assignment expression = %T, want Set", property.Expression)
	}
	index, ok := program.Statements[2].(ast.ExpressionStmt)
	if !ok {
		t.Fatalf("index assignment statement = %T", program.Statements[2])
	}
	if _, ok := index.Expression.(ast.IndexAssign); !ok {
		t.Fatalf("index assignment expression = %T, want IndexAssign", index.Expression)
	}
}
