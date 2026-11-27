package interpreter

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode/utf8"
)

func (runtime *Runtime) installExtendedBuiltins() {
	runtime.Globals.Define("map", builtin{"map", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("map", arguments)
		if err != nil {
			return nil, err
		}
		result := make([]Value, len(values))
		for index, value := range values {
			mapped, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("map callback at index %d: %w", index, err)
			}
			result[index] = mapped
		}
		return result, nil
	}})
	runtime.Globals.Define("filter", builtin{"filter", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("filter", arguments)
		if err != nil {
			return nil, err
		}
		result := make([]Value, 0, len(values))
		for index, value := range values {
			keep, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("filter callback at index %d: %w", index, err)
			}
			if Truthy(keep) {
				result = append(result, value)
			}
		}
		return result, nil
	}})
	runtime.Globals.Define("reduce", builtin{"reduce", func(arguments []Value) (Value, error) {
		if len(arguments) != 3 {
			return nil, fmt.Errorf("reduce expects 3 arguments: list, callback, initial value")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("reduce first argument must be a list")
		}
		if !isCallable(arguments[1]) {
			return nil, fmt.Errorf("reduce second argument must be a function")
		}
		result := arguments[2]
		for index, value := range values {
			next, err := runtime.call(arguments[1], []Value{result, value})
			if err != nil {
				return nil, fmt.Errorf("reduce callback at index %d: %w", index, err)
			}
			result = next
		}
		return result, nil
	}})
	runtime.Globals.Define("sorted", builtin{"sorted", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("sorted expects 1 argument")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("sorted argument must be a list")
		}
		result := append([]Value(nil), values...)
		var orderType string
		for _, value := range result {
			currentType := orderableType(value)
			if currentType == "" {
				return nil, fmt.Errorf("cannot sort %s values", TypeName(value))
			}
			if orderType == "" {
				orderType = currentType
			} else if currentType != orderType {
				return nil, fmt.Errorf("cannot sort mixed %s and %s values", orderType, currentType)
			}
		}
		sort.SliceStable(result, func(left, right int) bool {
			comparison, comparable := compare(result[left], result[right])
			return comparable && comparison < 0
		})
		return result, nil
	}})
	runtime.Globals.Define("reversed", builtin{"reversed", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("reversed expects 1 argument")
		}
		switch value := arguments[0].(type) {
		case []Value:
			result := make([]Value, len(value))
			for index := range value {
				result[len(value)-1-index] = value[index]
			}
			return result, nil
		case string:
			runes := []rune(value)
			for left, right := 0, len(runes)-1; left < right; left, right = left+1, right-1 {
				runes[left], runes[right] = runes[right], runes[left]
			}
			return string(runes), nil
		default:
			return nil, fmt.Errorf("reversed expects a list or string")
		}
	}})
	runtime.Globals.Define("enumerate", builtin{"enumerate", func(arguments []Value) (Value, error) {
		if len(arguments) < 1 || len(arguments) > 2 {
			return nil, fmt.Errorf("enumerate expects a list and optional starting index")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("enumerate first argument must be a list")
		}
		start := int64(0)
		if len(arguments) == 2 {
			start, ok = arguments[1].(int64)
			if !ok {
				return nil, fmt.Errorf("enumerate starting index must be an integer")
			}
		}
		result := make([]Value, len(values))
		for index, value := range values {
			result[index] = []Value{start + int64(index), value}
		}
		return result, nil
	}})
	runtime.Globals.Define("zip", builtin{"zip", func(arguments []Value) (Value, error) {
		if len(arguments) < 2 {
			return nil, fmt.Errorf("zip expects at least 2 lists")
		}
		lists := make([][]Value, len(arguments))
		length := math.MaxInt
		for index, argument := range arguments {
			list, ok := argument.([]Value)
			if !ok {
				return nil, fmt.Errorf("zip arguments must be lists")
			}
			lists[index] = list
			if len(list) < length {
				length = len(list)
			}
		}
		result := make([]Value, length)
		for row := 0; row < length; row++ {
			values := make([]Value, len(lists))
			for column, list := range lists {
				values[column] = list[row]
			}
			result[row] = values
		}
		return result, nil
	}})
	runtime.Globals.Define("keys", builtin{"keys", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("keys expects 1 argument")
		}
		mapping, ok := arguments[0].(map[string]Value)
		if !ok {
			return nil, fmt.Errorf("keys argument must be a map")
		}
		names := make([]string, 0, len(mapping))
		for name := range mapping {
			names = append(names, name)
		}
		sort.Strings(names)
		result := make([]Value, len(names))
		for index, name := range names {
			result[index] = name
		}
		return result, nil
	}})
	runtime.Globals.Define("values", builtin{"values", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("values expects 1 argument")
		}
		mapping, ok := arguments[0].(map[string]Value)
		if !ok {
			return nil, fmt.Errorf("values argument must be a map")
		}
		names := make([]string, 0, len(mapping))
		for name := range mapping {
			names = append(names, name)
		}
		sort.Strings(names)
		result := make([]Value, len(names))
		for index, name := range names {
			result[index] = mapping[name]
		}
		return result, nil
	}})
	runtime.Globals.Define("join", builtin{"join", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("join expects a list and separator")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("join first argument must be a list")
		}
		separator, ok := arguments[1].(string)
		if !ok {
			return nil, fmt.Errorf("join separator must be a string")
		}
		parts := make([]string, len(values))
		for index, value := range values {
			parts[index] = Format(value)
		}
		return strings.Join(parts, separator), nil
	}})
	runtime.Globals.Define("split", builtin{"split", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("split expects a string and separator")
		}
		value, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("split first argument must be a string")
		}
		separator, ok := arguments[1].(string)
		if !ok {
			return nil, fmt.Errorf("split separator must be a string")
		}
		parts := strings.Split(value, separator)
		result := make([]Value, len(parts))
		for index, part := range parts {
			result[index] = part
		}
		return result, nil
	}})
	runtime.Globals.Define("upper", stringTransform("upper", strings.ToUpper))
	runtime.Globals.Define("lower", stringTransform("lower", strings.ToLower))
	runtime.Globals.Define("trim", stringTransform("trim", strings.TrimSpace))
	runtime.Globals.Define("replace", builtin{"replace", func(arguments []Value) (Value, error) {
		if len(arguments) != 3 {
			return nil, fmt.Errorf("replace expects a string, old value, and new value")
		}
		value, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("replace first argument must be a string")
		}
		oldValue, oldOK := arguments[1].(string)
		newValue, newOK := arguments[2].(string)
		if !oldOK || !newOK {
			return nil, fmt.Errorf("replace search and replacement values must be strings")
		}
		return strings.ReplaceAll(value, oldValue, newValue), nil
	}})
	runtime.Globals.Define("starts_with", stringPredicate("starts_with", strings.HasPrefix))
	runtime.Globals.Define("ends_with", stringPredicate("ends_with", strings.HasSuffix))
	runtime.Globals.Define("sum", builtin{"sum", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("sum expects 1 list")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("sum argument must be a list")
		}
		var integerSum int64
		var floatSum float64
		hasFloat := false
		for _, value := range values {
			switch number := value.(type) {
			case int64:
				if (number > 0 && integerSum > math.MaxInt64-number) ||
					(number < 0 && integerSum < math.MinInt64-number) {
					return nil, fmt.Errorf("integer overflow in sum")
				}
				integerSum += number
			case float64:
				hasFloat = true
				floatSum += number
			default:
				return nil, fmt.Errorf("sum cannot add %s", TypeName(value))
			}
		}
		if hasFloat {
			return floatSum + float64(integerSum), nil
		}
		return integerSum, nil
	}})
	runtime.Globals.Define("all", collectionPredicate("all", func(value Value) bool { return Truthy(value) }))
	runtime.Globals.Define("any", collectionPredicate("any", func(value Value) bool { return Truthy(value) }))
	runtime.Globals.Define("assert", builtin{"assert", func(arguments []Value) (Value, error) {
		if len(arguments) < 1 || len(arguments) > 2 {
			return nil, fmt.Errorf("assert expects a condition and optional message")
		}
		if Truthy(arguments[0]) {
			return arguments[0], nil
		}
		message := "assertion failed"
		if len(arguments) == 2 {
			message = Format(arguments[1])
		}
		return nil, fmt.Errorf("%s", message)
	}})
	runtime.Globals.Define("first", builtin{"first", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("first expects 1 list")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("first argument must be a list")
		}
		if len(values) == 0 {
			return nil, nil
		}
		return values[0], nil
	}})
	runtime.Globals.Define("last", builtin{"last", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("last expects 1 list")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("last argument must be a list")
		}
		if len(values) == 0 {
			return nil, nil
		}
		return values[len(values)-1], nil
	}})
	runtime.Globals.Define("slice", builtin{"slice", func(arguments []Value) (Value, error) {
		if len(arguments) < 2 || len(arguments) > 3 {
			return nil, fmt.Errorf("slice expects a list, start, and optional end")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("slice first argument must be a list")
		}
		start, ok := arguments[1].(int64)
		if !ok {
			return nil, fmt.Errorf("slice start must be an integer")
		}
		end := int64(len(values))
		if len(arguments) == 3 {
			end, ok = arguments[2].(int64)
			if !ok {
				return nil, fmt.Errorf("slice end must be an integer")
			}
		}
		startIndex := clampSliceIndex(start, len(values))
		endIndex := clampSliceIndex(end, len(values))
		if endIndex < startIndex {
			endIndex = startIndex
		}
		return append([]Value(nil), values[startIndex:endIndex]...), nil
	}})
	runtime.Globals.Define("unique", builtin{"unique", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("unique expects 1 list")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("unique argument must be a list")
		}
		result := make([]Value, 0, len(values))
		for _, value := range values {
			if !sliceContains(result, value) {
				result = append(result, value)
			}
		}
		return result, nil
	}})
	runtime.Globals.Define("flatten", builtin{"flatten", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("flatten expects 1 list")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("flatten argument must be a list")
		}
		result := make([]Value, 0)
		for _, value := range values {
			if nested, ok := value.([]Value); ok {
				result = append(result, nested...)
			} else {
				result = append(result, value)
			}
		}
		return result, nil
	}})
	runtime.Globals.Define("chars", builtin{"chars", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("chars expects 1 string")
		}
		value, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("chars argument must be a string")
		}
		result := make([]Value, 0, utf8.RuneCountInString(value))
		for _, character := range value {
			result = append(result, string(character))
		}
		return result, nil
	}})
}

