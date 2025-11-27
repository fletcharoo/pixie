package parser

import (
	"errors"
	"fmt"
	"io"
	"pixie/lexer"
)

func New(lexer *lexer.Lexer) *Parser {
	return &Parser{
		lexer: lexer,
	}
}

type Parser struct {
	lexer *lexer.Lexer
}

func (p *Parser) Parse() (node Node, err error) {
	return p.parseBlock()
}

func (p *Parser) parseBlock() (block StmtBlock, err error) {
	var stmts []Stmt

	for {
		// Check the next token and see if it's EOF.
		_, err = p.lexer.PeekToken()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			err = fmt.Errorf("failed to peek token: %w", err)
			return
		}

		// Parse the next statement.
		var stmt Stmt
		stmt, err = p.parseStmt()
		if err != nil {
			err = fmt.Errorf("failed to parse statement: %w", err)
			return
		}

		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}

	return StmtBlock{
		Stmts: stmts,
	}, nil
}

func (p *Parser) parseStmt() (stmt Stmt, err error) {
	tok, err := p.lexer.PeekToken()
	if err != nil {
		err = fmt.Errorf("failed to peek token: %w", err)
		return
	}

	switch tok.Type {
	case lexer.TokenType_Label:
		stmt, err = p.parseStmtLabel()
		if err != nil {
			err = fmt.Errorf("failed to parse label: %w", err)
			return
		}
		return stmt, nil
	}

	return nil, nil
}

func (p *Parser) parseStmtLabel() (stmt Stmt, err error) {
	tokLabel, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get label token: %w", err)
		return
	}

	tokNext, err := p.lexer.PeekToken()
	if err != nil {
		err = fmt.Errorf("failed to peek token: %w", err)
		return
	}

	switch tokNext.Type {
	case lexer.TokenType_OpenParan:
		stmt, err = p.parseStmtCallFunction(tokLabel)
		if err != nil {
			err = fmt.Errorf("failed to parse statement call function: %w", err)
			return
		}
		return stmt, nil
	case lexer.TokenType_Colon:
		stmt, err = p.parseStmtVarDeclare(tokLabel)
		if err != nil {
			err = fmt.Errorf("failed to parse statement variable declare: %w", err)
			return
		}
		return stmt, nil
	case lexer.TokenType_Equal:
		stmt, err = p.parseStmtVarAssign(tokLabel)
		if err != nil {
			err = fmt.Errorf("failed to parse statement variable assign: %w", err)
			return
		}
		return stmt, nil
	default:
		err = fmt.Errorf("expected label statement, got %q %q", tokLabel.String(), tokNext.String())
		return
	}
}

func (p *Parser) parseStmtCallFunction(tokLabel lexer.Token) (stmt StmtCallFunction, err error) {
	// Consume the open paran token.
	if _, err = p.lexer.GetToken(); err != nil {
		err = fmt.Errorf("failed to get open paran token: %w", err)
		return
	}

	exprs := make([]Expr, 0)
	var expr Expr
	var tokNext lexer.Token
parseStmtCallFunctionLoop:
	for {
		expr, err = p.parseExpr()
		if err != nil {
			err = fmt.Errorf("failed to parse expression: %w", err)
			return
		}

		exprs = append(exprs, expr)

		tokNext, err = p.lexer.PeekToken()
		if err != nil {
			err = fmt.Errorf("failed to peek token: %w", err)
			return
		}

		switch tokNext.Type {
		case lexer.TokenType_CloseParan:
			_, err = p.lexer.GetToken()
			if err != nil {
				err = fmt.Errorf("failed to get close paran token: %w", err)
				return
			}
			break parseStmtCallFunctionLoop
		case lexer.TokenType_Comma:
			_, err = p.lexer.GetToken()
			if err != nil {
				err = fmt.Errorf("failed to get comma token: %w", err)
				return
			}
			continue
		default:
			err = fmt.Errorf("unexpected token %q", tokNext.String())
			return
		}
	}

	return StmtCallFunction{
		FunctionName: tokLabel.Value,
		Args:         exprs,
	}, nil
}

