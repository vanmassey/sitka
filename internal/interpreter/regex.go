package interpreter

import (
	"fmt"
	"regexp"
)

func (runtime *Runtime) installRegexBuiltins() {
	runtime.Globals.Define("regex_match", builtin{"regex_match", func(arguments []Value) (Value, error) {
		text, expression, err := regexArguments("regex_match", arguments)
		if err != nil {
			return nil, err
		}
		return expression.MatchString(text), nil
	}})
	runtime.Globals.Define("regex_find", builtin{"regex_find", func(arguments []Value) (Value, error) {
		text, expression, err := regexArguments("regex_find", arguments)
		if err != nil {
			return nil, err
		}
		matchRange := expression.FindStringIndex(text)
		if matchRange == nil {
			return nil, nil
		}
		return text[matchRange[0]:matchRange[1]], nil
	}})
	runtime.Globals.Define("regex_find_all", builtin{"regex_find_all", func(arguments []Value) (Value, error) {
		text, expression, err := regexArguments("regex_find_all", arguments)
		if err != nil {
			return nil, err
		}
		matches := expression.FindAllString(text, maxGeneratedItems+1)
		if len(matches) > maxGeneratedItems {
			return nil, fmt.Errorf("regex_find_all exceeds maximum result size of %d", maxGeneratedItems)
		}
		result := make([]Value, len(matches))
		for index, match := range matches {
			result[index] = match
		}
		return result, nil
	}})
	runtime.Globals.Define("regex_replace", builtin{"regex_replace", func(arguments []Value) (Value, error) {
		if len(arguments) != 3 {
			return nil, fmt.Errorf("regex_replace expects text, pattern, and replacement")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("regex_replace text must be a string")
		}
		pattern, ok := arguments[1].(string)
		if !ok {
			return nil, fmt.Errorf("regex_replace pattern must be a string")
		}
		replacement, ok := arguments[2].(string)
		if !ok {
			return nil, fmt.Errorf("regex_replace replacement must be a string")
		}
		expression, err := compileRegex("regex_replace", pattern)
		if err != nil {
			return nil, err
		}
		result := expression.ReplaceAllString(text, replacement)
		if len(result) > maxGeneratedStringBytes {
			return nil, fmt.Errorf("regex_replace result exceeds maximum size of %d bytes", maxGeneratedStringBytes)
		}
		return result, nil
	}})
	runtime.Globals.Define("regex_split", builtin{"regex_split", func(arguments []Value) (Value, error) {
		text, expression, err := regexArguments("regex_split", arguments)
		if err != nil {
			return nil, err
		}
		parts := expression.Split(text, maxGeneratedItems+1)
		if len(parts) > maxGeneratedItems {
			return nil, fmt.Errorf("regex_split exceeds maximum result size of %d", maxGeneratedItems)
		}
		result := make([]Value, len(parts))
		for index, part := range parts {
			result[index] = part
		}
		return result, nil
	}})
	runtime.Globals.Define("regex_captures", builtin{"regex_captures", func(arguments []Value) (Value, error) {
		text, expression, err := regexArguments("regex_captures", arguments)
		if err != nil {
			return nil, err
		}
		match := expression.FindStringSubmatch(text)
		if match == nil {
			return nil, nil
		}
		captures := make(map[string]Value)
		names := expression.SubexpNames()
		for index := 1; index < len(match); index++ {
			if names[index] != "" {
				captures[names[index]] = match[index]
			} else {
				captures[fmt.Sprintf("%d", index)] = match[index]
			}
		}
		return captures, nil
	}})
}

func regexArguments(name string, arguments []Value) (string, *regexp.Regexp, error) {
	if len(arguments) != 2 {
		return "", nil, fmt.Errorf("%s expects text and pattern", name)
	}
	text, ok := arguments[0].(string)
	if !ok {
		return "", nil, fmt.Errorf("%s text must be a string", name)
	}
	pattern, ok := arguments[1].(string)
	if !ok {
		return "", nil, fmt.Errorf("%s pattern must be a string", name)
	}
	expression, err := compileRegex(name, pattern)
	if err != nil {
		return "", nil, err
	}
	return text, expression, nil
}

func compileRegex(name, pattern string) (*regexp.Regexp, error) {
	expression, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("%s pattern: %w", name, err)
	}
	return expression, nil
}
