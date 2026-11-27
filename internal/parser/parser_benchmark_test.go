package parser

import (
	"testing"

	"github.com/vanmassey/sitka/internal/lexer"
)

const parserBenchmarkSource = `
fun report(values) {
    let total = 0
    for value in values {
        if value > 0 {
            total = total + value
        } else {
            print "ignored", value
        }
    }
    return {"total": total, "count": len(values)}
}
let result = report([1, 2, 3, 4, 5, 6, 7, 8, 9])
print result.total
`

func BenchmarkParseProgram(b *testing.B) {
	tokens, err := lexer.Lex(parserBenchmarkSource)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := Parse(tokens); err != nil {
			b.Fatal(err)
		}
	}
}
