package interpreter

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/vanmassey/sitka/internal/ast"
)

type Value any

const maxGeneratedItems = 1_000_000
const maxGeneratedStringBytes = 16 * 1024 * 1024

type Environment struct {
	values map[string]Value
	parent *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{values: make(map[string]Value), parent: parent}
}

func (environment *Environment) Define(name string, value Value) {
	environment.values[name] = value
}

func (environment *Environment) Get(name string) (Value, bool) {
	value, ok := environment.values[name]
	if ok {
		return value, true
	}
	if environment.parent != nil {
		return environment.parent.Get(name)
	}
	return nil, false
}

func (environment *Environment) Assign(name string, value Value) bool {
	if _, ok := environment.values[name]; ok {
		environment.values[name] = value
		return true
	}
	if environment.parent != nil {
		return environment.parent.Assign(name, value)
	}
	return false
}

type Runtime struct {
	Globals *Environment
}

type controlKind int

const (
	controlNone controlKind = iota
	controlBreak
	controlContinue
	controlReturn
)

type control struct {
	kind  controlKind
	value Value
}

type function struct {
	name       string
	parameters []string
	body       ast.Block
	closure    *Environment
}

type builtin struct {
	name string
	call func([]Value) (Value, error)
}

func New() *Runtime {
	globals := NewEnvironment(nil)
	runtime := &Runtime{Globals: globals}
	runtime.installBuiltins()
	return runtime
}

