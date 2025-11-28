package parser

import (
	"errors"
	"fmt"
	"io"
	"pixie/lexer"
	"pixie/shared"
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
	case lexer.TokenType_Label:
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

	if _, ok := shared.IllegalKeywords[tokLabel.Value]; ok {
		err = fmt.Errorf("variable name %q is illegal", tokLabel.Value)
		return
	}

	// Parse the data type
	dataType, err := p.parseDataType()
	if err != nil {
		err = fmt.Errorf("failed to parse data type: %w", err)
		return
	}

	// Check to see if there's an assignment
	var expr Expr
	tokEqual, err := p.lexer.PeekToken()
	if err != nil {
		if errors.Is(err, io.EOF) {
			// No assignment, just declaration - this is valid
			return StmtVarDeclare{
				VariableName: tokLabel.Value,
				DataType:     dataType,
				Expr:         nil,
			}, nil
		}
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
		DataType:     dataType,
		Expr:         expr,
	}, nil
}

func (p *Parser) parseDataType() (dataType shared.DataType, err error) {
	tokLabel, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get label token: %w", err)
		return
	}

	if tokLabel.Type != lexer.TokenType_Label {
		err = fmt.Errorf("expected label, got %q", lexer.TokenTypeString[tokLabel.Type])
		return
	}

	if len(tokLabel.Value) == 0 {
		err = fmt.Errorf("data type is empty")
		return
	}

	switch tokLabel.Value {
	case shared.Keyword_Number:
		return shared.Number{}, nil
	case shared.Keyword_String:
		return shared.String{}, nil
	case shared.Keyword_Boolean:
		return shared.Boolean{}, nil
	case shared.Keyword_List:
		dataType, err = p.parseDataTypeList()
		if err != nil {
			err = fmt.Errorf("failed to parse data type list: %w", err)
			return
		}
		return dataType, nil
	case shared.Keyword_Map:
		dataType, err = p.parseDataTypeMap()
		if err != nil {
			err = fmt.Errorf("failed to parse data type map: %w", err)
			return
		}
		return dataType, nil
	}

	err = fmt.Errorf("unknown type %q", tokLabel.Value)
	return
}

func (p *Parser) parseDataTypeList() (dataType shared.List, err error) {
	// Consume open bracket
	tokOpenBracket, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get open bracket token: %w", err)
		return
	}

	if tokOpenBracket.Type != lexer.TokenType_OpenBracket {
		err = fmt.Errorf("expected open bracket, got %q", tokOpenBracket.String())
		return
	}

	// Parse sub data type
	subDataType, err := p.parseDataType()
	if err != nil {
		err = fmt.Errorf("failed to parse list sub data type: %w", err)
		return
	}

	// Consume close bracket
	tokCloseBracket, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get close bracket token: %w", err)
		return
	}

	if tokCloseBracket.Type != lexer.TokenType_CloseBracket {
		err = fmt.Errorf("expected close bracket, got %q", tokCloseBracket.String())
		return
	}

	return shared.List{
		ListType: subDataType,
	}, nil
}

