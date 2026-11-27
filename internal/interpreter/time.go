package interpreter

import (
	"fmt"
	"math"
	"time"
)

func (runtime *Runtime) installTimeBuiltins() {
	runtime.Globals.Define("time_unix", builtin{"time_unix", func(arguments []Value) (Value, error) {
		if len(arguments) < 3 || len(arguments) > 6 {
			return nil, fmt.Errorf("time_unix expects year, month, day, and optional hour, minute, and second")
		}
		parts := [6]int{0, 1, 1, 0, 0, 0}
		for index, argument := range arguments {
			value, ok := argument.(int64)
			if !ok {
				return nil, fmt.Errorf("time_unix argument %d must be an integer", index+1)
			}
			if value < math.MinInt32 || value > math.MaxInt32 {
				return nil, fmt.Errorf("time_unix argument %d is out of range", index+1)
			}
			parts[index] = int(value)
		}
		if parts[0] < 1 || parts[0] > 9999 {
			return nil, fmt.Errorf("time_unix year must be between 1 and 9999")
		}
		if parts[1] < 1 || parts[1] > 12 || parts[2] < 1 || parts[2] > 31 ||
			parts[3] < 0 || parts[3] > 23 || parts[4] < 0 || parts[4] > 59 ||
			parts[5] < 0 || parts[5] > 59 {
			return nil, fmt.Errorf("time_unix received an invalid date or time")
		}
		value := time.Date(parts[0], time.Month(parts[1]), parts[2], parts[3], parts[4], parts[5], 0, time.UTC)
		if value.Year() != parts[0] || int(value.Month()) != parts[1] ||
			value.Day() != parts[2] || value.Hour() != parts[3] ||
			value.Minute() != parts[4] || value.Second() != parts[5] {
			return nil, fmt.Errorf("time_unix received an invalid date or time")
		}
		return float64(value.Unix()), nil
	}})
	runtime.Globals.Define("time_parse", builtin{"time_parse", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("time_parse expects text and layout")
		}
		text, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("time_parse text must be a string")
		}
		layout, ok := arguments[1].(string)
		if !ok {
			return nil, fmt.Errorf("time_parse layout must be a string")
		}
		parsed, err := time.Parse(layout, text)
		if err != nil {
			return nil, fmt.Errorf("time_parse: %w", err)
		}
		return unixSeconds(parsed), nil
	}})
	runtime.Globals.Define("time_format", builtin{"time_format", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("time_format expects a timestamp and layout")
		}
		value, err := timeArgument("time_format", arguments[0])
		if err != nil {
			return nil, err
		}
		layout, ok := arguments[1].(string)
		if !ok {
			return nil, fmt.Errorf("time_format layout must be a string")
		}
		return value.Format(layout), nil
	}})
	runtime.Globals.Define("time_utc", builtin{"time_utc", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("time_utc expects 1 timestamp")
		}
		value, err := timeArgument("time_utc", arguments[0])
		if err != nil {
			return nil, err
		}
		return value.UTC().Format(time.RFC3339Nano), nil
	}})
	runtime.Globals.Define("time_add", builtin{"time_add", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("time_add expects a timestamp and seconds")
		}
		value, err := timeArgument("time_add", arguments[0])
		if err != nil {
			return nil, err
		}
		seconds, ok := numericSeconds(arguments[1])
		if !ok {
			return nil, fmt.Errorf("time_add seconds must be a finite number")
		}
		duration := seconds * float64(time.Second)
		if duration >= float64(math.MaxInt64) || duration < float64(math.MinInt64) {
			return nil, fmt.Errorf("time_add duration is out of range")
		}
		return unixSeconds(value.Add(time.Duration(duration))), nil
	}})
	runtime.Globals.Define("time_diff", builtin{"time_diff", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("time_diff expects start and end timestamps")
		}
		start, err := timeArgument("time_diff", arguments[0])
		if err != nil {
			return nil, err
		}
		end, err := timeArgument("time_diff", arguments[1])
		if err != nil {
			return nil, err
		}
		return unixSeconds(end) - unixSeconds(start), nil
	}})
	runtime.Globals.Define("time_parts", builtin{"time_parts", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("time_parts expects 1 timestamp")
		}
		value, err := timeArgument("time_parts", arguments[0])
		if err != nil {
			return nil, err
		}
		value = value.UTC()
		return map[string]Value{
			"year":       int64(value.Year()),
			"month":      int64(value.Month()),
			"day":        int64(value.Day()),
			"hour":       int64(value.Hour()),
			"minute":     int64(value.Minute()),
			"second":     int64(value.Second()),
			"nanosecond": int64(value.Nanosecond()),
			"weekday":    int64(value.Weekday()),
			"year_day":   int64(value.YearDay()),
		}, nil
	}})
}

func timeArgument(name string, value Value) (time.Time, error) {
	seconds, ok := numericSeconds(value)
	if !ok {
		return time.Time{}, fmt.Errorf("%s timestamp must be a finite number", name)
	}
	whole, fraction := math.Modf(seconds)
	if whole < math.MinInt64 || whole >= math.MaxInt64 {
		return time.Time{}, fmt.Errorf("%s timestamp is out of range", name)
	}
	nanoseconds := int64(math.Round(fraction * float64(time.Second)))
	return time.Unix(int64(whole), nanoseconds), nil
}

func numericSeconds(value Value) (float64, bool) {
	var seconds float64
	switch value := value.(type) {
	case int64:
		seconds = float64(value)
	case float64:
		seconds = value
	default:
		return 0, false
	}
	return seconds, !math.IsNaN(seconds) && !math.IsInf(seconds, 0)
}

func unixSeconds(value time.Time) float64 {
	return float64(value.Unix()) + float64(value.Nanosecond())/float64(time.Second)
}
