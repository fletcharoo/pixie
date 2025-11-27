package compiler

import (
	"fmt"
	"pixie/parser"
	"strings"
)

func Compile(node parser.Node) (lua string, err error) {
	stmt, ok := node.(parser.Stmt)
	if !ok {
		err = fmt.Errorf("expected statement, got: %v", node)
		return
	}

	var sb strings.Builder
	c := &compiler{
		sb: &sb,
	}
	c.compileStmt(stmt)
	return sb.String(), nil
}

type compiler struct {
	sb *strings.Builder
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
	for _, s := range stmt.Stmts {
		err = c.compileStmt(s)
		if err != nil {
			err = fmt.Errorf("failed to compile stmt: %w", err)
			return
		}
		c.sb.WriteRune('\n')
	}
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
