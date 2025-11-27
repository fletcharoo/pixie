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
	compileStmt(&sb, stmt)
	return sb.String(), nil
}

func compileStmt(sb *strings.Builder, stmt parser.Stmt) (err error) {
	switch n := stmt.(type) {
	case parser.StmtBlock:
		err = compileStmtBlock(sb, n)
		if err != nil {
			err = fmt.Errorf("failed to compile statement block: %w", err)
			return
		}
	case parser.StmtCallFunction:
		err = compileStmtCallFunction(sb, n)
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

func compileExpr(sb *strings.Builder, expr parser.Expr) (err error) {
	switch n := expr.(type) {
	case parser.ExprBlock:
		err = compileExprBlock(sb, n)
		if err != nil {
			err = fmt.Errorf("failed to compile expression block: %w", err)
			return
		}
	case parser.ExprNumber:
		err = compileExprNumber(sb, n)
		if err != nil {
			err = fmt.Errorf("failed to compile expression number: %w", err)
			return
		}
	case parser.ExprString:
		err = compileExprString(sb, n)
		if err != nil {
			err = fmt.Errorf("failed to compile expression string: %w", err)
			return
		}
	case parser.ExprBoolean:
		err = compileExprBoolean(sb, n)
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

func compileStmtBlock(sb *strings.Builder, stmt parser.StmtBlock) (err error) {
	for _, s := range stmt.Stmts {
		err = compileStmt(sb, s)
		if err != nil {
			err = fmt.Errorf("failed to compile stmt: %w", err)
			return
		}
		sb.WriteRune('\n')
	}
	return nil
}

func compileStmtCallFunction(sb *strings.Builder, stmt parser.StmtCallFunction) (err error) {
	sb.WriteString(stmt.FunctionName)
	sb.WriteRune('(')
	argsLen := len(stmt.Args)

	for i, arg := range stmt.Args {
		if err = compileExpr(sb, arg); err != nil {
			err = fmt.Errorf("failed to compile argument %d: %w", i, err)
			return
		}

		if i < argsLen-1 {
			sb.WriteRune(',')
		}
	}

	sb.WriteRune(')')
	return nil
}

func compileExprBlock(sb *strings.Builder, expr parser.ExprBlock) (err error) {
	sb.WriteRune('(')
	if err = compileExpr(sb, expr.Value); err != nil {
		err = fmt.Errorf("failed to compile expression: %w", err)
		return
	}
	sb.WriteRune(')')
	return nil
}

func compileExprNumber(sb *strings.Builder, expr parser.ExprNumber) (err error) {
	sb.WriteString(expr.Value)
	return nil
}

func compileExprString(sb *strings.Builder, expr parser.ExprString) (err error) {
	sb.WriteRune('"')
	sb.WriteString(expr.Value)
	sb.WriteRune('"')
	return nil
}

func compileExprBoolean(sb *strings.Builder, expr parser.ExprBoolean) (err error) {
	sb.WriteString(expr.Value)
	return nil
}
