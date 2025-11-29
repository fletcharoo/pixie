// Package lexer provides tokenization functionality for the pixie programming language.
// It converts input strings into a sequence of tokens that can be processed by a parser.
package lexer

import (
	"errors"
	"fmt"
	"io"
	"pixie/shared"
	"unicode"
)

// TokenType represents the different types of tokens that can be recognized by the lexer.
const (
	TokenType_Undefined      = iota // TokenType_Undefined represents an undefined or unrecognized token type
	TokenType_Label                 // TokenType_Label represents an identifier or variable name token
	TokenType_NumberLiteral         // TokenType_NumberLiteral represents a numeric value token
	TokenType_StringLiteral         // TokenType_StringLiteral represents a string value token
	TokenType_BooleanLiteral        // TokenType_BooleanLiteral represents a boolean value token (true/false)
	TokenType_Plus                  // TokenType_Plus represents a + character
	TokenType_Minus                 // TokenType_Minus represents a - character
	TokenType_Asterisk              // TokenType_Asterisk represents a * character
	TokenType_ForwardSlash          // TokenType_ForwardSlash represents a / character
	TokenType_OpenParan             // TokenType_OpenParan represents a ( character
	TokenType_CloseParan            // TokenType_CloseParan represents a ) character
	TokenType_Comma                 // TokenType_Comma represents a , character
	TokenType_Equal                 // TokenType_Equal represents a = character
	TokenType_Colon                 // TokenType_Colon represents a : character
	TokenType_OpenBracket           // TokenType_OpenBracket representes a [ character
	TokenType_CloseBracket          // TokenType_CloseBracet represents a ] character
	TokenType_Period                // TokenType_Period represents a . character
	TokenType_OpenBrace             // TokenType_OpenBrace represents a { character
	TokenType_CloseBrace            // TokenType_CloseBrace represents a } character
)

// TokenTypeString maps token type constants to their string representations for debugging and display purposes.
var (
	TokenTypeString map[int]string = map[int]string{
		TokenType_Undefined:      "Undefined",
		TokenType_Label:          "Label",
		TokenType_NumberLiteral:  "NumberLiteral",
		TokenType_StringLiteral:  "StringLiteral",
		TokenType_BooleanLiteral: "BooleanLiteral",
		TokenType_Plus:           "Plus",
		TokenType_Minus:          "Minus",
		TokenType_Asterisk:       "Asterisk",
		TokenType_ForwardSlash:   "ForwardSlash",
		TokenType_OpenParan:      "OpenParan",
		TokenType_CloseParan:     "CloseParan",
		TokenType_Comma:          "Comma",
		TokenType_Equal:          "Equal",
		TokenType_Colon:          "Colon",
		TokenType_OpenBracket:    "OpenBracket",
		TokenType_CloseBracket:   "CloseBracket",
		TokenType_Period:         "Period",
		TokenType_OpenBrace:      "OpenBrace",
		TokenType_CloseBrace:     "CloseBrace",
	}

	TokenTypeCharactersMap map[rune]Token = map[rune]Token{
		'(': {Type: TokenType_OpenParan},
		')': {Type: TokenType_CloseParan},
		',': {Type: TokenType_Comma},
		'+': {Type: TokenType_Plus},
		'-': {Type: TokenType_Minus},
		'*': {Type: TokenType_Asterisk},
		'/': {Type: TokenType_ForwardSlash},
		'=': {Type: TokenType_Equal},
		':': {Type: TokenType_Colon},
		'[': {Type: TokenType_OpenBracket},
		']': {Type: TokenType_CloseBracket},
		'{': {Type: TokenType_OpenBrace},
		'}': {Type: TokenType_CloseBrace},
	}
)

var (
	errUnexpectedEOF = errors.New("unexpected EOF") // errUnexpectedEOF is returned when a string literal is not properly closed
	errInvalidRune   = errors.New("invalid rune")   // errInvalidRune is returned when an invalid character is encountered
)

// isLabelRune returns whether the provided rune is a valid label rune.
// Valid label runes are letters, numbers, and underscores.
func isLabelRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_'
}

// Token represents a single token from the input string with its type and value.
type Token struct {
	Type  int    // The type of the token as defined by the TokenType constants
	Value string // The actual text value of the token from the input string
}

// String makes Token implement the Stringer interface.
func (t Token) String() string {
	s, ok := TokenTypeString[t.Type]

	if !ok {
		return TokenTypeString[TokenType_Undefined]
	}

	return s
}

// New creates and returns a new Lexer configured to read from the provided input string.
// The lexer will tokenize the input string according to the rules defined in the lexer package.
func New(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
	}
}

// Lexer provides functionality to tokenize an input string into a sequence of tokens.
// It supports peeking at the next token without consuming it and handles various token types
// including numbers, strings, labels, and boolean literals.
type Lexer struct {
	input []rune // The input string converted to runes for proper Unicode handling
	index int    // Current position in the input
	buf   *Token // Buffered token for peeking functionality
}

// getRune returns the rune at the current index of the input and increments the index.
// If the index is beyond the length of the input, getRune returns an io.EOF error.
func (l *Lexer) getRune() (r rune, err error) {
	if l.index >= len(l.input) {
		return 0, io.EOF
	}

	r = l.input[l.index]
	l.index++
	return r, nil
}

// peekRune returns the rune at the current index of the input without incrementing the index.
// If the index is beyond the length of the input, peekRune returns an io.EOF error.
func (l *Lexer) peekRune() (r rune, err error) {
	if l.index >= len(l.input) {
		return 0, io.EOF
	}

	return l.input[l.index], nil
}

