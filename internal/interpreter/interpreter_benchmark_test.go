package interpreter

import (
	"testing"

	"github.com/vanmassey/sitka/internal/lexer"
	"github.com/vanmassey/sitka/internal/parser"
)

func BenchmarkExecuteArithmetic(b *testing.B) {
	tokens, err := lexer.Lex("let value = (12 + 30) * 2 - 7\nvalue / 5")
	if err != nil {
		b.Fatal(err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		b.Fatal(err)
	}
	runtime := New()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := runtime.Execute(program); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecuteCollectionPipeline(b *testing.B) {
	tokens, err := lexer.Lex(`
		let values = range(100)
		let selected = filter(values, fun(value) { return value % 3 == 0 })
		let squares = map(selected, fun(value) { return value * value })
		reduce(squares, fun(total, value) { return total + value }, 0)
	`)
	if err != nil {
		b.Fatal(err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		b.Fatal(err)
	}
	runtime := New()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := runtime.Execute(program); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkExecuteJSONEncoding(b *testing.B) {
	tokens, err := lexer.Lex(`json_encode({"name": "Sitka", "values": range(100)})`)
	if err != nil {
		b.Fatal(err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		b.Fatal(err)
	}
	runtime := New()
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := runtime.Execute(program); err != nil {
			b.Fatal(err)
		}
	}
}