func (runtime *Runtime) installBuiltins() {
	runtime.Globals.Define("len", builtin{"len", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("len expects 1 argument")
		}
		switch value := arguments[0].(type) {
		case string:
			return int64(len([]rune(value))), nil
		case []Value:
			return int64(len(value)), nil
		case map[string]Value:
			return int64(len(value)), nil
		default:
			return nil, fmt.Errorf("len cannot measure %s", TypeName(value))
		}
	}})
	runtime.Globals.Define("str", builtin{"str", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("str expects 1 argument")
		}
		return Format(arguments[0]), nil
	}})
	runtime.Globals.Define("int", builtin{"int", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("int expects 1 argument")
		}
		switch value := arguments[0].(type) {
		case int64:
			return value, nil
		case float64:
			if value >= math.Exp2(63) || value < -math.Exp2(63) {
				return nil, fmt.Errorf("integer out of range")
			}
			return int64(value), nil
		case string:
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot convert %q to int", value)
			}
			return parsed, nil
		default:
			return nil, fmt.Errorf("cannot convert %s to int", TypeName(value))
		}
	}})
	runtime.Globals.Define("float", builtin{"float", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("float expects 1 argument")
		}
		switch value := arguments[0].(type) {
		case int64:
			return float64(value), nil
		case float64:
			return value, nil
		case string:
			parsed, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot convert %q to float", value)
			}
			return parsed, nil
		default:
			return nil, fmt.Errorf("cannot convert %s to float", TypeName(value))
		}
	}})
	runtime.Globals.Define("type", builtin{"type", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("type expects 1 argument")
		}
		return TypeName(arguments[0]), nil
	}})
	runtime.Globals.Define("range", builtin{"range", func(arguments []Value) (Value, error) {
		if len(arguments) < 1 || len(arguments) > 3 {
			return nil, fmt.Errorf("range expects 1 to 3 arguments")
		}
		numbers := make([]int64, len(arguments))
		for index, argument := range arguments {
			number, ok := argument.(int64)
			if !ok {
				return nil, fmt.Errorf("range arguments must be integers")
			}
			numbers[index] = number
		}
		start, end, step := int64(0), int64(0), int64(1)
		switch len(numbers) {
		case 1:
			end = numbers[0]
		case 2:
			start, end = numbers[0], numbers[1]
		case 3:
			start, end, step = numbers[0], numbers[1], numbers[2]
		}
		if step == 0 {
			return nil, fmt.Errorf("range step cannot be zero")
		}
		values := make([]Value, 0)
		if step > 0 {
			for value := start; value < end; value += step {
				if len(values) >= maxGeneratedItems {
					return nil, fmt.Errorf("range exceeds maximum size of %d items", maxGeneratedItems)
				}
				values = append(values, value)
				if value > math.MaxInt64-step {
					break
				}
			}
		} else {
			for value := start; value > end; value += step {
				if len(values) >= maxGeneratedItems {
					return nil, fmt.Errorf("range exceeds maximum size of %d items", maxGeneratedItems)
				}
				values = append(values, value)
				if value < math.MinInt64-step {
					break
				}
			}
		}
		return values, nil
	}})
	runtime.Globals.Define("push", builtin{"push", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("push expects 2 arguments")
		}
		list, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("push first argument must be a list")
		}
		result := make([]Value, 0, len(list)+1)
		result = append(result, list...)
		result = append(result, arguments[1])
		return result, nil
	}})
	runtime.Globals.Define("pop", builtin{"pop", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("pop expects 1 argument")
		}
		list, ok := arguments[0].([]Value)
		if !ok {
			return nil, fmt.Errorf("pop argument must be a list")
		}
		if len(list) == 0 {
			return nil, fmt.Errorf("cannot pop an empty list")
		}
		return append([]Value(nil), list[:len(list)-1]...), nil
	}})
	runtime.Globals.Define("abs", builtin{"abs", func(arguments []Value) (Value, error) {
		if len(arguments) != 1 {
			return nil, fmt.Errorf("abs expects 1 argument")
		}
		switch value := arguments[0].(type) {
		case int64:
			if value == math.MinInt64 {
				return nil, fmt.Errorf("integer overflow")
			}
			if value < 0 {
				return -value, nil
			}
			return value, nil
		case float64:
			return math.Abs(value), nil
		default:
			return nil, fmt.Errorf("abs expects a number")
		}
	}})
	runtime.Globals.Define("min", builtin{"min", func(arguments []Value) (Value, error) {
		if len(arguments) == 0 {
			return nil, fmt.Errorf("min expects at least 1 argument")
		}
		first, ok := arguments[0].(int64)
		if !ok {
			return nil, fmt.Errorf("min arguments must be integers")
		}
		for _, argument := range arguments[1:] {
			number, ok := argument.(int64)
			if !ok {
				return nil, fmt.Errorf("min arguments must be integers")
			}
			if number < first {
				first = number
			}
		}
		return first, nil
	}})
	runtime.Globals.Define("max", builtin{"max", func(arguments []Value) (Value, error) {
		if len(arguments) == 0 {
			return nil, fmt.Errorf("max expects at least 1 argument")
		}
		first, ok := arguments[0].(int64)
		if !ok {
			return nil, fmt.Errorf("max arguments must be integers")
		}
		for _, argument := range arguments[1:] {
			number, ok := argument.(int64)
			if !ok {
				return nil, fmt.Errorf("max arguments must be integers")
			}
			if number > first {
				first = number
			}
		}
		return first, nil
	}})
	runtime.Globals.Define("clock", builtin{"clock", func(arguments []Value) (Value, error) {
		if len(arguments) != 0 {
			return nil, fmt.Errorf("clock expects no arguments")
		}
		return float64(time.Now().UnixNano()) / float64(time.Second), nil
	}})
	runtime.Globals.Define("contains", builtin{"contains", func(arguments []Value) (Value, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("contains expects 2 arguments")
		}
		switch container := arguments[0].(type) {
		case string:
			needle, ok := arguments[1].(string)
			return ok && strings.Contains(container, needle), nil
		case []Value:
			return sliceContains(container, arguments[1]), nil
		case map[string]Value:
			needle, ok := arguments[1].(string)
			if !ok {
				return false, nil
			}
			_, found := container[needle]
			return found, nil
		default:
			return nil, fmt.Errorf("contains cannot search %s", TypeName(container))
		}
	}})
	runtime.installExtendedBuiltins()
}

func (runtime *Runtime) Execute(program ast.Program) (Value, error) {
	value, control, err := runtime.executeBlock(ast.Block{Statements: program.Statements}, runtime.Globals)
	if err != nil {
		return nil, err
	}
	if control.kind != controlNone {
		return nil, fmt.Errorf("unexpected control flow")
	}
	return value, nil
}

