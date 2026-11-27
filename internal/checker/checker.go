package checker

import (
	"fmt"
	"sort"

	"github.com/vanmassey/sitka/internal/ast"
)

type Severity string

const (
	Error   Severity = "error"
	Warning Severity = "warning"
)

type Diagnostic struct {
	Severity Severity
	Code     string
	Message  string
	Line     int
}

func (diagnostic Diagnostic) String() string {
	if diagnostic.Line > 0 {
		return fmt.Sprintf("line %d: %s[%s]: %s", diagnostic.Line, diagnostic.Severity, diagnostic.Code, diagnostic.Message)
	}
	return fmt.Sprintf("%s[%s]: %s", diagnostic.Severity, diagnostic.Code, diagnostic.Message)
}

type symbol struct {
	name  string
	kind  string
	line  int
	arity int
	built bool
}

type scope struct {
	parent  *scope
	symbols map[string]*symbol
}

func newScope(parent *scope) *scope {
	return &scope{parent: parent, symbols: make(map[string]*symbol)}
}

func (scope *scope) resolve(name string) (*symbol, bool) {
	if found, ok := scope.symbols[name]; ok {
		return found, true
	}
	if scope.parent != nil {
		return scope.parent.resolve(name)
	}
	return nil, false
}

type analyzer struct {
	diagnostics []Diagnostic
	loops       int
	functions   int
}

var builtins = []string{
	"abs", "all", "any", "args", "assert", "ceil", "chars", "chunk", "clamp", "clock",
	"contains", "cos", "count", "count_by", "csv_decode", "csv_decode_maps",
	"csv_encode", "csv_encode_maps", "degrees", "drop", "drop_while", "ends_with", "enumerate",
	"e", "exp", "file_exists", "filter", "find", "first", "float", "flatten", "floor",
	"factorial", "divmod", "gcd", "lcm", "ln", "log2", "log10", "tan", "asin",
	"acos", "atan", "atan2", "hypot", "trunc",
	"index_of", "int", "is_finite", "join", "json_decode", "json_encode", "keys",
	"flat_map", "last", "key_by", "len", "list_dir", "lower", "make_dir",
	"map", "max", "max_by", "min", "min_by", "pi", "partition", "path_abs",
	"path_base", "path_clean", "path_dir", "path_ext",
	"path_is_abs", "path_join", "path_rel", "path_separator", "pop", "pow", "push",
	"range", "radians", "read_bytes", "read_lines", "read_text", "reduce", "regex_captures", "regex_find",
	"regex_find_all", "regex_match", "regex_replace", "regex_split", "replace",
	"reversed", "round", "sign", "sin", "slice", "sort_by", "sorted", "split", "sqrt",
	"starts_with", "str", "substring", "sum", "take", "take_while", "time_add",
	"time_diff", "time_format", "time_parse", "time_parts", "time_unix", "time_utc",
	"trim", "type", "unique",
	"upper", "url_encode", "url_decode", "url_parse", "url_build", "query_encode", "query_decode",
	"base64_encode", "base64_decode", "hex_encode", "hex_decode",
	"values", "write_text", "write_lines", "append_text", "write_bytes", "append_bytes",
	"sha256", "sha256_bytes", "zip",
	"cwd", "find_index", "group_by", "rune_at", "ord", "chr", "repeat",
	"pad_left", "pad_right", "split_lines", "join_lines",
}

// Check reports statically detectable errors without executing a program.
func Check(program ast.Program) []Diagnostic {
	root := newScope(nil)
	for _, name := range builtins {
		root.symbols[name] = &symbol{name: name, kind: "builtin", built: true}
	}

	analyzer := &analyzer{}
	analyzer.checkStatements(program.Statements, root)
	sort.SliceStable(analyzer.diagnostics, func(left, right int) bool {
		if analyzer.diagnostics[left].Line == analyzer.diagnostics[right].Line {
			return analyzer.diagnostics[left].Code < analyzer.diagnostics[right].Code
		}
		return analyzer.diagnostics[left].Line < analyzer.diagnostics[right].Line
	})
	return analyzer.diagnostics
}