func (p *Parser) parseStmtVarDeclare(tokLabel lexer.Token) (stmt StmtVarDeclare, err error) {
	if len(tokLabel.Value) == 0 {
		err = fmt.Errorf("variable name is empty")
		return
	}

	// Consume colon token
	if _, err = p.lexer.GetToken(); err != nil {
		err = fmt.Errorf("failed to consume colon token: %w", err)
		return
	}

	// Get the type of variable being declared
	tokType, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get type token: %w", err)
		return
	}

	if tokType.Type != lexer.TokenType_Label {
		err = fmt.Errorf("expected label got: %v", tokType)
		return
	}

	if len(tokType.Value) == 0 {
		err = fmt.Errorf("type is empty")
		return
	}

	// Check to see if there's an assignment
	var expr Expr
	tokEqual, err := p.lexer.PeekToken()
	if err != nil {
		err = fmt.Errorf("failed to peek equal token: %w", err)
		return
	}

	if tokEqual.Type == lexer.TokenType_Equal {
		// Consume equal token
		if _, err = p.lexer.GetToken(); err != nil {
			err = fmt.Errorf("failed to consume equal token: %w", err)
			return
		}

		expr, err = p.parseExpr()
		if err != nil {
			err = fmt.Errorf("failed to parse expression: %w", err)
			return
		}
	}

	return StmtVarDeclare{
		VariableName: tokLabel.Value,
		DataType:     tokType.Value,
		Expr:         expr,
	}, nil
}

func (p *Parser) parseStmtVarAssign(tokLabel lexer.Token) (stmt StmtVarAssign, err error) {
	if len(tokLabel.Value) == 0 {
		err = fmt.Errorf("variable name is empty")
		return
	}

	// Consume the equal token
	if _, err = p.lexer.GetToken(); err != nil {
		err = fmt.Errorf("failed to consume equal token: %w", err)
		return
	}

	var expr Expr
	expr, err = p.parseExpr()
	if err != nil {
		err = fmt.Errorf("failed to parse expression: %w", err)
		return
	}

	return StmtVarAssign{
		VariableName: tokLabel.Value,
		Expr:         expr,
	}, nil
}

func (p *Parser) parseExpr() (expr Expr, err error) {
	tok, err := p.lexer.PeekToken()
	if err != nil {
		err = fmt.Errorf("failed to peek token: %w", err)
		return
	}

	switch tok.Type {
	case lexer.TokenType_NumberLiteral:
		expr, err = p.parseExprNumberLiteral()
		if err != nil {
			err = fmt.Errorf("failed to parse number literal: %w", err)
			return
		}
		return expr, nil
	case lexer.TokenType_StringLiteral:
		expr, err = p.parseExprStringLiteral()
		if err != nil {
			err = fmt.Errorf("failed to parse string literal: %w", err)
			return
		}
		return expr, nil
	case lexer.TokenType_BooleanLiteral:
		expr, err = p.parseExprBooleanLiteral()
		if err != nil {
			err = fmt.Errorf("failed to parse boolean literal: %w", err)
			return
		}
		return expr, nil
	}

	err = fmt.Errorf("expected expression, got %q", tok.String())
	return
}

func (p *Parser) parseExprNumberLiteral() (expr ExprNumber, err error) {
	tok, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return
	}
	if len(tok.Value) == 0 {
		err = fmt.Errorf("token value is empty")
		return
	}

	expr.Value = tok.Value
	return expr, nil
}

func (p *Parser) parseExprStringLiteral() (expr ExprString, err error) {
	tok, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return
	}

	return ExprString{Value: tok.Value}, nil
}

func (p *Parser) parseExprBooleanLiteral() (expr ExprBoolean, err error) {
	tok, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get token: %w", err)
		return
	}

	if tok.Value != "true" && tok.Value != "false" {
		err = fmt.Errorf("expected boolean, got: %v", tok.String())
		return
	}

	return ExprBoolean{
		Value: tok.Value,
	}, nil
}
