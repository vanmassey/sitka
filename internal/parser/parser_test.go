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
