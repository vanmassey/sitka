package interpreter

import (
	"testing"
	"unicode/utf8"
)

func callBuiltinForTest(t *testing.T, runtime *Runtime, name string, arguments ...Value) (Value, error) {
	t.Helper()
	callable, found := runtime.Globals.Get(name)
	if !found {
		t.Fatalf("builtin %q is not installed", name)
	}
	return runtime.call(callable, arguments)
}

func FuzzCSVTextFieldRoundTrip(f *testing.F) {
	for _, field := range []string{
		"",
		"Sitka",
		"a,b",
		`a "quoted" field`,
		"first line\nsecond line",
		"🌲",
	} {
		f.Add(field)
	}
	f.Fuzz(func(t *testing.T, field string) {
		if !utf8.ValidString(field) {
			t.Skip()
		}
		runtime := New()
		encoded, err := callBuiltinForTest(t, runtime, "csv_encode", []Value{[]Value{field}})
		if err != nil {
			t.Fatal(err)
		}
		text, ok := encoded.(string)
		if !ok {
			t.Fatalf("csv_encode returned %T, want string", encoded)
		}
		decoded, err := callBuiltinForTest(t, runtime, "csv_decode", text)
		if err != nil {
			t.Fatal(err)
		}
		rows, ok := decoded.([]Value)
		if !ok || len(rows) != 1 {
			t.Fatalf("csv_decode(%q) returned %#v, want one row", text, decoded)
		}
		row, ok := rows[0].([]Value)
		if !ok || len(row) != 1 || row[0] != field {
			t.Fatalf("CSV round trip returned %#v, want %q", decoded, field)
		}
	})
}

func FuzzURLQueryRoundTrip(f *testing.F) {
	f.Add("query", "language runtime")
	f.Add("tag", "go&sitka=joyful")
	f.Add("empty", "")
	f.Add("🌲", "snow%")
	f.Fuzz(func(t *testing.T, key, value string) {
		if !utf8.ValidString(key) || !utf8.ValidString(value) {
			t.Skip()
		}
		runtime := New()
		encoded, err := callBuiltinForTest(t, runtime, "query_encode", map[string]Value{key: value})
		if err != nil {
			t.Fatal(err)
		}
		text, ok := encoded.(string)
		if !ok {
			t.Fatalf("query_encode returned %T, want string", encoded)
		}
		decoded, err := callBuiltinForTest(t, runtime, "query_decode", text)
		if err != nil {
			t.Fatal(err)
		}
		if !Equal(decoded, map[string]Value{key: value}) {
			t.Fatalf("query round trip returned %#v, want key/value %q=%q", decoded, key, value)
		}
	})
}

func FuzzSubstringReturnsValidUTF8(f *testing.F) {
	f.Add("Sitka 🌲", int64(0), int64(5))
	f.Add("こんにちは", int64(1), int64(4))
	f.Add("", int64(-2), int64(8))
	f.Fuzz(func(t *testing.T, text string, start, end int64) {
		if !utf8.ValidString(text) {
			t.Skip()
		}
		runes := []rune(text)
		left := clampSliceIndex(start, len(runes))
		right := clampSliceIndex(end, len(runes))
		if right < left {
			right = left
		}
		value, err := callBuiltinForTest(t, New(), "substring", text, start, end)
		if err != nil {
			t.Fatal(err)
		}
		result, ok := value.(string)
		if !ok || result != string(runes[left:right]) {
			t.Fatalf("substring(%q, %d, %d) = %#v, want %q", text, start, end, value, string(runes[left:right]))
		}
		if !utf8.ValidString(result) {
			t.Fatalf("substring returned invalid UTF-8: %q", result)
		}
	})
}

func FuzzRegexBuiltinsNeverPanic(f *testing.F) {
	f.Add("order-42", `[a-z]+-[0-9]+`)
	f.Add("a.b", `\.`)
	f.Add("", "")
	f.Fuzz(func(t *testing.T, text, pattern string) {
		if len(text) > 4096 || len(pattern) > 256 {
			t.Skip()
		}
		runtime := New()
		for _, name := range []string{"regex_match", "regex_find", "regex_find_all", "regex_split", "regex_captures"} {
			_, _ = callBuiltinForTest(t, runtime, name, text, pattern)
		}
	})
}