func (p *Parser) parseDataTypeMap() (dataType shared.Map, err error) {
	// Consume open bracket
	if err = p.lexer.ConsumeToken(lexer.TokenType_OpenBracket); err != nil {
		err = fmt.Errorf("failed to consume key open bracket: %w", err)
		return
	}

	// Parse key data type
	keyDataType, err := p.parseDataType()
	if err != nil {
		err = fmt.Errorf("failed to parse key data type: %w", err)
		return
	}

	// Consume close bracket
	if err = p.lexer.ConsumeToken(lexer.TokenType_CloseBracket); err != nil {
		err = fmt.Errorf("failed to consume key close bracket: %w", err)
		return
	}

	// Consume open bracket
	if err = p.lexer.ConsumeToken(lexer.TokenType_OpenBracket); err != nil {
		err = fmt.Errorf("failed to consume value open bracket: %w", err)
		return
	}

	// Parse value data type
	valueDataType, err := p.parseDataType()
	if err != nil {
		err = fmt.Errorf("failed to parse value data type: %w", err)
		return
	}

	// Consume close bracket
	if err = p.lexer.ConsumeToken(lexer.TokenType_CloseBracket); err != nil {
		err = fmt.Errorf("failed to consume value close bracket: %w", err)
		return
	}

	return shared.Map{
		KeyType:   keyDataType,
		ValueType: valueDataType,
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
	case lexer.TokenType_OpenBracket:
		expr, err = p.parseExprList()
		if err != nil {
			err = fmt.Errorf("failed to parse list expression: %w", err)
			return
		}
		return expr, nil
	case lexer.TokenType_OpenBrace:
		expr, err = p.parseExprTable()
		if err != nil {
			err = fmt.Errorf("failed to parse table expression: %w", err)
			return
		}
		return expr, nil
	case lexer.TokenType_Label:
		expr, err = p.parseExprLabel()
		if err != nil {
			err = fmt.Errorf("failed to parse label expression: %w", err)
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

func (p *Parser) parseExprList() (expr ExprList, err error) {
	// Consume open bracket
	tokOpenBracket, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to consume open bracket: %w", err)
		return
	}
	if tokOpenBracket.Type != lexer.TokenType_OpenBracket {
		err = fmt.Errorf("expected open bracket, got %q", tokOpenBracket.String())
		return
	}

	// parse the inside expressions
	exprs := make([]Expr, 0)
	var listExpr Expr
	var tokNext lexer.Token
parseExprListLoop:
	for {
		listExpr, err = p.parseExpr()
		if err != nil {
			err = fmt.Errorf("failed to parse expression: %w", err)
			return
		}

		exprs = append(exprs, listExpr)

		tokNext, err = p.lexer.PeekToken()
		if err != nil {
			err = fmt.Errorf("failed to peek token: %w", err)
			return
		}

		switch tokNext.Type {
		case lexer.TokenType_CloseBracket:
			_, err = p.lexer.GetToken()
			if err != nil {
				err = fmt.Errorf("failed to get close paran token: %w", err)
				return
			}
			break parseExprListLoop
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

	return ExprList{
		Values: exprs,
	}, nil
}

func (p *Parser) parseExprTable() (expr ExprTable, err error) {
	// Consume open brace
	if err = p.lexer.ConsumeToken(lexer.TokenType_OpenBrace); err != nil {
		err = fmt.Errorf("failed to consume open brace token: %w", err)
		return
	}

	pairs := make([]KeyValuePair, 0)

	// Parse inside fo table
	var tokNext lexer.Token
parseExprMapLoop:
	for {
		// Parse key expression
		var keyExpr Expr
		keyExpr, err = p.parseExpr()
		if err != nil {
			err = fmt.Errorf("failed to parse key expression: %w", err)
			return
		}

		// Consume colon
		if err = p.lexer.ConsumeToken(lexer.TokenType_Colon); err != nil {
			err = fmt.Errorf("failed to consume colon: %w", err)
			return
		}

		// Parse value expression
		var valueExpr Expr
		valueExpr, err = p.parseExpr()
		if err != nil {
			err = fmt.Errorf("failed to parse value expression: %w", err)
			return
		}

		// Append to pairs
		pairs = append(pairs, KeyValuePair{
			Key:   keyExpr,
			Value: valueExpr,
		})

		tokNext, err = p.lexer.PeekToken()
		if err != nil {
			err = fmt.Errorf("failed to peek token: %w", err)
			return
		}

		switch tokNext.Type {
		case lexer.TokenType_CloseBrace:
			_, err = p.lexer.GetToken()
			if err != nil {
				err = fmt.Errorf("failed to get close brace token: %w", err)
				return
			}
			break parseExprMapLoop
		case lexer.TokenType_Comma:
			if err = p.lexer.ConsumeToken(lexer.TokenType_Comma); err != nil {
				err = fmt.Errorf("failed to consume comma token: %w", err)
				return
			}
			continue
		default:
			err = fmt.Errorf("unexpected token %q", tokNext.String())
			return
		}
	}
	return ExprTable{
		Pairs: pairs,
	}, nil
}

func (p *Parser) parseExprLabel() (expr Expr, err error) {
	tokLabel, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to get label token: %w", err)
		return
	}

	if len(tokLabel.Value) == 0 {
		err = fmt.Errorf("label is empty")
		return
	}

	return ExprVariable{
		Name: tokLabel.Value,
	}, nil
}
