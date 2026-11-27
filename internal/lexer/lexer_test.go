package lexer

import "testing"

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
