package parser

import "pixie/shared"

const (
	NodeType_Undefined = iota
	NodeType_StmtBlock
	NodeType_StmtCallFunction
	NodeType_StmtVarDeclare
	NodeType_StmtVarAssign
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
	DataType() string
}

// Ensures all statements and expressions implement the Node interfaces
func (StmtBlock) Type() int        { return NodeType_StmtBlock }
func (StmtCallFunction) Type() int { return NodeType_StmtCallFunction }
func (StmtVarDeclare) Type() int   { return NodeType_StmtVarDeclare }
func (StmtVarAssign) Type() int    { return NodeType_StmtVarAssign }
func (ExprBlock) Type() int        { return NodeType_ExprBlock }
func (ExprNumber) Type() int       { return NodeType_ExprNumber }
func (ExprString) Type() int       { return NodeType_ExprString }
func (ExprBoolean) Type() int      { return NodeType_ExprBoolean }

// Ensures all statements implement the Stmt interface
func (StmtBlock) Stmt()        {}
func (StmtCallFunction) Stmt() {}
func (StmtVarDeclare) Stmt()   {}
func (StmtVarAssign) Stmt()    {}

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

type StmtVarDeclare struct {
	VariableName string
	DataType     string
	Expr         Expr
}

type StmtVarAssign struct {
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
	return shared.Keyword_Number
}

type ExprString struct {
	Value string
}

func (e ExprString) DataType() string {
	return shared.Keyword_String
}

type ExprBoolean struct {
	Value string
}

func (e ExprBoolean) DataType() string {
	return shared.Keyword_Boolean
}
