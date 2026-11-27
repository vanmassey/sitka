package interpreter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strconv"
)

func (runtime *Runtime) installJSONBuiltins() {
	runtime.Globals.Define("json_encode", builtin{"json_encode", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("json_encode expects 1 argument")
		}
		converted, err := toJSON(arguments[0], 0)
		if err != nil {
			return nil, fmt.Errorf("cannot encode %s as JSON: %w", TypeName(arguments[0]), err)
		}
		encoded, err := json.Marshal(converted)
		if err != nil {
			return nil, fmt.Errorf("cannot encode %s as JSON: %w", TypeName(arguments[0]), err)
		}
		return string(encoded), nil
	}})
	runtime.Globals.Define("json_decode", builtin{"json_decode", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("json_decode expects 1 string")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("json_decode argument must be a string")
		}
		decoder := json.NewDecoder(bytes.NewBufferString(text))
		decoder.UseNumber()
		var decoded any
		if err := decoder.Decode(&decoded); err != nil {
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		var trailing any
		if err := decoder.Decode(&trailing); err != io.EOF {
			if err == nil {
				return nil, fmt.Errorf("invalid JSON: multiple values")
			}
			return nil, fmt.Errorf("invalid JSON: %w", err)
		}
		return fromJSON(decoded)
	}})
}

func toJSON(value Value, depth int) (any, error) {
	if depth > 256 {
		return nil, fmt.Errorf("JSON value exceeds maximum nesting depth")
	}
	switch value := value.(type) {
	case nil, bool, string, int64:
		return value, nil
	case float64:
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("JSON cannot represent non-finite numbers")
		}
		return value, nil
	case []Value:
		result := make([]any, len(value))
		for index, item := range value {
			converted, err := toJSON(item, depth+1)
			if err != nil {
				return nil, err
			}
			result[index] = converted
		}
		return result, nil
	case map[string]Value:
		result := make(map[string]any, len(value))
		for key, item := range value {
			converted, err := toJSON(item, depth+1)
			if err != nil {
				return nil, err
			}
			result[key] = converted
		}
		return result, nil
	case *structInstance:
		fields := make(map[string]Value, len(value.fields))
		for name, field := range value.fields {
			fields[name] = field
		}
		return toJSON(fields, depth+1)
	default:
		return nil, fmt.Errorf("unsupported value type %s", TypeName(value))
	}
}

func fromJSON(value any) (Value, error) {
	switch value := value.(type) {
	case nil, bool, string:
		return value, nil
	case json.Number:
		if integer, err := strconv.ParseInt(string(value), 10, 64); err == nil {
			return integer, nil
		}
		number, err := strconv.ParseFloat(string(value), 64)
		if err != nil {
			return nil, fmt.Errorf("JSON number %q is out of range", value)
		}
		return number, nil
	case []any:
		values := make([]Value, len(value))
		for index, item := range value {
			converted, err := fromJSON(item)
			if err != nil {
				return nil, err
			}
			values[index] = converted
		}
		return values, nil
	case map[string]any:
		values := make(map[string]Value, len(value))
		for key, item := range value {
			converted, err := fromJSON(item)
			if err != nil {
				return nil, err
			}
			values[key] = converted
		}
		return values, nil
	default:
		return nil, fmt.Errorf("unsupported JSON value %T", value)
	}
}
