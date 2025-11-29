package parser

import "pixie/shared"

const (
	NodeType_Undefined = iota
	NodeType_StmtBlock
	NodeType_StmtCallFunction
	NodeType_StmtVarDeclare
	NodeType_StmtVarAssign
	NodeType_StmtObjDefine
	NodeType_ExprBlock
	NodeType_ExprNumber
	NodeType_ExprString
	NodeType_ExprBoolean
	NodeType_ExprList
	NodeType_ExprMap
	NodeType_ExprVariable
	NodeType_ExprIndex
	NodeType_ExprPropertyAccess
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
func (StmtVarDeclare) Type() int   { return NodeType_StmtVarDeclare }
func (StmtVarAssign) Type() int    { return NodeType_StmtVarAssign }
func (StmtObjDefine) Type() int    { return NodeType_StmtObjDefine }
func (ExprBlock) Type() int        { return NodeType_ExprBlock }
func (ExprNumber) Type() int       { return NodeType_ExprNumber }
func (ExprString) Type() int       { return NodeType_ExprString }
func (ExprBoolean) Type() int      { return NodeType_ExprBoolean }
func (ExprList) Type() int         { return NodeType_ExprList }
func (ExprTable) Type() int        { return NodeType_ExprMap }
func (ExprVariable) Type() int     { return NodeType_ExprVariable }
func (ExprIndex) Type() int        { return NodeType_ExprIndex }
func (ExprPropertyAccess) Type() int { return NodeType_ExprPropertyAccess }

// Ensures all statements implement the Stmt interface
func (StmtBlock) Stmt()        {}
func (StmtCallFunction) Stmt() {}
func (StmtVarDeclare) Stmt()   {}
func (StmtVarAssign) Stmt()    {}
func (StmtObjDefine) Stmt()    {}

// Ensures all expressions implement the Expr interface
func (ExprBlock) Expr()    {}
func (ExprNumber) Expr()   {}
func (ExprString) Expr()   {}
func (ExprBoolean) Expr()  {}
func (ExprList) Expr()     {}
func (ExprTable) Expr()    {}
func (ExprVariable) Expr() {}
func (ExprIndex) Expr()    {}
func (ExprPropertyAccess) Expr() {}

type StmtBlock struct {
	Stmts []Stmt
}

type StmtCallFunction struct {
	FunctionName string
	Args         []Expr
}

type StmtVarDeclare struct {
	VariableName string
	DataType     shared.DataType
	Expr         Expr
}

type StmtVarAssign struct {
	VariableName string
	Expr         Expr
}

type FieldTypePair struct {
	Field string
	Type  shared.DataType
}

type StmtObjDefine struct {
	Name   string
	Fields []FieldTypePair
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

type ExprList struct {
	Values []Expr
}

type TablePair struct {
	Key   Expr
	Value Expr
}

type ExprTable struct {
	Pairs []TablePair
}

type ExprVariable struct {
	Name string
}

type ExprIndex struct {
	Left  Expr
	Index Expr
}

type ExprPropertyAccess struct {
	Left   Expr
	Property string
}
