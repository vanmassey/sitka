package interpreter

import (
	"fmt"
	"sort"
)

func (runtime *Runtime) installCollectionBuiltins() {
	runtime.Globals.Define("flat_map", builtin{"flat_map", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("flat_map", arguments)
		if err != nil {
			return nil, err
		}
		result := make([]Value, 0)
		for index, value := range values {
			mapped, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("flat_map callback at index %d: %w", index, err)
			}
			items, ok := mapped.([]Value)
			if !ok {
				return nil, fmt.Errorf("flat_map callback at index %d returned %s, want list", index, TypeName(mapped))
			}
			if len(result)+len(items) > maxGeneratedItems {
				return nil, fmt.Errorf("flat_map result exceeds maximum size of %d items", maxGeneratedItems)
			}
			result = append(result, items...)
		}
		return result, nil
	}})
	runtime.Globals.Define("group_by", builtin{"group_by", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("group_by", arguments)
		if err != nil {
			return nil, err
		}
		groups := make(map[string]Value)
		for index, value := range values {
			key, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("group_by callback at index %d: %w", index, err)
			}
			name, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("group_by callback at index %d returned %s, want string", index, TypeName(key))
			}
			group, found := groups[name]
			if !found {
				group = []Value{}
			}
			groups[name] = append(group.([]Value), value)
		}
		return groups, nil
	}})
	runtime.Globals.Define("count_by", builtin{"count_by", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("count_by", arguments)
		if err != nil {
			return nil, err
		}
		counts := make(map[string]Value)
		for index, value := range values {
			key, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("count_by callback at index %d: %w", index, err)
			}
			name, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("count_by callback at index %d returned %s, want string", index, TypeName(key))
			}
			count, _ := counts[name].(int64)
			counts[name] = count + 1
		}
		return counts, nil
	}})
	runtime.Globals.Define("sort_by", builtin{"sort_by", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("sort_by", arguments)
		if err != nil {
			return nil, err
		}
		type entry struct {
			value Value
			key   Value
		}
		entries := make([]entry, len(values))
		var keyType string
		for index, value := range values {
			key, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("sort_by callback at index %d: %w", index, err)
			}
			currentType := orderableType(key)
			if currentType == "" {
				return nil, fmt.Errorf("sort_by callback at index %d returned non-orderable %s", index, TypeName(key))
			}
			if keyType != "" && currentType != keyType {
				return nil, fmt.Errorf("sort_by callback returned mixed %s and %s keys", keyType, currentType)
			}
			keyType = currentType
			entries[index] = entry{value: value, key: key}
		}
		sort.SliceStable(entries, func(left, right int) bool {
			comparison, comparable := compare(entries[left].key, entries[right].key)
			return comparable && comparison < 0
		})
		result := make([]Value, len(entries))
		for index, item := range entries {
			result[index] = item.value
		}
		return result, nil
	}})
	runtime.Globals.Define("take_while", builtin{"take_while", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("take_while", arguments)
		if err != nil {
			return nil, err
		}
		end := len(values)
		for index, value := range values {
			keep, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("take_while callback at index %d: %w", index, err)
			}
			if !Truthy(keep) {
				end = index
				break
			}
		}
		return append([]Value(nil), values[:end]...), nil
	}})
	runtime.Globals.Define("drop_while", builtin{"drop_while", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("drop_while", arguments)
		if err != nil {
			return nil, err
		}
		start := len(values)
		for index, value := range values {
			drop, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("drop_while callback at index %d: %w", index, err)
			}
			if !Truthy(drop) {
				start = index
				break
			}
		}
		return append([]Value(nil), values[start:]...), nil
	}})
	runtime.Globals.Define("find_index", builtin{"find_index", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("find_index", arguments)
		if err != nil {
			return nil, err
		}
		for index, value := range values {
			found, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("find_index callback at index %d: %w", index, err)
			}
			if Truthy(found) {
				return int64(index), nil
			}
		}
		return int64(-1), nil
	}})
	runtime.Globals.Define("key_by", builtin{"key_by", func(arguments []Value) (Value, error) {
		values, callback, err := listAndCallback("key_by", arguments)
		if err != nil {
			return nil, err
		}
		result := make(map[string]Value, len(values))
		for index, value := range values {
			key, err := runtime.call(callback, []Value{value})
			if err != nil {
				return nil, fmt.Errorf("key_by callback at index %d: %w", index, err)
			}
			name, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("key_by callback at index %d returned %s, want string", index, TypeName(key))
			}
			if _, found := result[name]; found {
				return nil, fmt.Errorf("key_by produced duplicate key %q", name)
			}
			result[name] = value
		}
		return result, nil
	}})
	runtime.Globals.Define("min_by", builtin{"min_by", func(arguments []Value) (Value, error) {
		return runtime.extremeBy("min_by", arguments, false)
	}})
	runtime.Globals.Define("max_by", builtin{"max_by", func(arguments []Value) (Value, error) {
		return runtime.extremeBy("max_by", arguments, true)
	}})
}

func (runtime *Runtime) extremeBy(name string, arguments []Value, maximum bool) (Value, error) {
	values, callback, err := listAndCallback(name, arguments)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	bestValue := values[0]
	bestKey, err := runtime.call(callback, []Value{bestValue})
	if err != nil {
		return nil, fmt.Errorf("%s callback at index 0: %w", name, err)
	}
	keyType := orderableType(bestKey)
	if keyType == "" {
		return nil, fmt.Errorf("%s callback at index 0 returned non-orderable %s", name, TypeName(bestKey))
	}
	for index, value := range values[1:] {
		key, err := runtime.call(callback, []Value{value})
		if err != nil {
			return nil, fmt.Errorf("%s callback at index %d: %w", name, index+1, err)
		}
		if currentType := orderableType(key); currentType == "" || currentType != keyType {
			return nil, fmt.Errorf("%s callback returned a non-orderable or mixed key at index %d", name, index+1)
		}
		comparison, _ := compare(key, bestKey)
		if (maximum && comparison > 0) || (!maximum && comparison < 0) {
			bestValue = value
			bestKey = key
		}
	}
	return bestValue, nil
}