func (runtime *Runtime) executeBlock(block ast.Block, environment *Environment) (Value, control, error) {
	var last Value
	for _, statement := range block.Statements {
		value, result, err := runtime.executeStatement(statement, environment)
		if err != nil {
			return nil, control{}, err
		}
		if result.kind != controlNone {
			return value, result, nil
		}
		last = value
	}
	return last, control{}, nil
}

func (runtime *Runtime) executeStatement(statement ast.Stmt, environment *Environment) (Value, control, error) {
	switch statement := statement.(type) {
	case ast.VarDecl:
		var value Value
		if statement.Initializer != nil {
			var err error
			value, err = runtime.evaluate(statement.Initializer, environment)
			if err != nil {
				return nil, control{}, err
			}
		}
		environment.Define(statement.Name, value)
		return nil, control{}, nil
	case ast.Assign:
		value, err := runtime.evaluate(statement.Value, environment)
		if err != nil {
			return nil, control{}, err
		}
		if !environment.Assign(statement.Name, value) {
			return nil, control{}, fmt.Errorf("undefined variable %q at line %d", statement.Name, statement.Line)
		}
		return nil, control{}, nil
	case ast.IndexAssign:
		object, err := runtime.evaluate(statement.Object, environment)
		if err != nil {
			return nil, control{}, err
		}
		index, err := runtime.evaluate(statement.Index, environment)
		if err != nil {
			return nil, control{}, err
		}
		value, err := runtime.evaluate(statement.Value, environment)
		if err != nil {
			return nil, control{}, err
		}
		switch container := object.(type) {
		case []Value:
			position, ok := index.(int64)
			if !ok {
				return nil, control{}, fmt.Errorf("list index must be an integer")
			}
			resolved, err := normalizeIndex(position, len(container))
			if err != nil {
				return nil, control{}, err
			}
			container[resolved] = value
		case map[string]Value:
			key, ok := index.(string)
			if !ok {
				return nil, control{}, fmt.Errorf("map key must be a string")
			}
			container[key] = value
		default:
			return nil, control{}, fmt.Errorf("cannot assign an index on %s", TypeName(object))
		}
		return nil, control{}, nil
	case ast.ExpressionStmt:
		value, err := runtime.evaluate(statement.Expression, environment)
		return value, control{}, err
	case ast.PrintStmt:
		parts := make([]string, 0, len(statement.Expressions))
		for _, expression := range statement.Expressions {
			value, err := runtime.evaluate(expression, environment)
			if err != nil {
				return nil, control{}, err
			}
			parts = append(parts, Format(value))
		}
		fmt.Println(strings.Join(parts, " "))
		return nil, control{}, nil
	case ast.IfStmt:
		condition, err := runtime.evaluate(statement.Condition, environment)
		if err != nil {
			return nil, control{}, err
		}
		if Truthy(condition) {
			return runtime.executeBlock(statement.Then, NewEnvironment(environment))
		}
		if statement.Else != nil {
			return runtime.executeBlock(*statement.Else, NewEnvironment(environment))
		}
		return nil, control{}, nil
	case ast.WhileStmt:
		for {
			condition, err := runtime.evaluate(statement.Condition, environment)
			if err != nil {
				return nil, control{}, err
			}
			if !Truthy(condition) {
				break
			}
			value, result, err := runtime.executeBlock(statement.Body, NewEnvironment(environment))
			if err != nil {
				return nil, control{}, err
			}
			if result.kind == controlBreak {
				return value, control{}, nil
			}
			if result.kind == controlReturn {
				return value, result, nil
			}
		}
		return nil, control{}, nil
	case ast.ForStmt:
		iterable, err := runtime.evaluate(statement.Iterable, environment)
		if err != nil {
			return nil, control{}, err
		}
		values, ok := iterable.([]Value)
		if !ok {
			return nil, control{}, fmt.Errorf("'for' iterable must be a list")
		}
		for _, item := range values {
			loopEnvironment := NewEnvironment(environment)
			loopEnvironment.Define(statement.Variable, item)
			value, result, err := runtime.executeBlock(statement.Body, loopEnvironment)
			if err != nil {
				return nil, control{}, err
			}
			if result.kind == controlBreak {
				return value, control{}, nil
			}
			if result.kind == controlContinue {
				continue
			}
			if result.kind == controlReturn {
				return value, result, nil
			}
		}
		return nil, control{}, nil
	case ast.BreakStmt:
		return nil, control{kind: controlBreak}, nil
	case ast.ContinueStmt:
		return nil, control{kind: controlContinue}, nil
	case ast.ReturnStmt:
		var value Value
		if statement.Value != nil {
			var err error
			value, err = runtime.evaluate(statement.Value, environment)
			if err != nil {
				return nil, control{}, err
			}
		}
		return value, control{kind: controlReturn, value: value}, nil
	case ast.FunctionDecl:
		runtimeFunction := &function{
			name:       statement.Name,
			parameters: statement.Parameters,
			body:       statement.Body,
			closure:    environment,
		}
		environment.Define(statement.Name, runtimeFunction)
		return nil, control{}, nil
	case ast.Block:
		return runtime.executeBlock(statement, NewEnvironment(environment))
	default:
		return nil, control{}, fmt.Errorf("unknown statement")
	}
}

