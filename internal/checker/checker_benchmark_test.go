package checker

import (
	"testing"

	"github.com/vanmassey/sitka/internal/lexer"
	"github.com/vanmassey/sitka/internal/parser"
)

func BenchmarkCheckProgram(b *testing.B) {
	tokens, err := lexer.Lex(`
		fun sum(values) {
			let total = 0
			for value in values {
				if value >= 0 {
					total = total + value
				}
			}
			return total
		}
		let result = sum(range(100))
		print result
	`)
	if err != nil {
		b.Fatal(err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = Check(program)
	}
}