func listAndCallback(name string, arguments []Value) ([]Value, Value, error) {
	if len(arguments) != 2 {
		return nil, nil, fmt.Errorf("%s expects a list and callback", name)
	}
	values, ok := arguments[0].([]Value)
	if !ok {
		return nil, nil, fmt.Errorf("%s first argument must be a list", name)
	}
	if !isCallable(arguments[1]) {
		return nil, nil, fmt.Errorf("%s second argument must be a function", name)
	}
	return values, arguments[1], nil
}

func isCallable(value Value) bool {
	switch value.(type) {
	case builtin, *function:
		return true
	default:
		return false
	}
}

func stringTransform(name string, transform func(string) string) builtin {
	return builtin{name, func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("%s expects 1 string", name)
		}
		value, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("%s argument must be a string", name)
		}
		return transform(value), nil
	}}
}

func stringPredicate(name string, predicate func(string, string) bool) builtin {
	return builtin{name, func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("%s expects 2 strings", name)
		}
		value, valueOK := arguments[0].(string)
		prefix, prefixOK := arguments[1].(string)
		if !valueOK || !prefixOK {
			return nil, fmt.Errorf("%s arguments must be strings", name)
		}
		return predicate(value, prefix), nil
	}}
}

func collectionPredicate(name string, predicate func(Value) bool) builtin {
	return builtin{name, func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("%s expects 1 list", name)
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("%s argument must be a list", name)
		}
		for _, value := range values {
			if predicate(value) {
				if name == "any" {
					return true, nil
				}
			} else if name == "all" {
				return false, nil
			}
		}
		return name == "all", nil
	}}
}

func clampSliceIndex(index int64, length int) int {
	if index < 0 {
		index += int64(length)
	}
	if index < 0 {
		return 0
	}
	if index > int64(length) {
		return length
	}
	return int(index)
}

func orderableType(value Value) string {
	switch value.(type) {
	case int64, float64:
		return "number"
	case string:
		return "string"
	default:
		return ""
	}
}
