package lexer

import (
	"strings"
	"testing"
)

func TestLexExpression(t *testing.T) {
	tokens, err := Lex("12 + (3 * 4)")
	if err != nil {
		t.Fatal(err)
	}

	kinds := []TokenKind{
		TokenInteger, TokenPlus, TokenLeftParen, TokenInteger,
		TokenStar, TokenInteger, TokenRightParen, TokenEnd,
	}
	if len(tokens) != len(kinds) {
		t.Fatalf("got %d tokens, want %d", len(tokens), len(kinds))
	}
	for index, kind := range kinds {
		if tokens[index].Kind != kind {
			t.Errorf("token %d kind = %v, want %v", index, tokens[index].Kind, kind)
		}
	}
}

func TestLexRejectsUnknownCharacters(t *testing.T) {
	if _, err := Lex("1 + @"); err == nil {
		t.Fatal("Lex accepted an unknown character")
	}
}

func TestLexesKeywordsStringsAndComments(t *testing.T) {
	tokens, err := Lex("let greeting = \"hello\\nworld\" // comment\n/* note */ print greeting")
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Kind != TokenLet {
		t.Fatalf("first token kind = %v, want let", tokens[0].Kind)
	}
	if tokens[3].Kind != TokenString || tokens[3].Value != "hello\nworld" {
		t.Fatalf("string token = %#v", tokens[3])
	}
	foundNewline := false
	for _, token := range tokens {
		if token.Kind == TokenNewline {
			foundNewline = true
		}
	}
	if !foundNewline {
		t.Fatal("lexer did not preserve the statement-separating newline")
	}
}

func TestLexRejectsMalformedNumbers(t *testing.T) {
	for _, source := range []string{"1..2", "9223372036854775808", "1 + \u0661"} {
		if _, err := Lex(source); err == nil {
			t.Errorf("Lex(%q) unexpectedly succeeded", source)
		}
	}
}

func TestLexKeywordsAndPositions(t *testing.T) {
	source := "try {\n  throw \"🌲\"\n} catch problem { print problem }\nlet café = 1"
	tokens, err := Lex(source)
	if err != nil {
		t.Fatal(err)
	}
	found := map[TokenKind]bool{}
	for _, token := range tokens {
		switch token.Kind {
		case TokenTry, TokenCatch, TokenThrow:
			found[token.Kind] = true
		}
		if token.Text == "café" {
			if token.Line != 4 {
				t.Errorf("café token line = %d, want 4", token.Line)
			}
			if token.Byte != strings.LastIndex(source, "café") {
				t.Errorf("café token byte = %d, want %d", token.Byte, strings.LastIndex(source, "café"))
			}
		}
	}
	for _, kind := range []TokenKind{TokenTry, TokenCatch, TokenThrow} {
		if !found[kind] {
			t.Errorf("missing keyword token %s", kind)
		}
	}
}

func TestLexSuppressesContinuationNewlines(t *testing.T) {
	tokens, err := Lex("let values = [1,\n2 + 3]\nprint values")
	if err != nil {
		t.Fatal(err)
	}
	newlines := 0
	for _, token := range tokens {
		if token.Kind == TokenNewline {
			newlines++
		}
	}
	if newlines != 1 {
		t.Fatalf("got %d statement-separating newlines, want 1", newlines)
	}
}

func TestLexRejectsInvalidUTF8AndUnterminatedStrings(t *testing.T) {
	for _, source := range []string{
		string([]byte{0xff}),
		`"missing end`,
		`/* missing end`,
	} {
		if _, err := Lex(source); err == nil {
			t.Errorf("Lex(%q) unexpectedly succeeded", source)
		}
	}
}
