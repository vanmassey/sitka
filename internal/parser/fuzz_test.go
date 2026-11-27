package parser

import (
	"testing"

	"github.com/vanmassey/sitka/internal/lexer"
)

func FuzzLexAndParseNeverPanics(f *testing.F) {
	for _, source := range []string{
		"",
		"let value = [1, 2, 3]",
		"fun square(value) { return value * value }",
		"try { throw \"problem\" } catch err { print err }",
		"struct Point { x, y }\nPoint(2, 3)",
		"/* nested-ish * comment */ print \"🌲\"",
		string([]byte{0xff, 0xfe, 0xfd}),
	} {
		f.Add(source)
	}
	f.Fuzz(func(t *testing.T, source string) {
		tokens, err := lexer.Lex(source)
		if err != nil {
			return
		}
		_, _ = Parse(tokens)
	})
}
