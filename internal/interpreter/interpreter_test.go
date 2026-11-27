package interpreter

import (
	"testing"

	"github.com/vanmassey/sitka/internal/lexer"
	"github.com/vanmassey/sitka/internal/parser"
)

func executeForTest(source string) (Value, error) {
	tokens, err := lexer.Lex(source)
	if err != nil {
		return nil, err
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		return nil, err
	}
	return New().Execute(program)
}

func TestVariablesAndArithmetic(t *testing.T) {
	value, err := executeForTest("let x = 2\nlet y = 3\nx + y * 4")
	if err != nil {
		t.Fatal(err)
	}
	if value != int64(14) {
		t.Fatalf("got %v, want 14", value)
	}
}

func TestFunctionsAndCalls(t *testing.T) {
	value, err := executeForTest("fun add(a, b) { return a + b }\nadd(10, 5)")
	if err != nil {
		t.Fatal(err)
	}
	if value != int64(15) {
		t.Fatalf("got %v, want 15", value)
	}
}

func TestForLoopAndList(t *testing.T) {
	value, err := executeForTest("let total = 0\nfor n in [1, 2, 3] { total = total + n }\ntotal")
	if err != nil {
		t.Fatal(err)
	}
	if value != int64(6) {
		t.Fatalf("got %v, want 6", value)
	}
}

func TestLoopBreakAndContinue(t *testing.T) {
	value, err := executeForTest(`
		let total = 0
		for number in range(10) {
			if number == 2 {
				continue
			}
			if number == 5 {
				break
			}
			total = total + number
		}
		total
	`)
	if err != nil {
		t.Fatal(err)
	}
	if value != int64(8) {
		t.Fatalf("got %v, want 8", value)
	}
}

func TestLogicalOperatorsShortCircuit(t *testing.T) {
	value, err := executeForTest(`
		let calls = 0
		fun mark() {
			calls = calls + 1
			return true
		}
		false && mark()
		true || mark()
		calls
	`)
	if err != nil {
		t.Fatal(err)
	}
	if value != int64(0) {
		t.Fatalf("logical operators evaluated a short-circuited branch: got %v", value)
	}
}

func TestBlockScopeAndOuterAssignment(t *testing.T) {
	value, err := executeForTest(`
		let value = 1
		{
			let hidden = 99
			value = value + 2
		}
		value
	`)
	if err != nil {
		t.Fatal(err)
	}
	if value != int64(3) {
		t.Fatalf("got %v, want 3", value)
	}
}

func TestListIndexAssignmentAndNegativeIndex(t *testing.T) {
	value, err := executeForTest(`
		let numbers = [10, 20, 30]
		numbers[-1] = 40
		[numbers[0], numbers[-1]]
	`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), "[10, 40]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestRejectsDivisionByZero(t *testing.T) {
	_, err := executeForTest("10 / 0")
	if err == nil {
		t.Fatal("Execute accepted division by zero")
	}
}

func TestHigherOrderListFunctions(t *testing.T) {
	value, err := executeForTest(`
		fun square(number) { return number * number }
		fun add(total, number) { return total + number }
		fun isEven(number) { return number % 2 == 0 }
		let squares = map([1, 2, 3, 4], square)
		let evenSquares = filter(squares, isEven)
		reduce(evenSquares, add, 0)
	`)
	if err != nil {
		t.Fatal(err)
	}
	if value != int64(20) {
		t.Fatalf("got %v, want 20", value)
	}
}

func TestAnonymousFunctionsAsCallbacks(t *testing.T) {
	value, err := executeForTest(`
		let base = 5
		let values = map([1, 2, 3], fun(number) {
			return number + base
		})
		let large = filter(values, fun(number) {
			return number > 6
		})
		large
	`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), "[7, 8]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestExtendedCollectionAndStringFunctions(t *testing.T) {
	value, err := executeForTest(`
		fun capitalize(word) { return upper(word) }
		let words = split("sitka;is;joyful", ";")
		let labels = map(words, capitalize)
		let joined = join(labels, " ")
		let uniqueNumbers = unique([3, 1, 3, 2, 1])
		let ordered = sorted(uniqueNumbers)
		[joined, ordered, sum(ordered), slice(ordered, 1)]
	`)
	if err != nil {
		t.Fatal(err)
	}
	got := Format(value)
	want := `[SITKA IS JOYFUL, [1, 2, 3], 6, [2, 3]]`
	if got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestMapIterationOrderAndPredicates(t *testing.T) {
	value, err := executeForTest(`
		let mapping = {"z": 3, "a": 1, "m": 2}
		[keys(mapping), values(mapping), all([true, 1]), any([false, nil, "yes"])]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[[a, m, z], [1, 2, 3], true, true]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestUnicodeStringHelpers(t *testing.T) {
	value, err := executeForTest(`
		let letters = chars("雪だるま")
		[join(reversed(letters), ""), len("雪だるま")]
	`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), `[まるだ雪, 4]`; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestBuiltinsReportInvalidArguments(t *testing.T) {
	for _, source := range []string{
		`map([1], 2)`,
		`sorted([1, "two"])`,
		`range(0, 3, 0)`,
		`assert(false, "not true")`,
		`sum([1, "two"])`,
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := executeForTest(source); err == nil {
				t.Fatalf("expected %q to report an error", source)
			}
		})
	}
}

func TestCollectionHelpersDoNotMutateInputLists(t *testing.T) {
	value, err := executeForTest(`
		let original = [1, 2]
		let appended = push(original, 3)
		let shortened = pop(original)
		[original, appended, shortened]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[[1, 2], [1, 2, 3], [1]]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestRangeRejectsExcessiveOutput(t *testing.T) {
	_, err := executeForTest("range(0, 1000001)")
	if err == nil {
		t.Fatal("range accepted output above its configured limit")
	}
}

func TestMixedNumericOperations(t *testing.T) {
	value, err := executeForTest("1 + 2.5 * 2")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), "6"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestSignedIntegerMultiplicationAndOverflow(t *testing.T) {
	for _, test := range []struct {
		source string
		want   string
	}{
		{source: "-2 * -3", want: "6"},
		{source: "-2 * 3", want: "-6"},
		{source: "2 * -3", want: "-6"},
		{source: "(-9223372036854775807 - 1) * -1", want: ""},
		{source: "9223372036854775807 * 2", want: ""},
	} {
		value, err := executeForTest(test.source)
		if test.want == "" {
			if err == nil {
				t.Errorf("%s unexpectedly returned %v", test.source, value)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s returned error: %v", test.source, err)
			continue
		}
		if got := Format(value); got != test.want {
			t.Errorf("%s = %s, want %s", test.source, got, test.want)
		}
	}
}

func TestMapEqualityIsStructural(t *testing.T) {
	value, err := executeForTest(`{"a": [1, 2], "b": true} == {"b": true, "a": [1, 2]}`)
	if err != nil {
		t.Fatal(err)
	}
	if value != true {
		t.Fatalf("got %v, want true", value)
	}
}

func TestFloatToIntegerRange(t *testing.T) {
	for _, source := range []string{
		`int(9223372036854777856.0)`,
		`int(-9223372036854777856.0)`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly converted out-of-range value", source)
		}
	}
}

func TestNestedClosuresCaptureAndAssignOuterVariables(t *testing.T) {
	value, err := executeForTest(`
		fun counter() {
			let value = 0
			fun increment() {
				value = value + 1
				return value
			}
			return increment
		}
		let next = counter()
		[next(), next(), next()]
	`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), "[1, 2, 3]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}
