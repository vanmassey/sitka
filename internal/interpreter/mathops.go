package interpreter

import (
	"fmt"
	"math"
	"math/bits"
)

func (runtime *Runtime) installMathBuiltins() {
	runtime.Globals.Define("exp", numericFunction("exp", 1, func(values []float64) (Value, error) {
		return math.Exp(values[0]), nil
	}))
	runtime.Globals.Define("ln", numericFunction("ln", 1, func(values []float64) (Value, error) {
		if values[0] <= 0 {
			return nil, fmt.Errorf("ln argument must be positive")
		}
		return math.Log(values[0]), nil
	}))
	runtime.Globals.Define("log2", numericFunction("log2", 1, func(values []float64) (Value, error) {
		if values[0] <= 0 {
			return nil, fmt.Errorf("log2 argument must be positive")
		}
		return math.Log2(values[0]), nil
	}))
	runtime.Globals.Define("log10", numericFunction("log10", 1, func(values []float64) (Value, error) {
		if values[0] <= 0 {
			return nil, fmt.Errorf("log10 argument must be positive")
		}
		return math.Log10(values[0]), nil
	}))
	runtime.Globals.Define("tan", numericFunction("tan", 1, func(values []float64) (Value, error) {
		return math.Tan(values[0]), nil
	}))
	runtime.Globals.Define("asin", numericFunction("asin", 1, func(values []float64) (Value, error) {
		if values[0] < -1 || values[0] > 1 {
			return nil, fmt.Errorf("asin argument must be between -1 and 1")
		}
		return math.Asin(values[0]), nil
	}))
	runtime.Globals.Define("acos", numericFunction("acos", 1, func(values []float64) (Value, error) {
		if values[0] < -1 || values[0] > 1 {
			return nil, fmt.Errorf("acos argument must be between -1 and 1")
		}
		return math.Acos(values[0]), nil
	}))
	runtime.Globals.Define("atan", numericFunction("atan", 1, func(values []float64) (Value, error) {
		return math.Atan(values[0]), nil
	}))
	runtime.Globals.Define("atan2", numericFunction("atan2", 2, func(values []float64) (Value, error) {
		return math.Atan2(values[0], values[1]), nil
	}))
	runtime.Globals.Define("hypot", numericFunction("hypot", 2, func(values []float64) (Value, error) {
		return math.Hypot(values[0], values[1]), nil
	}))
	runtime.Globals.Define("trunc", numericFunction("trunc", 1, func(values []float64) (Value, error) {
		return math.Trunc(values[0]), nil
	}))
	runtime.Globals.Define("divmod", builtin{"divmod", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("divmod expects 2 integers")
		}
		dividend, ok := arguments[0].(int64)
		if !ok {
			return nil, fmt.Errorf("divmod dividend must be an integer")
		}
		divisor, ok := arguments[1].(int64)
		if !ok {
			return nil, fmt.Errorf("divmod divisor must be an integer")
		}
		if divisor == 0 {
			return nil, fmt.Errorf("divmod divisor cannot be zero")
		}
		if dividend == math.MinInt64 && divisor == -1 {
			return nil, fmt.Errorf("divmod result is out of integer range")
		}
		return []Value{dividend / divisor, dividend % divisor}, nil
	}})
	runtime.Globals.Define("gcd", builtin{"gcd", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("gcd expects 2 integers")
		}
		left, ok := arguments[0].(int64)
		if !ok {
			return nil, fmt.Errorf("gcd first argument must be an integer")
		}
		right, ok := arguments[1].(int64)
		if !ok {
			return nil, fmt.Errorf("gcd second argument must be an integer")
		}
		result := gcdUnsigned(unsignedAbs(left), unsignedAbs(right))
		if result > math.MaxInt64 {
			return nil, fmt.Errorf("gcd result is out of integer range")
		}
		return int64(result), nil
	}})
	runtime.Globals.Define("lcm", builtin{"lcm", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("lcm expects 2 integers")
		}
		left, ok := arguments[0].(int64)
		if !ok {
			return nil, fmt.Errorf("lcm first argument must be an integer")
		}
		right, ok := arguments[1].(int64)
		if !ok {
			return nil, fmt.Errorf("lcm second argument must be an integer")
		}
		leftAbs, rightAbs := unsignedAbs(left), unsignedAbs(right)
		if leftAbs == 0 || rightAbs == 0 {
			return int64(0), nil
		}
		factor := leftAbs / gcdUnsigned(leftAbs, rightAbs)
		high, low := bits.Mul64(factor, rightAbs)
		if high != 0 || low > math.MaxInt64 {
			return nil, fmt.Errorf("lcm result is out of integer range")
		}
		return int64(low), nil
	}})
	runtime.Globals.Define("factorial", builtin{"factorial", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("factorial expects 1 integer")
		}
		value, ok := arguments[0].(int64)
		if !ok || value < 0 {
			return nil, fmt.Errorf("factorial argument must be a non-negative integer")
		}
		if value > 20 {
			return nil, fmt.Errorf("factorial argument exceeds maximum of 20")
		}
		result := int64(1)
		for factor := int64(2); factor <= value; factor++ {
			result *= factor
		}
		return result, nil
	}})
}

func unsignedAbs(value int64) uint64 {
	if value < 0 {
		return uint64(-(value + 1)) + 1
	}
	return uint64(value)
}

func gcdUnsigned(left, right uint64) uint64 {
	for right != 0 {
		left, right = right, left%right
	}
	return left
}
