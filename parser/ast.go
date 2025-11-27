package parser

const (
	NodeType_Undefined = iota
	NodeType_StmtBlock
	NodeType_StmtCallFunction
	NodeType_ExprBlock
	NodeType_ExprNumber
	NodeType_ExprString
	NodeType_ExprBoolean
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
}

// Ensures all statements and expressions implement the Node interfaces
func (StmtBlock) Type() int        { return NodeType_StmtBlock }
func (StmtCallFunction) Type() int { return NodeType_StmtCallFunction }
func (ExprBlock) Type() int        { return NodeType_ExprBlock }
func (ExprNumber) Type() int       { return NodeType_ExprNumber }
func (ExprString) Type() int       { return NodeType_ExprString }
func (ExprBoolean) Type() int      { return NodeType_ExprBoolean }

// Ensures all statements implement the Stmt interface
func (StmtBlock) Stmt()        {}
func (StmtCallFunction) Stmt() {}

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

type ExprBlock struct {
	Value Expr
}

type ExprNumber struct {
	Value string
}

type ExprString struct {
	Value string
}

type ExprBoolean struct {
	Value string
}
