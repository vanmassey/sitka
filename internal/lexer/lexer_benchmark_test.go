package lexer

import "testing"

const benchmarkSource = `
// A source fragment large enough to exercise all common token paths.
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

func BenchmarkLexProgram(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		if _, err := Lex(benchmarkSource); err != nil {
			b.Fatal(err)
		}
	}
}
