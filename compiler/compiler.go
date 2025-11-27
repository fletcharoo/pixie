package compiler

import (
	"errors"
	"fmt"
	"pixie/parser"
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
	dataType string
}

func (c *compiler) compileStmt(stmt parser.Stmt) (err error) {
	switch n := stmt.(type) {
	case parser.StmtBlock:
		err = c.compileStmtBlock(n)
		if err != nil {
			err = fmt.Errorf("failed to compile statement block: %w", err)
			return
		}
	case parser.StmtCallFunction:
		err = c.compileStmtCallFunction(n)
		if err != nil {
			err = fmt.Errorf("failed to compile statement call function: %w", err)
			return
		}
	case parser.StmtAssign:
		err = c.compileStmtAssign(n)
		if err != nil {
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
		err = c.compileExprBlock(n)
		if err != nil {
			err = fmt.Errorf("failed to compile expression block: %w", err)
			return
		}
	case parser.ExprNumber:
		err = c.compileExprNumber(n)
		if err != nil {
			err = fmt.Errorf("failed to compile expression number: %w", err)
			return
		}
	case parser.ExprString:
		err = c.compileExprString(n)
		if err != nil {
			err = fmt.Errorf("failed to compile expression string: %w", err)
			return
		}
	case parser.ExprBoolean:
		err = c.compileExprBoolean(n)
		if err != nil {
			err = fmt.Errorf("failed to compile expression boolean: %w", err)
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

func (c *compiler) compileStmtCallFunction(stmt parser.StmtCallFunction) (err error) {
	c.sb.WriteString(stmt.FunctionName)
	c.sb.WriteRune('(')
	argsLen := len(stmt.Args)

	for i, arg := range stmt.Args {
		if err = c.compileExpr(arg); err != nil {
			err = fmt.Errorf("failed to compile argument %d: %w", i, err)
			return
		}

		if i < argsLen-1 {
			c.sb.WriteRune(',')
		}
	}

	c.sb.WriteRune(')')
	return nil
}

func (c *compiler) compileStmtAssign(stmt parser.StmtAssign) (err error) {
	if len(stmt.VariableName) == 0 {
		err = fmt.Errorf("variable name is empty")
		return
	}

	dataType := stmt.Expr.DataType()
	v, ok := c.variables[stmt.VariableName]

	if ok {
		return c.compileStmtAssignExists(stmt, dataType, v)
	}

	return c.compileStmtAssignNotExists(stmt, dataType)
}

func (c *compiler) compileStmtAssignExists(stmt parser.StmtAssign, dataType string, v variable) (err error) {
	if dataType != v.dataType {
		err = errors.Join(ErrInvalidTypeAssign, fmt.Errorf("variable %q wants %q got %q", stmt.VariableName, v.dataType, dataType))
		return
	}

	c.sb.WriteString(stmt.VariableName)
	c.sb.WriteString(" = ")
	if err = c.compileExpr(stmt.Expr); err != nil {
		err = fmt.Errorf("failed to compile expression: %w", err)
		return
	}

	return nil
}

func (c *compiler) compileStmtAssignNotExists(stmt parser.StmtAssign, dataType string) (err error) {
	c.variables[stmt.VariableName] = variable{
		scope:    c.scope,
		dataType: dataType,
	}

	if c.scope != globalScope {
		c.sb.WriteString("local ")
	}

	c.sb.WriteString(stmt.VariableName)
	c.sb.WriteString(" = ")
	if err = c.compileExpr(stmt.Expr); err != nil {
		err = fmt.Errorf("failed to compile expression: %w", err)
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