func (runtime *Runtime) evaluate(expression ast.Expr, environment *Environment) (Value, error) {
	switch expression := expression.(type) {
	case ast.Integer:
		return expression.Value, nil
	case ast.Float:
		return expression.Value, nil
	case ast.String:
		return expression.Value, nil
	case ast.Boolean:
		return expression.Value, nil
	case ast.Nil:
		return nil, nil
	case ast.Variable:
		value, ok := environment.Get(expression.Name)
		if !ok {
			return nil, fmt.Errorf("undefined variable %q at line %d", expression.Name, expression.Line)
		}
		return value, nil
	case ast.Lambda:
		return &function{
			name:       "<anonymous>",
			parameters: expression.Parameters,
			body:       expression.Body,
			closure:    environment,
		}, nil
	case ast.Assign:
		value, err := runtime.evaluate(expression.Value, environment)
		if err != nil {
			return nil, err
		}
		if !environment.Assign(expression.Name, value) {
			return nil, fmt.Errorf("undefined variable %q at line %d", expression.Name, expression.Line)
		}
		return value, nil
	case ast.Unary:
		right, err := runtime.evaluate(expression.Right, environment)
		if err != nil {
			return nil, err
		}
		switch expression.Operator {
		case "!":
			return !Truthy(right), nil
		case "-":
			switch number := right.(type) {
			case int64:
				if number == math.MinInt64 {
					return nil, fmt.Errorf("integer overflow at line %d", expression.Line)
				}
				return -number, nil
			case float64:
				return -number, nil
			default:
				return nil, fmt.Errorf("cannot negate %s at line %d", TypeName(right), expression.Line)
			}
		default:
			return nil, fmt.Errorf("unknown unary operator %q", expression.Operator)
		}
	case ast.Binary:
		left, err := runtime.evaluate(expression.Left, environment)
		if err != nil {
			return nil, err
		}
		right, err := runtime.evaluate(expression.Right, environment)
		if err != nil {
			return nil, err
		}
		return evaluateBinary(expression.Operator, left, right, expression.Line)
	case ast.IndexAssign:
		object, err := runtime.evaluate(expression.Object, environment)
		if err != nil {
			return nil, err
		}
		index, err := runtime.evaluate(expression.Index, environment)
		if err != nil {
			return nil, err
		}
		value, err := runtime.evaluate(expression.Value, environment)
		if err != nil {
			return nil, err
		}
		switch container := object.(type) {
		case []Value:
			position, ok := index.(int64)
			if !ok {
				return nil, fmt.Errorf("list index must be an integer")
			}
			resolved, err := normalizeIndex(position, len(container))
			if err != nil {
				return nil, err
			}
			container[resolved] = value
		case map[string]Value:
			key, ok := index.(string)
			if !ok {
				return nil, fmt.Errorf("map key must be a string")
			}
			container[key] = value
		default:
			return nil, fmt.Errorf("cannot assign an index on %s", TypeName(object))
		}
		return value, nil
	case ast.Logical:
		left, err := runtime.evaluate(expression.Left, environment)
		if err != nil {
			return nil, err
		}
		if expression.Operator == "&&" {
			if !Truthy(left) {
				return false, nil
			}
		} else if Truthy(left) {
			return true, nil
		}
		right, err := runtime.evaluate(expression.Right, environment)
		if err != nil {
			return nil, err
		}
		if expression.Operator == "&&" {
			return Truthy(right), nil
		}
		return Truthy(right), nil
	case ast.List:
		items := make([]Value, 0, len(expression.Items))
		for _, itemExpression := range expression.Items {
			item, err := runtime.evaluate(itemExpression, environment)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		}
		return items, nil
	case ast.Map:
		items := make(map[string]Value)
		for _, entry := range expression.Entries {
			key, err := runtime.evaluate(entry.Key, environment)
			if err != nil {
				return nil, err
			}
			value, err := runtime.evaluate(entry.Value, environment)
			if err != nil {
				return nil, err
			}
			keyText, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("map keys must be strings at line %d", expression.Line)
			}
			items[keyText] = value
		}
		return items, nil
	case ast.Index:
		object, err := runtime.evaluate(expression.Object, environment)
		if err != nil {
			return nil, err
		}
		index, err := runtime.evaluate(expression.Index, environment)
		if err != nil {
			return nil, err
		}
		switch container := object.(type) {
		case []Value:
			position, ok := index.(int64)
			if !ok {
				return nil, fmt.Errorf("list index must be an integer")
			}
			resolved, err := normalizeIndex(position, len(container))
			if err != nil {
				return nil, err
			}
			return container[resolved], nil
		case map[string]Value:
			key, ok := index.(string)
			if !ok {
				return nil, fmt.Errorf("map key must be a string")
			}
			value, found := container[key]
			if !found {
				return nil, nil
			}
			return value, nil
		case string:
			position, ok := index.(int64)
			if !ok {
				return nil, fmt.Errorf("string index must be an integer")
			}
			runes := []rune(container)
			resolved, err := normalizeIndex(position, len(runes))
			if err != nil {
				return nil, err
			}
			return string(runes[resolved]), nil
		default:
			return nil, fmt.Errorf("cannot index %s", TypeName(object))
		}
	case ast.Get:
		object, err := runtime.evaluate(expression.Object, environment)
		if err != nil {
			return nil, err
		}
		if mapping, ok := object.(map[string]Value); ok {
			value, found := mapping[expression.Name]
			if !found {
				return nil, fmt.Errorf("map has no key %q", expression.Name)
			}
			return value, nil
		}
		return nil, fmt.Errorf("cannot access property %q on %s", expression.Name, TypeName(object))
	case ast.Call:
		callee, err := runtime.evaluate(expression.Callee, environment)
		if err != nil {
			return nil, err
		}
		arguments := make([]Value, 0, len(expression.Arguments))
		for _, argumentExpression := range expression.Arguments {
			argument, err := runtime.evaluate(argumentExpression, environment)
			if err != nil {
				return nil, err
			}
			arguments = append(arguments, argument)
		}
		return runtime.call(callee, arguments)
	default:
		return nil, fmt.Errorf("unknown expression")
	}
}

