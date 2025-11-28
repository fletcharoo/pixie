package compiler

import (
	"errors"
	"fmt"
	"pixie/parser"
	"pixie/shared"
	"strings"
)

const (
	globalScope = 1
)

var (
	ErrInvalidTypeAssign = fmt.Errorf("invalid type assign")
)

func Compile(node parser.Node) (lua string, err error) {
	stmt, ok := node.(parser.Stmt)
	if !ok {
		err = fmt.Errorf("expected statement, got: %v", node)
		return
	}

	var sb strings.Builder
	c := &compiler{
		sb:        &sb,
		variables: make(map[string]variable, 0),
	}
	if err = c.compileStmt(stmt); err != nil {
		err = fmt.Errorf("failed to compile statement: %w", err)
		return
	}
	return sb.String(), nil
}

type compiler struct {
	sb        *strings.Builder
	scope     int
	variables map[string]variable
}

type variable struct {
	scope    int
	dataType shared.DataType
}

func (c *compiler) compileStmt(stmt parser.Stmt) (err error) {
	switch n := stmt.(type) {
	case parser.StmtBlock:
		if err = c.compileStmtBlock(n); err != nil {
			err = fmt.Errorf("failed to compile statement block: %w", err)
			return
		}
	case parser.StmtCallFunction:
		if err = c.compileStmtCallFunction(n); err != nil {
			err = fmt.Errorf("failed to compile statement call function: %w", err)
			return
		}
	case parser.StmtVarDeclare:
		if err = c.compileStmtVarDeclare(n); err != nil {
			err = fmt.Errorf("failed to compile statement variable declare: %w", err)
			return
		}
	case parser.StmtVarAssign:
		if err = c.compileStmtVarAssign(n); err != nil {
			err = fmt.Errorf("failed to compile statement assign: %w", err)
			return
		}
	default:
		err = fmt.Errorf("expected statement, got: %v", n)
		return
	}

	return nil
}

func (c *compiler) compileExpr(expr parser.Expr) (err error) {
	switch n := expr.(type) {
	case parser.ExprBlock:
		if err = c.compileExprBlock(n); err != nil {
			err = fmt.Errorf("failed to compile expression block: %w", err)
			return
		}
	case parser.ExprNumber:
		if err = c.compileExprNumber(n); err != nil {
			err = fmt.Errorf("failed to compile expression number: %w", err)
			return
		}
	case parser.ExprString:
		if err = c.compileExprString(n); err != nil {
			err = fmt.Errorf("failed to compile expression string: %w", err)
			return
		}
	case parser.ExprBoolean:
		if err = c.compileExprBoolean(n); err != nil {
			err = fmt.Errorf("failed to compile expression boolean: %w", err)
			return
		}
	case parser.ExprList:
		if err = c.compileExprList(n); err != nil {
			err = fmt.Errorf("failed to compile expression list: %w", err)
			return
		}
	case parser.ExprTable:
		if err = c.compileExprTable(n); err != nil {
			err = fmt.Errorf("failed to compile expression table: %w", err)
			return
		}
	default:
		err = fmt.Errorf("expected expr, got: %v", n)
		return
	}

	return nil
}

func (c *compiler) compileStmtBlock(stmt parser.StmtBlock) (err error) {
	c.scope += 1
	for _, s := range stmt.Stmts {
		err = c.compileStmt(s)
		if err != nil {
			err = fmt.Errorf("failed to compile stmt: %w", err)
			return
		}
		c.sb.WriteRune('\n')
	}

	variablesToRemove := make([]string, 0, len(c.variables))
	for k, v := range c.variables {
		if v.scope == c.scope {
			variablesToRemove = append(variablesToRemove, k)
		}
	}

	for _, name := range variablesToRemove {
		delete(c.variables, name)
	}

	c.scope -= 1
	return nil
}

func (c *compiler) compileCommaSeparatedExpressions(exprs []parser.Expr) (err error) {
	argsLen := len(exprs)
	for i, arg := range exprs {
		if err = c.compileExpr(arg); err != nil {
			err = fmt.Errorf("failed to compile argument %d: %w", i, err)
			return
		}

		if i < argsLen-1 {
			c.sb.WriteRune(',')
		}
	}
	return nil
}

func (c *compiler) compileStmtCallFunction(stmt parser.StmtCallFunction) (err error) {
	c.sb.WriteString(stmt.FunctionName)
	c.sb.WriteRune('(')
	if err = c.compileCommaSeparatedExpressions(stmt.Args); err != nil {
		err = fmt.Errorf("failed to compile comma separated expressions: %w", err)
		return
	}
	c.sb.WriteRune(')')
	return nil
}

func (c *compiler) compileStmtVarDeclare(stmt parser.StmtVarDeclare) (err error) {
	_, ok := c.variables[stmt.VariableName]
	if ok {
		err = fmt.Errorf("variable %q already exists", stmt.VariableName)
		return
	}

	variable := variable{
		scope:    c.scope,
		dataType: stmt.DataType,
	}
	c.variables[stmt.VariableName] = variable

	// Check if we need to declare a local variable
	if c.scope != globalScope {
		c.sb.WriteString(shared.Keyword_Local)
		c.sb.WriteRune(' ')
	}

	// Write the variable name
	c.sb.WriteString(stmt.VariableName)
	c.sb.WriteString(" = ")

	// Write the expression
	if stmt.Expr == nil {
		c.sb.WriteString(variable.dataType.ZeroValue())
	} else {
		if err = c.compileExpr(stmt.Expr); err != nil {
			err = fmt.Errorf("failed to parse expression: %w", err)
			return
		}
	}

	return nil
}

