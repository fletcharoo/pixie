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
		if tokNext.Value == shared.Keyword_Object {
			stmt, err = p.parseStmtObjDefine(tokLabel)
			if err != nil {
				err = fmt.Errorf("failed to parse statement object define: %w", err)
				return
			}
			return stmt, err
		}

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

func (p *Parser) parseStmtObjDefine(tokLabel lexer.Token) (stmt StmtObjDefine, err error) {
	if len(tokLabel.Value) == 0 {
		err = fmt.Errorf("object name is empty")
		return
	}

	if _, ok := shared.IllegalKeywords[tokLabel.Value]; ok {
		err = fmt.Errorf("variable name %q is illegal", tokLabel.Value)
		return
	}

	// Consume the obj token
	tokObj, err := p.lexer.GetToken()
	if err != nil {
		err = fmt.Errorf("failed to consume the obj token: %w", err)
		return
	}

	if tokObj.Value != shared.Keyword_Object {
		err = fmt.Errorf("expected \"obj\" got %q", tokObj.Value)
		return
	}

	// Consume open brace
	if err = p.lexer.ConsumeToken(lexer.TokenType_OpenBrace); err != nil {
		err = fmt.Errorf("failed to consume open brace: %w", err)
		return
	}
	fields := make([]FieldTypePair, 0)
	var tokNext lexer.Token
parseStmtObjDefine:
	for {
		// Parse field name label
		var tokFieldName lexer.Token
		tokFieldName, err = p.lexer.GetToken()
		if err != nil {
			err = fmt.Errorf("failed to get field name token: %w", err)
			return
		}
		if tokFieldName.Type != lexer.TokenType_Label {
			err = fmt.Errorf("expected label, got %q", tokFieldName.String())
			return
		}

		// Parse field type
		var fieldType shared.DataType
		fieldType, err = p.parseDataType()
		if err != nil {
			err = fmt.Errorf("failed to parse field type: %w", err)
			return
		}

		// Append to pairs
		fields = append(fields, FieldTypePair{
			Field: tokFieldName.Value,
			Type:  fieldType,
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
			break parseStmtObjDefine
		case lexer.TokenType_Label:
			continue
		default:
			err = fmt.Errorf("unexpected token %q", tokNext.String())
			return
		}
	}

	return StmtObjDefine{
		Name:   tokLabel.Value,
		Fields: fields,
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

	// Check if it's a built-in data type.
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

	// If it's not built-in it must be custom
	return shared.Custom{Name: tokLabel.Value}, nil
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
		err = fmt.Errorf("failed to consume open bracket: %w", err)
		return
	}

	// Parse key data type
	keyDataType, err := p.parseDataType()
	if err != nil {
		err = fmt.Errorf("failed to parse key data type: %w", err)
		return
	}

	// Consume colon separator
	if err = p.lexer.ConsumeToken(lexer.TokenType_Colon); err != nil {
		err = fmt.Errorf("failed to consume colon separator: %w", err)
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
		err = fmt.Errorf("failed to consume close bracket: %w", err)
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
	// Parse binary expression with precedence
	expr, err = p.parseExprWithPrecedence(0)
	if err != nil {
		return expr, err
	}

	// Then handle any indexing operations
	for {
		tok, err := p.lexer.PeekToken()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return expr, nil
			}
			return expr, fmt.Errorf("failed to peek token: %w", err)
		}

		switch tok.Type {
		case lexer.TokenType_OpenBracket:
			// Handle bracket indexing [index]
			_, err = p.lexer.GetToken() // consume '['
			if err != nil {
				return expr, fmt.Errorf("failed to consume open bracket token: %w", err)
			}

			indexExpr, err := p.parseExpr()
			if err != nil {
				return expr, fmt.Errorf("failed to parse index expression: %w", err)
			}

			tokCloseBracket, err := p.lexer.PeekToken()
			if err != nil {
				return expr, fmt.Errorf("failed to peek close bracket token: %w", err)
			}

			if tokCloseBracket.Type != lexer.TokenType_CloseBracket {
				return expr, fmt.Errorf("expected ']', got %q", tokCloseBracket.String())
			}

			_, err = p.lexer.GetToken() // consume ']'
			if err != nil {
				return expr, fmt.Errorf("failed to consume close bracket token: %w", err)
			}

			expr = ExprIndex{
				Left:  expr,
				Index: indexExpr,
			}
		case lexer.TokenType_Period:
			// Handle property access .property
			_, err = p.lexer.GetToken() // consume '.'
			if err != nil {
				return expr, fmt.Errorf("failed to consume period token: %w", err)
			}

			tokLabel, err := p.lexer.GetToken()
			if err != nil {
				return expr, fmt.Errorf("failed to get property label token: %w", err)
			}

			if tokLabel.Type != lexer.TokenType_Label {
				return expr, fmt.Errorf("expected label after '.', got %q", tokLabel.String())
			}

			expr = ExprPropertyAccess{
				Left:     expr,
				Property: tokLabel.Value,
			}
		default:
			return expr, nil
		}
	}
}

