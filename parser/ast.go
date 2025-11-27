package parser

const (
	NodeType_Undefined = iota
	NodeType_StmtBlock
	NodeType_StmtCallFunction
	NodeType_StmtAssign
	NodeType_ExprBlock
	NodeType_ExprNumber
	NodeType_ExprString
	NodeType_ExprBoolean

	DataType_Number  = "num"
	DataType_String  = "str"
	DataType_Boolean = "bool"
)

var (
	BuiltInDataTypes = map[string]struct{}{
		DataType_Number:  {},
		DataType_String:  {},
		DataType_Boolean: {},
	}
)

type Node interface {
	Type() int
}

type Stmt interface {
	Node
	Stmt()
}

type Expr interface {
	Node
	Expr()
	DataType() string
}

// Ensures all statements and expressions implement the Node interfaces
func (StmtBlock) Type() int        { return NodeType_StmtBlock }
func (StmtCallFunction) Type() int { return NodeType_StmtCallFunction }
func (StmtAssign) Type() int       { return NodeType_StmtAssign }
func (ExprBlock) Type() int        { return NodeType_ExprBlock }
func (ExprNumber) Type() int       { return NodeType_ExprNumber }
func (ExprString) Type() int       { return NodeType_ExprString }
func (ExprBoolean) Type() int      { return NodeType_ExprBoolean }

// Ensures all statements implement the Stmt interface
func (StmtBlock) Stmt()        {}
func (StmtCallFunction) Stmt() {}
func (StmtAssign) Stmt()       {}

// Ensures all expressions implement the Expr interface
func (ExprBlock) Expr()   {}
func (ExprNumber) Expr()  {}
func (ExprString) Expr()  {}
func (ExprBoolean) Expr() {}

type StmtBlock struct {
	Stmts []Stmt
}

type StmtCallFunction struct {
	FunctionName string
	Args         []Expr
}

type StmtAssign struct {
	VariableName string
	Expr         Expr
}

type ExprBlock struct {
	Value Expr
}

func (e ExprBlock) DataType() string {
	return e.Value.DataType()
}

type ExprNumber struct {
	Value string
}

func (e ExprNumber) DataType() string {
	return DataType_Number
}

type ExprString struct {
	Value string
}

func (e ExprString) DataType() string {
	return DataType_String
}

type ExprBoolean struct {
	Value string
}

func (e ExprBoolean) DataType() string {
	return DataType_Boolean
}