func (analyzer *analyzer) checkStatements(statements []ast.Stmt, current *scope) {
	analyzer.declareStatements(statements, current)
	for _, statement := range statements {
		analyzer.checkStatement(statement, current)
	}
}

func (analyzer *analyzer) declareStatements(statements []ast.Stmt, current *scope) {
	for _, statement := range statements {
		switch statement := statement.(type) {
		case ast.VarDecl:
			analyzer.declare(current, symbol{name: statement.Name, kind: "variable", line: statement.Line})
		case ast.FunctionDecl:
			analyzer.declare(current, symbol{
				name:  statement.Name,
				kind:  "function",
				line:  statement.Line,
				arity: len(statement.Parameters),
			})
		case ast.ImportStmt:
			analyzer.declare(current, symbol{name: statement.Alias, kind: "module", line: statement.Line})
		case ast.StructDecl:
			analyzer.declare(current, symbol{
				name:  statement.Name,
				kind:  "struct",
				line:  statement.Line,
				arity: len(statement.Fields),
			})
		}
	}
}

func (analyzer *analyzer) declare(current *scope, value symbol) {
	if previous, found := current.symbols[value.name]; found {
		if previous.built {
			copy := value
			current.symbols[value.name] = &copy
			return
		}
		analyzer.report(Error, "duplicate-name", fmt.Sprintf("%q is already declared on line %d", value.name, previous.line), value.line)
		return
	}
	copy := value
	current.symbols[value.name] = &copy
}

func (analyzer *analyzer) checkStatement(statement ast.Stmt, current *scope) {
	switch statement := statement.(type) {
	case ast.VarDecl:
		if statement.Initializer != nil {
			analyzer.checkExpression(statement.Initializer, current)
		}
	case ast.ImportStmt:
		return
	case ast.StructDecl:
		fields := make(map[string]bool, len(statement.Fields))
		for _, field := range statement.Fields {
			if fields[field] {
				analyzer.report(Error, "duplicate-field", fmt.Sprintf("%q is declared more than once in %s", field, statement.Name), statement.Line)
			}
			fields[field] = true
		}
	case ast.FunctionDecl:
		local := newScope(current)
		for _, parameter := range statement.Parameters {
			analyzer.declare(local, symbol{name: parameter, kind: "parameter", line: statement.Line})
		}
		analyzer.functions++
		previousLoops := analyzer.loops
		analyzer.loops = 0
		analyzer.checkStatements(statement.Body.Statements, local)
		analyzer.loops = previousLoops
		analyzer.functions--
	case ast.Block:
		analyzer.checkStatements(statement.Statements, newScope(current))
	case ast.ExpressionStmt:
		analyzer.checkExpression(statement.Expression, current)
	case ast.PrintStmt:
		for _, expression := range statement.Expressions {
			analyzer.checkExpression(expression, current)
		}
	case ast.IfStmt:
		analyzer.checkExpression(statement.Condition, current)
		analyzer.checkStatements(statement.Then.Statements, newScope(current))
		if statement.Else != nil {
			analyzer.checkStatements(statement.Else.Statements, newScope(current))
		}
	case ast.WhileStmt:
		analyzer.checkExpression(statement.Condition, current)
		analyzer.loops++
		analyzer.checkStatements(statement.Body.Statements, newScope(current))
		analyzer.loops--
	case ast.ForStmt:
		analyzer.checkExpression(statement.Iterable, current)
		loopScope := newScope(current)
		analyzer.declare(loopScope, symbol{name: statement.Variable, kind: "loop variable", line: statement.Line})
		analyzer.loops++
		analyzer.checkStatements(statement.Body.Statements, loopScope)
		analyzer.loops--
	case ast.BreakStmt:
		if analyzer.loops == 0 {
			analyzer.report(Error, "break-outside-loop", "'break' can only be used inside a loop", statement.Line)
		}
	case ast.ContinueStmt:
		if analyzer.loops == 0 {
			analyzer.report(Error, "continue-outside-loop", "'continue' can only be used inside a loop", statement.Line)
		}
	case ast.ReturnStmt:
		if analyzer.functions == 0 {
			analyzer.report(Error, "return-outside-function", "'return' can only be used inside a function", statement.Line)
		}
		if statement.Value != nil {
			analyzer.checkExpression(statement.Value, current)
		}
	case ast.TryStmt:
		analyzer.checkStatements(statement.Body.Statements, newScope(current))
		catchScope := newScope(current)
		analyzer.declare(catchScope, symbol{name: statement.ErrorName, kind: "catch variable", line: statement.Line})
		analyzer.checkStatements(statement.Catch.Statements, catchScope)
	case ast.ThrowStmt:
		analyzer.checkExpression(statement.Value, current)
	case ast.Assign:
		analyzer.checkExpression(statement, current)
	case ast.IndexAssign:
		analyzer.checkExpression(statement.Object, current)
		analyzer.checkExpression(statement.Index, current)
		analyzer.checkExpression(statement.Value, current)
	case ast.Set:
		analyzer.checkExpression(statement, current)
	default:
		analyzer.report(Error, "unknown-statement", fmt.Sprintf("cannot check statement type %T", statement), 0)
	}
}