// Operator precedence levels
const (
	precedenceLowest = iota
	precedenceComparison // ==, !=, <, <=, >, >=
	precedenceSum        // +, -
	precedenceProduct    // *, /
)

func (p *Parser) parseExprWithPrecedence(precedence int) (expr Expr, err error) {
	expr, err = p.parseExprBase()
	if err != nil {
		return
	}

	for {
		tok, err := p.lexer.PeekToken()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return expr, nil
			}
			return expr, fmt.Errorf("failed to peek token: %w", err)
		}

		nextPrecedence := p.getPrecedence(tok.Type)
		if nextPrecedence <= precedence {
			break
		}

		operator := tok.Type
		_, err = p.lexer.GetToken() // consume the operator
		if err != nil {
			return expr, fmt.Errorf("failed to consume operator token: %w", err)
		}

		right, err := p.parseExprWithPrecedence(nextPrecedence)
		if err != nil {
			return expr, fmt.Errorf("failed to parse right side of operator: %w", err)
		}

		expr = ExprBinary{
			Left:     expr,
			Operator: operator,
			Right:    right,
		}
	}

	return expr, nil
}

func (p *Parser) getPrecedence(tokenType int) int {
	switch tokenType {
	case lexer.TokenType_EqualEqual, lexer.TokenType_BangEqual, lexer.TokenType_LessThan,
		 lexer.TokenType_LessThanEqual, lexer.TokenType_GreaterThan, lexer.TokenType_GreaterThanEqual:
		return precedenceComparison
	case lexer.TokenType_Plus, lexer.TokenType_Minus:
		return precedenceSum
	case lexer.TokenType_Asterisk, lexer.TokenType_ForwardSlash:
		return precedenceProduct
	default:
		return precedenceLowest
	}
}

func (p *Parser) parseExprBase() (expr Expr, err error) {
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
	case lexer.TokenType_OpenParan: // Handle parentheses for grouping
		_, err = p.lexer.GetToken() // consume the opening parenthesis
		if err != nil {
			err = fmt.Errorf("failed to consume open parenthesis: %w", err)
			return
		}

		// Parse the expression inside the parentheses
		expr, err = p.parseExpr()
		if err != nil {
			err = fmt.Errorf("failed to parse expression in parentheses: %w", err)
			return
		}

		// Consume the closing parenthesis
		if err = p.lexer.ConsumeToken(lexer.TokenType_CloseParan); err != nil {
			err = fmt.Errorf("failed to consume close parenthesis: %w", err)
			return
		}

		// Return as an expression block
		return ExprBlock{Value: expr}, nil
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

	pairs := make([]TablePair, 0)

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
		pairs = append(pairs, TablePair{
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
