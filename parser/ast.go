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
	NodeType_ExprList
	NodeType_ExprMap
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
	DataType() shared.DataType
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
func (ExprList) Type() int         { return NodeType_ExprList }
func (ExprTable) Type() int        { return NodeType_ExprMap }

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
func (ExprList) Expr()    {}
func (ExprTable) Expr()   {}

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

type ExprBlock struct {
	Value Expr
}

func (e ExprBlock) DataType() shared.DataType {
	return e.Value.DataType()
}

type ExprNumber struct {
	Value string
}

func (e ExprNumber) DataType() shared.DataType {
	return shared.Number{}
}

type ExprString struct {
	Value string
}

func (e ExprString) DataType() shared.DataType {
	return shared.String{}
}

type ExprBoolean struct {
	Value string
}

func (e ExprBoolean) DataType() shared.DataType {
	return shared.Boolean{}
}

type ExprList struct {
	Values []Expr
}

func (e ExprList) DataType() shared.DataType {
	if len(e.Values) == 0 {
		return shared.List{}
	}

	return shared.List{
		ListType: e.Values[0].DataType(),
	}
}

type KeyValuePair struct {
	Key   Expr
	Value Expr
}

type ExprTable struct {
	Pairs []KeyValuePair
}

func (e ExprTable) DataType() shared.DataType {
	if len(e.Pairs) == 0 {
		return shared.Map{}
	}

	return shared.Map{
		KeyType:   e.Pairs[0].Key.DataType(),
		ValueType: e.Pairs[0].Value.DataType(),
	}
}
