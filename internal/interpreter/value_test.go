package interpreter

import (
	"math"
	"testing"
)

func TestTypeNamesCoverRuntimeValues(t *testing.T) {
	runtime := New()
	callable, ok := runtime.Globals.Get("len")
	if !ok {
		t.Fatal("len builtin is missing")
	}
	structure := &structType{name: "Point", fields: []string{"x"}}
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "nil", value: nil, want: "nil"},
		{name: "boolean", value: true, want: "bool"},
		{name: "integer", value: int64(1), want: "int"},
		{name: "float", value: float64(1), want: "float"},
		{name: "string", value: "text", want: "string"},
		{name: "list", value: []Value{}, want: "list"},
		{name: "map", value: map[string]Value{}, want: "map"},
		{name: "struct type", value: structure, want: "type"},
		{name: "struct instance", value: &structInstance{kind: structure, fields: map[string]Value{"x": int64(1)}}, want: "struct Point"},
		{name: "builtin", value: callable, want: "function"},
		{name: "unknown", value: struct{}{}, want: "unknown"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := TypeName(test.value); got != test.want {
				t.Fatalf("TypeName(%T) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

func TestValueFormatting(t *testing.T) {
	runtime := New()
	callable, _ := runtime.Globals.Get("len")
	structure := &structType{name: "Pair"}
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "nil", value: nil, want: "nil"},
		{name: "true", value: true, want: "true"},
		{name: "false", value: false, want: "false"},
		{name: "integer", value: int64(-42), want: "-42"},
		{name: "float", value: 1.25, want: "1.25"},
		{name: "string", value: "text", want: "text"},
		{name: "list", value: []Value{int64(1), "two"}, want: "[1, two]"},
		{name: "map sorted", value: map[string]Value{"z": int64(2), "a": int64(1)}, want: `{"a": 1, "z": 2}`},
		{name: "struct", value: &structInstance{kind: structure, fields: map[string]Value{"z": int64(2), "a": int64(1)}}, want: "Pair{a: 1, z: 2}"},
		{name: "struct type", value: structure, want: "<struct Pair>"},
		{name: "builtin", value: callable, want: "<builtin len>"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Format(test.value); got != test.want {
				t.Fatalf("Format(%T) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}

func TestValueEqualityAcrossCompositeTypes(t *testing.T) {
	point := &structType{name: "Point", fields: []string{"x", "y"}}
	tests := []struct {
		name  string
		left  Value
		right Value
		want  bool
	}{
		{name: "nil values", left: nil, right: nil, want: true},
		{name: "nil and false", left: nil, right: false},
		{name: "integer and equal float", left: int64(3), right: float64(3), want: true},
		{name: "integer and different float", left: int64(3), right: float64(3.1)},
		{name: "nested lists", left: []Value{int64(1), []Value{"x"}}, right: []Value{float64(1), []Value{"x"}}, want: true},
		{name: "different list lengths", left: []Value{int64(1)}, right: []Value{int64(1), int64(2)}},
		{name: "maps independent of insertion order", left: map[string]Value{"a": int64(1), "b": []Value{int64(2)}}, right: map[string]Value{"b": []Value{int64(2)}, "a": float64(1)}, want: true},
		{name: "different map keys", left: map[string]Value{"a": int64(1)}, right: map[string]Value{"b": int64(1)}},
		{name: "same struct values", left: &structInstance{kind: point, fields: map[string]Value{"x": int64(1), "y": int64(2)}}, right: &structInstance{kind: point, fields: map[string]Value{"x": float64(1), "y": int64(2)}}, want: true},
		{name: "different struct names", left: &structInstance{kind: point, fields: map[string]Value{"x": int64(1), "y": int64(2)}}, right: &structInstance{kind: &structType{name: "Other"}, fields: map[string]Value{"x": int64(1), "y": int64(2)}}},
		{name: "NaN", left: math.NaN(), right: math.NaN()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Equal(test.left, test.right); got != test.want {
				t.Fatalf("Equal(%v, %v) = %t, want %t", test.left, test.right, got, test.want)
			}
		})
	}
}

func TestTruthyValues(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  bool
	}{
		{name: "nil"},
		{name: "false", value: false},
		{name: "true", value: true, want: true},
		{name: "zero integer", value: int64(0), want: true},
		{name: "zero float", value: float64(0), want: true},
		{name: "empty string", value: "", want: true},
		{name: "empty list", value: []Value{}, want: true},
		{name: "empty map", value: map[string]Value{}, want: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Truthy(test.value); got != test.want {
				t.Fatalf("Truthy(%#v) = %t, want %t", test.value, got, test.want)
			}
		})
	}
}

func TestNormalizeIndexBoundaries(t *testing.T) {
	tests := []struct {
		name   string
		index  int64
		length int
		want   int
		err    bool
	}{
		{name: "first", index: 0, length: 3, want: 0},
		{name: "last", index: 2, length: 3, want: 2},
		{name: "negative last", index: -1, length: 3, want: 2},
		{name: "negative first", index: -3, length: 3, want: 0},
		{name: "too large", index: 3, length: 3, err: true},
		{name: "too negative", index: -4, length: 3, err: true},
		{name: "empty", index: 0, length: 0, err: true},
		{name: "minimum integer", index: math.MinInt64, length: 3, err: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := normalizeIndex(test.index, test.length)
			if (err != nil) != test.err {
				t.Fatalf("normalizeIndex(%d, %d) error = %v, want error %t", test.index, test.length, err, test.err)
			}
			if err == nil && got != test.want {
				t.Fatalf("normalizeIndex(%d, %d) = %d, want %d", test.index, test.length, got, test.want)
			}
		})
	}
}

func TestEnvironmentAssignmentRespectsNearestScope(t *testing.T) {
	root := NewEnvironment(nil)
	root.Define("value", int64(1))
	child := NewEnvironment(root)
	child.Define("local", int64(2))
	if value, found := child.Get("value"); !found || value != int64(1) {
		t.Fatalf("child lookup = (%v, %t), want (1, true)", value, found)
	}
	if !child.Assign("value", int64(3)) {
		t.Fatal("assignment to parent binding failed")
	}
	if value, _ := root.Get("value"); value != int64(3) {
		t.Fatalf("parent binding = %v, want 3", value)
	}
	child.Define("value", int64(4))
	if value, _ := child.Get("value"); value != int64(4) {
		t.Fatalf("shadowed child binding = %v, want 4", value)
	}
	if value, _ := root.Get("value"); value != int64(3) {
		t.Fatalf("shadowing changed parent binding to %v", value)
	}
	if child.Assign("missing", int64(5)) {
		t.Fatal("assignment created an undeclared binding")
	}
}