func (runtime *Runtime) call(callee Value, arguments []Value) (Value, error) {
	switch callable := callee.(type) {
	case builtin:
		return callable.call(arguments)
	case *function:
		if len(arguments) != len(callable.parameters) {
			return nil, fmt.Errorf("function %q expects %d arguments, got %d", callable.name, len(callable.parameters), len(arguments))
		}
		local := NewEnvironment(callable.closure)
		for index, parameter := range callable.parameters {
			local.Define(parameter, arguments[index])
		}
		value, result, err := runtime.executeBlock(callable.body, local)
		if err != nil {
			return nil, err
		}
		if result.kind == controlReturn {
			return result.value, nil
		}
		return value, nil
	default:
		return nil, fmt.Errorf("%s is not callable", TypeName(callee))
	}
}

func evaluateBinary(operator string, left, right Value, line int) (Value, error) {
	switch operator {
	case "==":
		return Equal(left, right), nil
	case "!=":
		return !Equal(left, right), nil
	case "+":
		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case int64:
				if (right > 0 && left > math.MaxInt64-right) || (right < 0 && left < math.MinInt64-right) {
					return nil, fmt.Errorf("integer overflow at line %d", line)
				}
				return left + right, nil
			case float64:
				return float64(left) + right, nil
			}
		case float64:
			switch right := right.(type) {
			case float64:
				return left + right, nil
			case int64:
				return left + float64(right), nil
			}
		case string:
			if rightString, ok := right.(string); ok {
				return left + rightString, nil
			}
		case []Value:
			if rightList, ok := right.([]Value); ok {
				result := make([]Value, 0, len(left)+len(rightList))
				result = append(result, left...)
				result = append(result, rightList...)
				return result, nil
			}
		}
	case "-":
		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case int64:
				if (right < 0 && left > math.MaxInt64+right) || (right > 0 && left < math.MinInt64+right) {
					return nil, fmt.Errorf("integer overflow at line %d", line)
				}
				return left - right, nil
			case float64:
				return float64(left) - right, nil
			}
		case float64:
			switch right := right.(type) {
			case float64:
				return left - right, nil
			case int64:
				return left - float64(right), nil
			}
		}
	case "*":
		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case int64:
				product, ok := checkedMultiply(left, right)
				if !ok {
					return nil, fmt.Errorf("integer overflow at line %d", line)
				}
				return product, nil
			case float64:
				return float64(left) * right, nil
			}
		case float64:
			switch right := right.(type) {
			case float64:
				return left * right, nil
			case int64:
				return left * float64(right), nil
			}
		case string:
			if count, ok := right.(int64); ok {
				if count < 0 {
					return nil, fmt.Errorf("cannot repeat a string a negative number of times")
				}
				if len(left) > 0 && count > int64(maxGeneratedStringBytes/len(left)) {
					return nil, fmt.Errorf("repeated string exceeds maximum size of %d bytes", maxGeneratedStringBytes)
				}
				return strings.Repeat(left, int(count)), nil
			}
		}
	case "/":
		switch left := left.(type) {
		case int64:
			switch right := right.(type) {
			case int64:
				if right == 0 {
					return nil, fmt.Errorf("division by zero at line %d", line)
				}
				if left == math.MinInt64 && right == -1 {
					return nil, fmt.Errorf("integer overflow at line %d", line)
				}
				return left / right, nil
			case float64:
				if right == 0 {
					return nil, fmt.Errorf("division by zero at line %d", line)
				}
				return float64(left) / right, nil
			}
		case float64:
			switch right := right.(type) {
			case float64:
				if right == 0 {
					return nil, fmt.Errorf("division by zero at line %d", line)
				}
				return left / right, nil
			case int64:
				if right == 0 {
					return nil, fmt.Errorf("division by zero at line %d", line)
				}
				return left / float64(right), nil
			}
		}
	case "%":
		leftNumber, leftOK := left.(int64)
		rightNumber, rightOK := right.(int64)
		if leftOK && rightOK {
			if rightNumber == 0 {
				return nil, fmt.Errorf("division by zero at line %d", line)
			}
			return leftNumber % rightNumber, nil
		}
	case "<", "<=", ">", ">=":
		comparison, comparable := compare(left, right)
		if !comparable {
			break
		}
		switch operator {
		case "<":
			return comparison < 0, nil
		case "<=":
			return comparison <= 0, nil
		case ">":
			return comparison > 0, nil
		case ">=":
			return comparison >= 0, nil
		}
	}
	return nil, fmt.Errorf("operator %q cannot be applied to %s and %s at line %d", operator, TypeName(left), TypeName(right), line)
}

