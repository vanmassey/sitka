package ast

// Node is implemented by every type in the Sitka syntax tree.
type Node interface {
	isNode()
}

type Expr interface {
	Node
	isExpr()
}

type Stmt interface {
	Node
	isStmt()
}

type Program struct {
	Statements []Stmt
	Line       int
}

func (Program) isNode() {}

type Block struct {
	Statements []Stmt
	Line       int
}

func (Block) isNode() {}
func (Block) isStmt() {}

type VarDecl struct {
	Name        string
	Initializer Expr
	Line        int
}

func (VarDecl) isNode() {}
func (VarDecl) isStmt() {}

type Assign struct {
	Name  string
	Value Expr
	Line  int
}

func (Assign) isNode() {}
func (Assign) isStmt() {}
func (Assign) isExpr() {}

type IndexAssign struct {
	Object Expr
	Index  Expr
	Value  Expr
	Line   int
}

func (IndexAssign) isNode() {}
func (IndexAssign) isStmt() {}
func (IndexAssign) isExpr() {}

type ExpressionStmt struct {
	Expression Expr
	Line       int
}

func (ExpressionStmt) isNode() {}
func (ExpressionStmt) isStmt() {}

type PrintStmt struct {
	Expressions []Expr
	Line        int
}

func (PrintStmt) isNode() {}
func (PrintStmt) isStmt() {}

type IfStmt struct {
	Condition Expr
	Then      Block
	Else      *Block
	Line      int
}

func (IfStmt) isNode() {}
func (IfStmt) isStmt() {}

type WhileStmt struct {
	Condition Expr
	Body      Block
	Line      int
}

func (WhileStmt) isNode() {}
func (WhileStmt) isStmt() {}

type ForStmt struct {
	Variable string
	Iterable Expr
	Body     Block
	Line     int
}

func (ForStmt) isNode() {}
func (ForStmt) isStmt() {}

type BreakStmt struct {
	Line int
}

func (BreakStmt) isNode() {}
func (BreakStmt) isStmt() {}

type ContinueStmt struct {
	Line int
}

func (ContinueStmt) isNode() {}
func (ContinueStmt) isStmt() {}

type ReturnStmt struct {
	Value Expr
	Line  int
}

func (ReturnStmt) isNode() {}
func (ReturnStmt) isStmt() {}

type FunctionDecl struct {
	Name       string
	Parameters []string
	Body       Block
	Line       int
}

func (FunctionDecl) isNode() {}
func (FunctionDecl) isStmt() {}

type Lambda struct {
	Parameters []string
	Body       Block
	Line       int
}

func (Lambda) isNode() {}
func (Lambda) isExpr() {}

type Integer struct {
	Value int64
	Line  int
}

func (Integer) isNode() {}
func (Integer) isExpr() {}

type Float struct {
	Value float64
	Line  int
}

func (Float) isNode() {}
func (Float) isExpr() {}

type String struct {
	Value string
	Line  int
}

func (String) isNode() {}
func (String) isExpr() {}

type Boolean struct {
	Value bool
	Line  int
}

func (Boolean) isNode() {}
func (Boolean) isExpr() {}

type Nil struct {
	Line int
}

func (Nil) isNode() {}
func (Nil) isExpr() {}

type Variable struct {
	Name string
	Line int
}

func (Variable) isNode() {}
func (Variable) isExpr() {}

type Unary struct {
	Operator string
	Right    Expr
	Line     int
}

func (Unary) isNode() {}
func (Unary) isExpr() {}

type Binary struct {
	Left     Expr
	Operator string
	Right    Expr
	Line     int
}

func (Binary) isNode() {}
func (Binary) isExpr() {}

type Logical struct {
	Left     Expr
	Operator string
	Right    Expr
	Line     int
}

func (Logical) isNode() {}
func (Logical) isExpr() {}

type Call struct {
	Callee    Expr
	Arguments []Expr
	Line      int
}

func (Call) isNode() {}
func (Call) isExpr() {}

type List struct {
	Items []Expr
	Line  int
}

func (List) isNode() {}
func (List) isExpr() {}

type Map struct {
	Entries []MapEntry
	Line    int
}

type MapEntry struct {
	Key   Expr
	Value Expr
}

func (Map) isNode() {}
func (Map) isExpr() {}

type Index struct {
	Object Expr
	Index  Expr
	Line   int
}

func (Index) isNode() {}
func (Index) isExpr() {}

type Get struct {
	Object Expr
	Name   string
	Line   int
}

func (Get) isNode() {}
func (Get) isExpr() {}
