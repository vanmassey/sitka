package interpreter

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func (runtime *Runtime) installStringBuiltins() {
	runtime.Globals.Define("substring", builtin{"substring", func(arguments []Value) (Value, error) {
		if len(arguments) < 2 || len(arguments) > 3 {
			return nil, fmt.Errorf("substring expects text, start, and optional end")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("substring text must be a string")
		}
		start, ok := arguments[1].(int64)
		if !ok {
			return nil, fmt.Errorf("substring start must be an integer")
		}
		runes := []rune(text)
		end := int64(len(runes))
		if len(arguments) == 3 {
			end, ok = arguments[2].(int64)
			if !ok {
				return nil, fmt.Errorf("substring end must be an integer")
			}
		}
		startIndex := clampSliceIndex(start, len(runes))
		endIndex := clampSliceIndex(end, len(runes))
		if endIndex < startIndex {
			endIndex = startIndex
		}
		return string(runes[startIndex:endIndex]), nil
	}})
	runtime.Globals.Define("rune_at", builtin{"rune_at", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("rune_at expects text and index")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("rune_at text must be a string")
		}
		index, ok := arguments[1].(int64)
		if !ok {
			return nil, fmt.Errorf("rune_at index must be an integer")
		}
		runes := []rune(text)
		position, err := normalizeIndex(index, len(runes))
		if err != nil {
			return nil, fmt.Errorf("rune_at: %w", err)
		}
		return string(runes[position]), nil
	}})
	runtime.Globals.Define("ord", builtin{"ord", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("ord expects 1 character")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("ord argument must be a string")
		}
		if !utf8.ValidString(text) || utf8.RuneCountInString(text) != 1 {
			return nil, fmt.Errorf("ord expects a string containing exactly one Unicode character")
		}
		return int64([]rune(text)[0]), nil
	}})
	runtime.Globals.Define("chr", builtin{"chr", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("chr expects 1 integer")
		}
		value, ok := arguments[0].(int64)
		if !ok || value < 0 || value > utf8.MaxRune || !utf8.ValidRune(rune(value)) {
			return nil, fmt.Errorf("chr argument must be a valid Unicode code point")
		}
		return string(rune(value)), nil
	}})
	runtime.Globals.Define("repeat", builtin{"repeat", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("repeat expects text and count")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("repeat text must be a string")
		}
		count, ok := arguments[1].(int64)
		if !ok || count < 0 {
			return nil, fmt.Errorf("repeat count must be a non-negative integer")
		}
		if count > maxGeneratedItems {
			return nil, fmt.Errorf("repeat count exceeds maximum size of %d", maxGeneratedItems)
		}
		if len(text) > 0 && count > int64(maxGeneratedStringBytes/len(text)) {
			return nil, fmt.Errorf("repeat result exceeds maximum size of %d bytes", maxGeneratedStringBytes)
		}
		return strings.Repeat(text, int(count)), nil
	}})
	runtime.Globals.Define("pad_left", stringPadBuiltin("pad_left", true))
	runtime.Globals.Define("pad_right", stringPadBuiltin("pad_right", false))
	runtime.Globals.Define("split_lines", builtin{"split_lines", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("split_lines expects 1 string")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("split_lines argument must be a string")
		}
		return splitTextLines(text)
	}})
	runtime.Globals.Define("join_lines", builtin{"join_lines", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("join_lines expects 1 list")
		}
		values, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("join_lines argument must be a list of strings")
		}
		lines := make([]string, len(values))
		totalBytes := len(values) - 1
		if totalBytes < 0 {
			totalBytes = 0
		}
		for index, value := range values {
			line, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("join_lines item %d must be a string", index+1)
			}
			if len(line) > maxGeneratedStringBytes-totalBytes {
				return nil, fmt.Errorf("join_lines result exceeds maximum size of %d bytes", maxGeneratedStringBytes)
			}
			totalBytes += len(line)
			lines[index] = line
		}
		return strings.Join(lines, "\n"), nil
	}})
}

func stringPadBuiltin(name string, left bool) builtin {
	return builtin{name, func(arguments []Value) (Value, error) {
		if len(arguments) != 3 {
			return nil, fmt.Errorf("%s expects text, width, and padding character", name)
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("%s text must be a string", name)
		}
		width, ok := arguments[1].(int64)
		if !ok || width < 0 {
			return nil, fmt.Errorf("%s width must be a non-negative integer", name)
		}
		padding, ok := arguments[2].(string)
		if !ok || utf8.RuneCountInString(padding) != 1 {
			return nil, fmt.Errorf("%s padding must be one Unicode character", name)
		}
		runes := []rune(text)
		if width <= int64(len(runes)) {
			return text, nil
		}
		paddingCount := width - int64(len(runes))
		paddingRune := []rune(padding)[0]
		if paddingCount > int64(maxGeneratedStringBytes/utf8.RuneLen(paddingRune)) {
			return nil, fmt.Errorf("%s result exceeds maximum size of %d bytes", name, maxGeneratedStringBytes)
		}
		paddingText := strings.Repeat(padding, int(paddingCount))
		if left {
			return paddingText + text, nil
		}
		return text + paddingText, nil
	}}
}