func checkedMultiply(left, right int64) (int64, bool) {
	if left == 0 || right == 0 {
		return 0, true
	}
	if left > 0 {
		if right > 0 && left > math.MaxInt64/right {
			return 0, false
		}
		if right < 0 && right < math.MinInt64/left {
			return 0, false
		}
	} else {
		if right > 0 && left < math.MinInt64/right {
			return 0, false
		}
		if right < 0 && left < math.MaxInt64/right {
			return 0, false
		}
	}
	return left * right, true
}

func compare(left, right Value) (int, bool) {
	switch left := left.(type) {
	case int64:
		switch right := right.(type) {
		case int64:
			if left < right {
				return -1, true
			}
			if left > right {
				return 1, true
			}
			return 0, true
		case float64:
			return compareFloat(float64(left), right)
		}
	case float64:
		switch right := right.(type) {
		case float64:
			return compareFloat(left, right)
		case int64:
			return compareFloat(left, float64(right))
		}
	case string:
		rightString, ok := right.(string)
		if !ok {
			return 0, false
		}
		if left < rightString {
			return -1, true
		}
		if left > rightString {
			return 1, true
		}
		return 0, true
	}
	return 0, false
}

func compareFloat(left, right float64) (int, bool) {
	if math.IsNaN(left) || math.IsNaN(right) {
		return 0, false
	}
	if left < right {
		return -1, true
	}
	if left > right {
		return 1, true
	}
	return 0, true
}

