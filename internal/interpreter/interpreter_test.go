package interpreter

import (
	"os"
	"path/filepath"
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

func TestRuntimeArguments(t *testing.T) {
	tokens, err := lexer.Lex("args")
	if err != nil {
		t.Fatal(err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		t.Fatal(err)
	}
	value, err := NewWithArgs([]string{"first", "second"}).Execute(program)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), "[first, second]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
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

func TestStructConstructionPropertiesAndEquality(t *testing.T) {
	value, err := executeForTest(`
		struct Person {
			name,
			age
		}
		let person = Person("Ada", 37)
		person.age = person.age + 1
		[
			person.name,
			person["age"],
			type(person),
			person == Person("Ada", 38),
			json_encode(person)
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[Ada, 38, struct Person, true, {"age":38,"name":"Ada"}]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestStructConstructionAndFieldErrors(t *testing.T) {
	for _, source := range []string{
		"struct Point { x, y }\nPoint(1)",
		"struct Point { x }\nPoint(1).z",
		"struct Point { x }\nlet point = Point(1)\npoint.z = 2",
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestCSVDecodeEncodeAndMapRecords(t *testing.T) {
	value, err := executeForTest(`
			let rows = csv_decode("name,age\nAda,36\nLin,28\n")
			let people = csv_decode_maps("name,age\nAda,36\nLin,28\n")
			[
				rows,
				people[0].name,
				csv_encode([["name", "note"], ["Ada", "hello, world"]]),
				csv_encode_maps(people, ["name", "age"])
			]
		`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[[[name, age], [Ada, 36], [Lin, 28]], Ada, name,note
Ada,"hello, world"
, name,age
Ada,36
Lin,28
]`
	if got := Format(value); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCSVMapHeadersAreDeterministic(t *testing.T) {
	value, err := executeForTest(`
			csv_encode_maps([{"z": 2, "a": 1}, {"a": 3, "z": 4}])
		`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := value, "a,z\n1,2\n3,4\n"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestCSVEmptyFieldRoundTripAndBlankLineHandling(t *testing.T) {
	encoded, err := encodeCSV([][]string{{""}})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := encoded, "\"\"\n"; got != want {
		t.Fatalf("encoded empty field = %q, want %q", got, want)
	}
	decoded, err := decodeCSV(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 1 || len(decoded[0]) != 1 || decoded[0][0] != "" {
		t.Fatalf("decoded empty field = %#v, want one empty field", decoded)
	}
	blank, err := decodeCSV("\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(blank) != 0 {
		t.Fatalf("blank CSV line produced records %#v", blank)
	}
}

func TestCSVMapDecoderReturnsEmptyListForEmptyInput(t *testing.T) {
	value, err := executeForTest(`csv_decode_maps("")`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), "[]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestCSVRejectsMalformedAndUnsupportedValues(t *testing.T) {
	for _, source := range []string{
		`csv_decode("\"unterminated")`,
		`csv_decode_maps("name,name\nAda,Ada\n")`,
		`csv_decode_maps("name,age\nAda\n")`,
		`csv_encode([[{"nested": true}]])`,
		`csv_encode_maps([{"extra": 1}], ["known"])`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestRegexBuiltins(t *testing.T) {
	value, err := executeForTest(`
			[
				regex_match("user-42", "^[a-z]+-[0-9]+$"),
				regex_find("id=42; id=7", "[0-9]+"),
				regex_find_all("one 2 three 4", "[0-9]+"),
				regex_replace("a1 b2", "[0-9]", "X"),
				regex_split("a, b,,c", ", *"),
				regex_captures("Ada Lovelace", "(?P<first>[A-Z][a-z]+) (?P<last>[A-Z][a-z]+)")
			]
		`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[true, 42, [2, 4], aX bX, [a, b, , c], {"first": Ada, "last": Lovelace}]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestRegexEmptyMatchAndInvalidPattern(t *testing.T) {
	value, err := executeForTest(`regex_find("abc", "x*")`)
	if err != nil {
		t.Fatal(err)
	}
	if value != "" {
		t.Fatalf("got %#v, want an empty string match", value)
	}
	for _, source := range []string{
		`regex_match("text", "[")`,
		`regex_replace("text", "(", "x")`,
		`regex_find(1, "x")`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestPathBuiltins(t *testing.T) {
	value, err := executeForTest(`
		[
			path_join("docs", "..", "README.md"),
			path_clean("docs/../README.md"),
			path_base("docs/language.md"),
			path_dir("docs/language.md"),
			path_ext("docs/language.md"),
			path_is_abs(path_abs(".")),
			path_rel("a/b", "a/c"),
			path_abs(".") == cwd(),
			path_separator
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[README.md, README.md, language.md, docs, .md, true, ../c, true, /]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestPathBuiltinsRejectWrongArguments(t *testing.T) {
	for _, source := range []string{
		`path_join()`,
		`path_join("a", 1)`,
		`path_base(42)`,
		`path_rel("a")`,
		`cwd("extra")`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestTimeBuiltinsRoundTripAndArithmetic(t *testing.T) {
	value, err := executeForTest(`
		let timestamp = time_unix(2024, 2, 29, 12, 30, 15)
		[
			time_utc(timestamp),
			time_format(timestamp, "2006-01-02"),
			time_diff(timestamp, time_add(timestamp, 1.5)),
			time_parts(timestamp).weekday,
			time_parse("2024-02-29", "2006-01-02") == time_unix(2024, 2, 29)
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[2024-02-29T12:30:15Z, 2024-02-29, 1.5, 4, true]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestTimeBuiltinsRejectInvalidInputs(t *testing.T) {
	for _, source := range []string{
		`time_unix(2023, 2, 29)`,
		`time_unix(2024, 13, 1)`,
		`time_parse("not a date", "2006-01-02")`,
		`time_format("now", "2006")`,
		`time_add(0, "tomorrow")`,
		`time_utc(1 / 0)`,
		`time_parts()`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestTimeBuiltinsPreserveNegativeFractionalTimestamps(t *testing.T) {
	value, err := executeForTest(`
		[
			time_utc(-0.5),
			time_utc(time_add(-0.5, 0.25)),
			time_diff(-0.5, 0.25)
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[1969-12-31T23:59:59.5Z, 1969-12-31T23:59:59.75Z, 0.75]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestCollectionTransformBuiltins(t *testing.T) {
	value, err := executeForTest(`
		let words = ["pear", "plum", "apple", "apricot"]
		let doubled = flat_map([1, 2], fun(value) { return [value, value * 10] })
		let grouped = group_by(words, fun(word) { return first(chars(word)) })
		let counts = count_by(words, fun(word) { return first(chars(word)) })
		[
			doubled,
			grouped.p,
			counts.a,
			sort_by(words, fun(word) { return len(word) }),
			min_by(words, fun(word) { return word }),
			max_by(words, fun(word) { return word }),
			take_while([1, 2, 3, 1], fun(value) { return value < 3 }),
			drop_while([1, 2, 3, 1], fun(value) { return value < 3 }),
			find_index(words, fun(word) { return word == "apple" }),
			key_by([{"id": "a"}, {"id": "b"}], fun(row) { return row.id })
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[[1, 10, 2, 20], [pear, plum], 2, [pear, plum, apple, apricot], apple, plum, [1, 2], [3, 1], 2, {"a": {"id": a}, "b": {"id": b}}]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestCollectionTransformBuiltinsValidateCallbacksAndKeys(t *testing.T) {
	for _, source := range []string{
		`flat_map([1], fun(value) { return value })`,
		`group_by([1], fun(value) { return value })`,
		`sort_by([1], fun(value) { return nil })`,
		`sort_by([1, 2], fun(value) { if value == 1 { return value } else { return "two" } })`,
		`key_by([1, 2], fun(value) { return "same" })`,
		`min_by([1], fun(value) { return 0 / 0 })`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestUnicodeStringBuiltins(t *testing.T) {
	value, err := executeForTest(`
		[
			substring("こんにちは世界", 1, 4),
			rune_at("A🌲", -1),
			ord("é"),
			chr(128640),
			pad_left("Sitka", 8, "🙂"),
			pad_right("x", 4, "0"),
			repeat("ab", 3),
			split_lines("one\r\ntwo\n"),
			join_lines(["first", "second"])
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[んにち, 🌲, 233, 🚀, 🙂🙂🙂Sitka, x000, ababab, [one, two], first
second]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestUnicodeStringBuiltinsRejectInvalidInputs(t *testing.T) {
	for _, source := range []string{
		`rune_at("", 0)`,
		`ord("two")`,
		`chr(-1)`,
		`chr(55296)`,
		`repeat("x", -1)`,
		`pad_left("x", 4, "xy")`,
		`join_lines(["ok", 1])`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestMathBuiltins(t *testing.T) {
	value, err := executeForTest(`
		[
			exp(0),
			ln(e),
			log2(8),
			log10(100),
			tan(0),
			asin(0),
			acos(1),
			atan(0),
			atan2(0, 1),
			hypot(3, 4),
			trunc(-2.8),
			divmod(17, 5),
			gcd(54, 24),
			lcm(6, 8),
			factorial(20)
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[1, 1, 3, 2, 0, 0, 0, 0, 0, 5, -2, [3, 2], 6, 24, 2432902008176640000]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestMathBuiltinsRejectInvalidInputsAndOverflow(t *testing.T) {
	for _, source := range []string{
		`ln(0)`,
		`log2(-1)`,
		`asin(2)`,
		`divmod(1, 0)`,
		`divmod(-9223372036854775808, -1)`,
		`gcd(-9223372036854775808, 0)`,
		`lcm(9223372036854775807, 2)`,
		`factorial(21)`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestTextEncodingAndQueryBuiltins(t *testing.T) {
	value, err := executeForTest(`
		[
			base64_decode(base64_encode("Sitka 🌲")),
			hex_decode(hex_encode("café")),
			url_encode("a b+c"),
			url_decode("a+b%2Bc"),
			query_decode("tag=go&tag=sitka&q=a+b"),
			query_encode({"q": "a b", "tag": ["go", "sitka"]})
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[Sitka 🌲, café, a+b%2Bc, a b+c, {"q": a b, "tag": [go, sitka]}, q=a+b&tag=go&tag=sitka]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestTextEncodingRejectsInvalidInput(t *testing.T) {
	for _, source := range []string{
		`base64_decode("not base64!")`,
		`base64_decode("/w==")`,
		`hex_decode("xyz")`,
		`hex_decode("ff")`,
		`url_decode("%")`,
		`query_decode("bad=%")`,
		`query_encode({"invalid": 1})`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestURLParseAndBuild(t *testing.T) {
	value, err := executeForTest(`
		let parsed = url_parse("https://ada:secret@example.com:8443/books/sitka?q=go#intro")
		[
			parsed.scheme,
			parsed.hostname,
			parsed.port,
			parsed.user,
			parsed.password,
			url_build({
				"scheme": parsed.scheme,
				"host": parsed.host,
				"path": parsed.path,
				"raw_query": parsed.raw_query,
				"fragment": parsed.fragment,
				"user": parsed.user,
				"password": parsed.password
			})
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[https, example.com, 8443, ada, secret, https://ada:secret@example.com:8443/books/sitka?q=go#intro]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestSHA256Builtins(t *testing.T) {
	value, err := executeForTest(`
		[sha256("abc"), sha256_bytes([97, 98, 99])]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad, ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestURLAndHashBuiltinsRejectInvalidInputs(t *testing.T) {
	for _, source := range []string{
		`url_parse(1)`,
		`url_parse("http://[")`,
		`url_build({"unknown": "value"})`,
		`url_build({"password": "secret"})`,
		`url_build({"scheme": 1})`,
		`sha256_bytes([256])`,
		`sha256(1)`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestTryCatchHandlesThrownAndRuntimeErrors(t *testing.T) {
	value, err := executeForTest(`
		let thrown = nil
		let runtime = nil
		try {
			throw "custom failure"
		} catch problem {
			thrown = problem
		}
		try {
			10 / 0
		} catch problem {
			runtime = problem
		}
		[thrown, contains(runtime, "division by zero")]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[thrown on line 5: custom failure, true]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestTryCatchPreservesControlFlowAndScopesCatchVariable(t *testing.T) {
	value, err := executeForTest(`
		fun findValue() {
			try {
				return 42
			} catch problem {
				return 0
			}
		}
		let loopCount = 0
		for value in [1, 2, 3] {
			try {
				if value == 2 { break }
			} catch problem {
				loopCount = loopCount + 100
			}
			loopCount = loopCount + 1
		}
		[findValue(), loopCount]
	`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), "[42, 1]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	if _, err := executeForTest(`
		try { throw "local" } catch problem { nil }
		problem
	`); err == nil {
		t.Fatal("catch variable escaped its scope")
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

func TestNumericAndCollectionUtilities(t *testing.T) {
	value, err := executeForTest(`
		fun isEven(value) { return value % 2 == 0 }
		let numbers = [1, 2, 3, 4, 5, 6]
		[
			sqrt(81),
			pow(2, 8),
			floor(3.9),
			ceil(3.1),
			round(3.5),
			clamp(10, 0, 5),
			sign(-7),
			degrees(pi),
			chunk(numbers, 4),
			take(numbers, 2),
			drop(numbers, 4),
			partition(numbers, isEven),
			find(numbers, fun(value) { return value > 4 }),
			count(numbers, 2),
			index_of(numbers, 5)
		]
	`)
	if err != nil {
		t.Fatal(err)
	}
	want := `[9, 256, 3, 4, 4, 5, -1, 180, [[1, 2, 3, 4], [5, 6]], [1, 2], [5, 6], [[2, 4, 6], [1, 3, 5]], 5, 1, 4]`
	if got := Format(value); got != want {
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
		`sqrt(-1)`,
		`clamp(1, 5, 0)`,
		`chunk([1, 2], 0)`,
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

func TestImportsResolveRelativeToImporter(t *testing.T) {
	root := t.TempDir()
	modules := filepath.Join(root, "modules")
	if err := os.Mkdir(modules, 0o700); err != nil {
		t.Fatal(err)
	}
	writeSitkaFile(t, filepath.Join(modules, "math.sitka"), `
		let base = 40
		let _secret = "private"
		fun addBase(value) { return value + base }
	`)
	mainPath := filepath.Join(root, "main.sitka")
	writeSitkaFile(t, mainPath, `
		import "modules/math.sitka" as math
		math.addBase(2)
	`)

	value, err := New().ExecuteFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if value != int64(42) {
		t.Fatalf("got %v, want 42", value)
	}
}

func TestImportedPrivateNamesAreNotExported(t *testing.T) {
	root := t.TempDir()
	writeSitkaFile(t, filepath.Join(root, "library.sitka"), `
		let visible = 1
		let _hidden = 2
	`)
	mainPath := filepath.Join(root, "main.sitka")
	writeSitkaFile(t, mainPath, `
		import "library.sitka" as library
		library._hidden
	`)
	if _, err := New().ExecuteFile(mainPath); err == nil {
		t.Fatal("private module name was accessible")
	}
}

func TestImportedModuleIsCached(t *testing.T) {
	root := t.TempDir()
	writeSitkaFile(t, filepath.Join(root, "counter.sitka"), `
		let count = 0
		fun increment() {
			count = count + 1
			return count
		}
	`)
	mainPath := filepath.Join(root, "main.sitka")
	writeSitkaFile(t, mainPath, `
		import "counter.sitka" as first
		import "counter.sitka" as second
		[first.increment(), second.increment()]
	`)
	value, err := New().ExecuteFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), "[1, 2]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestCircularImportsReportError(t *testing.T) {
	root := t.TempDir()
	writeSitkaFile(t, filepath.Join(root, "first.sitka"), `import "second.sitka" as second`)
	writeSitkaFile(t, filepath.Join(root, "second.sitka"), `import "first.sitka" as first`)
	_, err := New().ExecuteFile(filepath.Join(root, "first.sitka"))
	if err == nil {
		t.Fatal("circular module import unexpectedly succeeded")
	}
}

func TestImportedFunctionsRetainTheirModulePath(t *testing.T) {
	root := t.TempDir()
	modules := filepath.Join(root, "modules")
	if err := os.Mkdir(modules, 0o700); err != nil {
		t.Fatal(err)
	}
	writeSitkaFile(t, filepath.Join(modules, "helper.sitka"), `let answer = 42`)
	writeSitkaFile(t, filepath.Join(modules, "api.sitka"), `
		fun value() {
			import "helper.sitka" as helper
			return helper.answer
		}
	`)
	mainPath := filepath.Join(root, "main.sitka")
	writeSitkaFile(t, mainPath, `
		import "modules/api.sitka" as api
		api.value()
	`)
	value, err := New().ExecuteFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if value != int64(42) {
		t.Fatalf("got %v, want 42", value)
	}
}

func TestJSONRoundTripPreservesSitkaValues(t *testing.T) {
	value, err := executeForTest(`
		let original = {
			"name": "Sitka",
			"count": 3,
			"ratio": 2.5,
			"enabled": true,
			"empty": nil,
			"items": [1, "two", false]
		}
		let encoded = json_encode(original)
		let decoded = json_decode(encoded)
		[decoded == original, decoded.name, decoded.items[1], type(decoded.count)]
	`)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), "[true, Sitka, two, int]"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestJSONReportsMalformedAndUnsupportedValues(t *testing.T) {
	for _, source := range []string{
		`json_decode("{")`,
		`json_decode("1 2")`,
		`json_encode(fun() { return nil })`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
}

func TestFileBuiltinsUseScriptRelativePaths(t *testing.T) {
	root := t.TempDir()
	mainPath := filepath.Join(root, "main.sitka")
	writeSitkaFile(t, mainPath, `
		let bytes = write_text("notes.txt", "Sitka")
		append_text("notes.txt", " is joyful")
		let exists = file_exists("notes.txt")
		let text = read_text("notes.txt")
		[bytes, exists, text, list_dir(".")]
	`)
	value, err := New().ExecuteFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := Format(value), `[5, true, Sitka is joyful, [main.sitka, notes.txt]]`; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	content, err := os.ReadFile(filepath.Join(root, "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "Sitka is joyful" {
		t.Fatalf("file contents = %q", content)
	}
}

func TestFileBuiltinsSurfaceFilesystemErrors(t *testing.T) {
	_, err := executeForTest(`read_text("does-not-exist.sitka")`)
	if err == nil {
		t.Fatal("read_text silently accepted a missing file")
	}
}

func TestLineAndByteFileBuiltins(t *testing.T) {
	root := t.TempDir()
	mainPath := filepath.Join(root, "main.sitka")
	writeSitkaFile(t, mainPath, `
		write_lines("lines.txt", ["first", "second", ""])
		let lines = read_lines("lines.txt")
		let size = write_bytes("data.bin", [0, 10, 255])
		let bytes = read_bytes("data.bin")
		let appended = append_bytes("data.bin", [65, 66])
		let fullBytes = read_bytes("data.bin")
		[lines, size, bytes, appended, fullBytes]
	`)
	value, err := New().ExecuteFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	want := `[[first, second, ], 3, [0, 10, 255], 2, [0, 10, 255, 65, 66]]`
	if got := Format(value); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	content, err := os.ReadFile(filepath.Join(root, "lines.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(content), "first\nsecond\n\n"; got != want {
		t.Fatalf("written lines = %q, want %q", got, want)
	}
}

func TestByteFileBuiltinsValidateByteValuesAndBinaryText(t *testing.T) {
	for _, source := range []string{
		`write_bytes("data.bin", [256])`,
		`write_bytes("data.bin", [-1])`,
		`write_bytes("data.bin", [1.5])`,
		`append_bytes("data.bin", "not a list")`,
	} {
		if _, err := executeForTest(source); err == nil {
			t.Errorf("%s unexpectedly succeeded", source)
		}
	}
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "invalid.txt"), []byte{0xff}, 0o600); err != nil {
		t.Fatal(err)
	}
	mainPath := filepath.Join(root, "main.sitka")
	writeSitkaFile(t, mainPath, `read_lines("invalid.txt")`)
	if _, err := New().ExecuteFile(mainPath); err == nil {
		t.Fatal("read_lines accepted invalid UTF-8 data")
	}
}

func writeSitkaFile(t *testing.T, path, source string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatal(err)
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