func (analyzer *analyzer) checkExpression(expression ast.Expr, current *scope) {
	switch expression := expression.(type) {
	case ast.Integer, ast.Float, ast.String, ast.Boolean, ast.Nil:
		return
	case ast.Variable:
		if _, found := current.resolve(expression.Name); !found {
			analyzer.report(Error, "undefined-name", fmt.Sprintf("undefined name %q", expression.Name), expression.Line)
		}
	case ast.Assign:
		if _, found := current.resolve(expression.Name); !found {
			analyzer.report(Error, "undefined-name", fmt.Sprintf("cannot assign to undefined name %q", expression.Name), expression.Line)
		}
		analyzer.checkExpression(expression.Value, current)
	case ast.IndexAssign:
		analyzer.checkExpression(expression.Object, current)
		analyzer.checkExpression(expression.Index, current)
		analyzer.checkExpression(expression.Value, current)
	case ast.Set:
		analyzer.checkExpression(expression.Object, current)
		analyzer.checkExpression(expression.Value, current)
	case ast.Unary:
		analyzer.checkExpression(expression.Right, current)
	case ast.Binary:
		analyzer.checkExpression(expression.Left, current)
		analyzer.checkExpression(expression.Right, current)
	case ast.Logical:
		analyzer.checkExpression(expression.Left, current)
		analyzer.checkExpression(expression.Right, current)
	case ast.Call:
		analyzer.checkExpression(expression.Callee, current)
		for _, argument := range expression.Arguments {
			analyzer.checkExpression(argument, current)
		}
		if variable, ok := expression.Callee.(ast.Variable); ok {
			if target, found := current.resolve(variable.Name); found &&
				(target.kind == "function" || target.kind == "struct") &&
				len(expression.Arguments) != target.arity {
				analyzer.report(
					Error,
					"wrong-arity",
					fmt.Sprintf("%s expects %d arguments, got %d", variable.Name, target.arity, len(expression.Arguments)),
					expression.Line,
				)
			}
		}
	case ast.Lambda:
		local := newScope(current)
		for _, parameter := range expression.Parameters {
			analyzer.declare(local, symbol{name: parameter, kind: "parameter", line: expression.Line})
		}
		analyzer.functions++
		previousLoops := analyzer.loops
		analyzer.loops = 0
		analyzer.checkStatements(expression.Body.Statements, local)
		analyzer.loops = previousLoops
		analyzer.functions--
	case ast.List:
		for _, item := range expression.Items {
			analyzer.checkExpression(item, current)
		}
	case ast.Map:
		for _, entry := range expression.Entries {
			analyzer.checkExpression(entry.Key, current)
			analyzer.checkExpression(entry.Value, current)
		}
	case ast.Index:
		analyzer.checkExpression(expression.Object, current)
		analyzer.checkExpression(expression.Index, current)
	case ast.Get:
		analyzer.checkExpression(expression.Object, current)
	default:
		analyzer.report(Error, "unknown-expression", fmt.Sprintf("cannot check expression type %T", expression), 0)
	}
}

func (analyzer *analyzer) report(severity Severity, code, message string, line int) {
	analyzer.diagnostics = append(analyzer.diagnostics, Diagnostic{
		Severity: severity,
		Code:     code,
		Message:  message,
		Line:     line,
	})
}