func (c *compiler) compileStmtVarAssign(stmt parser.StmtVarAssign) (err error) {
	v, ok := c.variables[stmt.VariableName]
	if !ok {
		err = fmt.Errorf("variable %q does not exist", stmt.VariableName)
		return
	}

	if err = c.checkExpressionValidDataType(v.dataType, stmt.Expr); err != nil {
		err = errors.Join(ErrInvalidTypeAssign, fmt.Errorf("%s", err.Error())) // for some reason it wouldn't show the second error when I joined it with the err variable
		return
	}

	c.sb.WriteString(stmt.VariableName)
	c.sb.WriteString(" = ")
	if err = c.compileExpr(stmt.Expr); err != nil {
		err = fmt.Errorf("failed to parse expression: %w", err)
		return
	}

	return nil
}

func (c *compiler) compileExprBlock(expr parser.ExprBlock) (err error) {
	c.sb.WriteRune('(')
	if err = c.compileExpr(expr.Value); err != nil {
		err = fmt.Errorf("failed to compile expression: %w", err)
		return
	}
	c.sb.WriteRune(')')
	return nil
}

func (c *compiler) compileExprNumber(expr parser.ExprNumber) (err error) {
	c.sb.WriteString(expr.Value)
	return nil
}

func (c *compiler) compileExprString(expr parser.ExprString) (err error) {
	c.sb.WriteRune('"')
	c.sb.WriteString(expr.Value)
	c.sb.WriteRune('"')
	return nil
}

func (c *compiler) compileExprBoolean(expr parser.ExprBoolean) (err error) {
	c.sb.WriteString(expr.Value)
	return nil
}

func (c *compiler) compileExprList(expr parser.ExprList) (err error) {
	c.sb.WriteRune('[')
	if err = c.compileCommaSeparatedExpressions(expr.Values); err != nil {
		err = fmt.Errorf("failed to compile comma separated expressions: %w", err)
		return
	}
	c.sb.WriteRune(']')
	return nil
}
func (c *compiler) compileExprTable(expr parser.ExprTable) (err error) {
	c.sb.WriteRune('{')
	argsLen := len(expr.Pairs)
	for i, pair := range expr.Pairs {
		if err = c.compileExpr(pair.Key); err != nil {
			err = fmt.Errorf("failed to compile key %d: %w", i, err)
			return
		}

		c.sb.WriteRune(':')

		if err = c.compileExpr(pair.Value); err != nil {
			err = fmt.Errorf("failed to compile value %d: %w", i, err)
			return
		}

		if i < argsLen-1 {
			c.sb.WriteRune(',')
		}
	}
	c.sb.WriteRune('}')
	return nil
}

func (c *compiler) checkExpressionValidDataType(dataType shared.DataType, expr parser.Expr) (err error) {
	switch d := dataType.(type) {
	case shared.Number:
		return c.checkExpressionValidNumber(d, expr)
	case shared.String:
		return c.checkExpressionValidString(d, expr)
	case shared.Boolean:
		return c.checkExpressionValidBoolean(d, expr)
	case shared.List:
		return c.checkExpressionValidList(d, expr)
	case shared.Map:
		return c.checkExpressionValidMap(d, expr)
	}

	err = fmt.Errorf("unknown type: %#v", dataType)
	return
}

func (c *compiler) checkExpressionValidNumber(dataType shared.Number, expr parser.Expr) (err error) {
	switch e := expr.(type) {
	case parser.ExprNumber:
		return nil
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}
}

func (c *compiler) checkExpressionValidString(dataType shared.String, expr parser.Expr) (err error) {
	switch e := expr.(type) {
	case parser.ExprString:
		return nil
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}
}

func (c *compiler) checkExpressionValidBoolean(dataType shared.Boolean, expr parser.Expr) (err error) {
	switch e := expr.(type) {
	case parser.ExprBoolean:
		return nil
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}
}

func (c *compiler) checkExpressionValidList(dataType shared.List, expr parser.Expr) (err error) {
	var exprList parser.ExprList
	switch e := expr.(type) {
	case parser.ExprList:
		exprList = e
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}

	for _, value := range exprList.Values {
		if err = c.checkExpressionValidDataType(dataType.ListType, value); err != nil {
			err = fmt.Errorf("failed to check if list type is valid data type: %w", err)
			return
		}
	}
	return nil
}

func (c *compiler) checkExpressionValidMap(dataType shared.Map, expr parser.Expr) (err error) {
	var exprTable parser.ExprTable
	switch e := expr.(type) {
	case parser.ExprTable:
		exprTable = e
	default:
		return fmt.Errorf("expected %s got %T", dataType.String(), e)
	}

	for _, pair := range exprTable.Pairs {
		if err = c.checkExpressionValidDataType(dataType.KeyType, pair.Key); err != nil {
			err = fmt.Errorf("failed to check if map key type is valid data type: %w", err)
			return
		}

		if err = c.checkExpressionValidDataType(dataType.ValueType, pair.Value); err != nil {
			err = fmt.Errorf("failed to check if map value type is valid data type: %w", err)
			return
		}
	}
	return nil
}