func Equal(left, right Value) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	switch left := left.(type) {
	case int64:
		switch right := right.(type) {
		case int64:
			return left == right
		case float64:
			return float64(left) == right
		}
	case float64:
		switch right := right.(type) {
		case float64:
			return left == right
		case int64:
			return left == float64(right)
		}
	case bool:
		rightBool, ok := right.(bool)
		return ok && left == rightBool
	case string:
		rightString, ok := right.(string)
		return ok && left == rightString
	case []Value:
		rightList, ok := right.([]Value)
		if !ok || len(left) != len(rightList) {
			return false
		}
		for index := range left {
			if !Equal(left[index], rightList[index]) {
				return false
			}
		}
		return true
	case map[string]Value:
		rightMap, ok := right.(map[string]Value)
		if !ok || len(left) != len(rightMap) {
			return false
		}
		for key, value := range left {
			other, found := rightMap[key]
			if !found || !Equal(value, other) {
				return false
			}
		}
		return true
	}
	return false
}

func Truthy(value Value) bool {
	switch value := value.(type) {
	case nil:
		return false
	case bool:
		return value
	default:
		return true
	}
}

func TypeName(value Value) string {
	switch value.(type) {
	case nil:
		return "nil"
	case bool:
		return "bool"
	case int64:
		return "int"
	case float64:
		return "float"
	case string:
		return "string"
	case []Value:
		return "list"
	case map[string]Value:
		return "map"
	case builtin, *function:
		return "function"
	default:
		return "unknown"
	}
}

func Format(value Value) string {
	switch value := value.(type) {
	case nil:
		return "nil"
	case bool:
		if value {
			return "true"
		}
		return "false"
	case int64:
		return strconv.FormatInt(value, 10)
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64)
	case string:
		return value
	case []Value:
		parts := make([]string, len(value))
		for index, item := range value {
			parts[index] = Format(item)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]Value:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, strconv.Quote(key)+": "+Format(value[key]))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case builtin:
		return "<builtin " + value.name + ">"
	case *function:
		return "<function " + value.name + ">"
	default:
		return fmt.Sprintf("%v", value)
	}
}

func normalizeIndex(index int64, length int) (int, error) {
	if length == 0 {
		return 0, fmt.Errorf("index out of range")
	}
	position := index
	if position < 0 {
		position = int64(length) + position
	}
	if position < 0 || position >= int64(length) {
		return 0, fmt.Errorf("index out of range")
	}
	return int(position), nil
}

func sliceContains(values []Value, target Value) bool {
	for _, value := range values {
		if Equal(value, target) {
			return true
		}
	}
	return false
}