func (l *Lexer) ConsumeToken(expected int) (err error) {
	tok, err := l.GetToken()
	if err != nil {
		return err
	}
	if tok.Type != expected {
		err = fmt.Errorf("unexpected token, wanted %q got %q", TokenTypeString[expected], tok.String())
	}
	return nil
}

// GetToken returns the next token in the input string.
// If there are no more tokens to process in the string, GetToken returns an
// io.EOF error. The lexer recognizes the following token types:
// - Numbers (integers and floats)
// - Strings (enclosed in double quotes)
// - Labels (identifiers starting with letters, numbers, or underscores)
// - Boolean literals (true and false)
func (l *Lexer) GetToken() (tok Token, err error) {
	if l.buf != nil {
		tok = *l.buf
		l.buf = nil
		return tok, nil
	}

	var r rune
	var ok bool

	for {
		r, err = l.peekRune()

		if err != nil {
			return tok, err
		}

		if unicode.IsSpace(r) {
			l.index++
			continue
		}

		if unicode.IsNumber(r) {
			return l.getTokenNumberLiteral()
		}

		if isLabelRune(r) {
			return l.getTokenLabel()
		}

		switch r {
		case '"':
			l.index++
			return l.getTokenStringLiteral()
		case '.':
			// Handle period as a token
			// Note: This means that decimals starting with a period (like ".5")
			// will be tokenized as [Period, Number] which is the expected behavior
			// based on the test cases
			l.index++
			return Token{Type: TokenType_Period}, nil
		case '/':
			// Handle comments first: check if next character is also / to form a comment
			nextIndex := l.index + 1
			if nextIndex >= len(l.input) {
				// No next character, so just return the forward slash as a token
				l.index++
				return Token{Type: TokenType_ForwardSlash}, nil
			}
			if l.input[nextIndex] == '/' {
				// This is a comment (//), skip both characters and the comment content
				l.index += 2 // skip both '/'
				l.skipComments()
				continue
			} else {
				// This is a division operator, not a comment
				l.index++
				return Token{Type: TokenType_ForwardSlash}, nil
			}
		}

		tok, ok = TokenTypeCharactersMap[r]
		if ok {
			l.index++
			return tok, nil
		}

		// If the rune is unknown.
		err = fmt.Errorf("%w: %s", errInvalidRune, string(r))
		return
	}
}

// PeekToken returns the current token in the input string without advancing the lexer position.
// This allows looking ahead at the next token without consuming it.
// If there are no more tokens to process in the string, PeekToken returns an
// io.EOF error.
func (l *Lexer) PeekToken() (tok Token, err error) {
	if l.buf != nil {
		tok = *l.buf
		return tok, nil
	}

	tok, err = l.GetToken()
	if err != nil {
		return tok, err
	}
	l.buf = &tok
	return tok, err
}

// getTokenNumberLiteral scans and returns a number literal token from the current position.
// It handles both integer and floating-point numbers (with decimal points).
// If there are no more characters to process in the string, getTokenNumberLiteral returns an
// io.EOF error.
func (l *Lexer) getTokenNumberLiteral() (tok Token, err error) {
	tok.Type = TokenType_NumberLiteral
	var r rune
	hasDecimalPoint := false

	for {
		r, err = l.peekRune()

		if err == io.EOF {
			return tok, nil
		}

		if err != nil {
			return tok, err
		}

		if unicode.IsNumber(r) {
			l.index++
			tok.Value += string(r)
			continue
		}

		if r == '.' && !hasDecimalPoint {
			l.index++
			tok.Value += string(r)
			hasDecimalPoint = true
			continue
		}

		return tok, nil
	}
}

// getTokenLabel scans and returns a label token from the current position.
// A label is defined as a sequence of letters, numbers, and underscores.
// If the label matches "true" or "false", it is returned as a boolean literal token instead.
// If there are no more characters to process in the string, getTokenLabel returns an
// io.EOF error.
func (l *Lexer) getTokenLabel() (tok Token, err error) {
	tok.Type = TokenType_Label
	var r rune

	for {
		r, err = l.peekRune()

		if err != nil {
			break
		}

		if isLabelRune(r) {
			l.index++
			tok.Value += string(r)
			continue
		}

		break
	}

	// Check if this is a boolean literal
	if tok.Value == shared.Keyword_True || tok.Value == shared.Keyword_False {
		tok.Type = TokenType_BooleanLiteral
	}

	if err == io.EOF {
		return tok, nil
	}

	return tok, err
}

// getTokenStringLiteral scans and returns a string literal token from the current position.
// It reads characters until it encounters a closing double quote.
// If the end of input is reached before finding a closing quote,
// getTokenStringLiteral returns an errUnexpectedEOF error.
func (l *Lexer) getTokenStringLiteral() (tok Token, err error) {
	tok.Type = TokenType_StringLiteral
	var r rune

	for {
		r, err = l.getRune()

		if err == io.EOF {
			err = errUnexpectedEOF
			return
		}

		if err != nil {
			return
		}

		if r == '"' {
			break
		}

		tok.Value += string(r)
	}

	return tok, nil
}

func (l *Lexer) skipComments() error {
	for {
		r, err := l.peekRune()
		if err != nil {
			return err
		}

		if r == '\n' || r == '\r' {
			break
		}
		l.index++
	}

	return nil
}
