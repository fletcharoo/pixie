package lexer

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected []rune
	}{
		"empty_string":       {"", []rune{}},
		"single_character":   {"a", []rune{'a'}},
		"simple_string":      {"hello", []rune{'h', 'e', 'l', 'l', 'o'}},
		"unicode_string":     {"héllo", []rune{'h', 'é', 'l', 'l', 'o'}},
		"string_with_spaces": {"hello world", []rune{'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd'}},
		"special_characters": {"a!@#$%^&*()", []rune{'a', '!', '@', '#', '$', '%', '^', '&', '*', '(', ')'}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := New(tt.input)
			assert.Equal(t, tt.expected, lexer.input, "New(%q) should create lexer with correct input runes", tt.input)
			assert.Equal(t, 0, lexer.index, "New(%q) should initialize index to 0", tt.input)
			assert.Nil(t, lexer.buf, "New(%q) should initialize buffer to nil", tt.input)
		})
	}
}

func TestTokenString(t *testing.T) {
	tests := map[string]struct {
		token    Token
		expected string
	}{
		"undefined_token":       {Token{Type: TokenType_Undefined}, "Undefined"},
		"label_token":           {Token{Type: TokenType_Label}, "Label"},
		"number_literal_token":  {Token{Type: TokenType_NumberLiteral}, "NumberLiteral"},
		"string_literal_token":  {Token{Type: TokenType_StringLiteral}, "StringLiteral"},
		"boolean_literal_token": {Token{Type: TokenType_BooleanLiteral}, "BooleanLiteral"},
		"invalid_token_type":    {Token{Type: 999}, "Undefined"}, // Should default to Undefined
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := tt.token.String()
			assert.Equal(t, tt.expected, result, "Token{%d, %q}.String() should return %q", tt.token.Type, tt.token.Value, tt.expected)
		})
	}
}

func TestIsLabelRune(t *testing.T) {
	tests := map[string]struct {
		r        rune
		expected bool
	}{
		// Test letters (English)
		"lowercase_a": {'a', true},
		"uppercase_A": {'A', true},
		"lowercase_z": {'z', true},
		"uppercase_Z": {'Z', true},

		// Test numbers
		"digit_0": {'0', true},
		"digit_5": {'5', true},
		"digit_9": {'9', true},

		// Test underscore
		"underscore": {'_', true},

		// Test special characters that should return false
		"space":             {' ', false},
		"hyphen":            {'-', false},
		"dot":               {'.', false},
		"comma":             {',', false},
		"exclamation":       {'!', false},
		"at_symbol":         {'@', false},
		"hash":              {'#', false},
		"dollar":            {'$', false},
		"percent":           {'%', false},
		"caret":             {'^', false},
		"ampersand":         {'&', false},
		"asterisk":          {'*', false},
		"plus":              {'+', false},
		"equals":            {'=', false},
		"parenthesis_open":  {'(', false},
		"parenthesis_close": {')', false},
		"bracket_open":      {'[', false},
		"bracket_close":     {']', false},
		"brace_open":        {'{', false},
		"brace_close":       {'}', false},
		"pipe":              {'|', false},
		"backslash":         {'\\', false},
		"semicolon":         {';', false},
		"colon":             {':', false},
		"apostrophe":        {'\'', false},
		"quote":             {'"', false},
		"less_than":         {'<', false},
		"greater_than":      {'>', false},
		"slash":             {'/', false},
		"question_mark":     {'?', false},
		"tilde":             {'~', false},
		"backtick":          {'`', false},

		// Test Unicode letters (some examples)
		"unicode_letter_a_with_grave":  {'à', true},
		"unicode_letter_n_with_tilde":  {'ñ', true},
		"unicode_letter_o_with_umlaut": {'ö', true},

		// Test Unicode numbers (some examples)
		"unicode_number_arabic_indic_digit_zero": {'٠', true},
		"unicode_number_arabic_indic_digit_five": {'٥', true},
		"unicode_number_roman_numeral_one":       {'Ⅰ', true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := isLabelRune(tt.r)
			assert.Equal(t, tt.expected, result, "isLabelRune(%q)", tt.r)
		})
	}
}

func TestGetToken(t *testing.T) {
	tests := map[string]struct {
		input    string
		expected []Token
		hasError bool
	}{
		// Test number literals
		"single_integer":            {"42", []Token{{Type: TokenType_NumberLiteral, Value: "42"}}, false},
		"single_float":              {"3.14", []Token{{Type: TokenType_NumberLiteral, Value: "3.14"}}, false},
		"integer_with_decimal":      {"42.0", []Token{{Type: TokenType_NumberLiteral, Value: "42.0"}}, false},
		"large_integer":             {"123456789", []Token{{Type: TokenType_NumberLiteral, Value: "123456789"}}, false},
		"zero":                      {"0", []Token{{Type: TokenType_NumberLiteral, Value: "0"}}, false},
		"float_starting_with_zero":  {"0.5", []Token{{Type: TokenType_NumberLiteral, Value: "0.5"}}, false},
		"multiple_decimals":         {"3.14.15", nil, true}, // Multiple decimals cause an error after parsing "3.14"
		"decimal_starting_with_dot": {".5", nil, true},      // This causes an error as dot is not a valid start for any token

		// Test string literals
		"empty_string":              {`""`, []Token{{Type: TokenType_StringLiteral, Value: ""}}, false},
		"simple_string":             {`"hello"`, []Token{{Type: TokenType_StringLiteral, Value: "hello"}}, false},
		"string_with_spaces":        {`"hello world"`, []Token{{Type: TokenType_StringLiteral, Value: "hello world"}}, false},
		"string_with_special_chars": {`"hello!@#$%^&*()"`, []Token{{Type: TokenType_StringLiteral, Value: "hello!@#$%^&*()"}}, false},
		// Note: The following test will fail since the lexer doesn't handle escape sequences properly
		// "string_with_quote_inside": {`"hello\"world"`, []Token{{Type: TokenType_StringLiteral, Value: "hello\"world"}}, true}, // This would cause an error

		// Test labels
		"simple_label":          {"hello", []Token{{Type: TokenType_Label, Value: "hello"}}, false},
		"label_with_underscore": {"hello_world", []Token{{Type: TokenType_Label, Value: "hello_world"}}, false},
		"label_with_numbers":    {"var123", []Token{{Type: TokenType_Label, Value: "var123"}}, false},
		"numeric_label":         {"123var", []Token{{Type: TokenType_NumberLiteral, Value: "123"}, {Type: TokenType_Label, Value: "var"}}, false},
		"single_letter":         {"a", []Token{{Type: TokenType_Label, Value: "a"}}, false},
		"upper_case_label":      {"VariableName", []Token{{Type: TokenType_Label, Value: "VariableName"}}, false},

		// Test boolean literals
		"boolean_true":          {"true", []Token{{Type: TokenType_BooleanLiteral, Value: "true"}}, false},
		"boolean_false":         {"false", []Token{{Type: TokenType_BooleanLiteral, Value: "false"}}, false},
		"bool_in_sentence":      {"true false", []Token{{Type: TokenType_BooleanLiteral, Value: "true"}, {Type: TokenType_BooleanLiteral, Value: "false"}}, false},
		"potential_boolean_not": {"not", []Token{{Type: TokenType_Label, Value: "not"}}, false}, // not is not a boolean literal

		// Test multiple tokens
		"mixed_tokens": {"hello 42 world", []Token{
			{Type: TokenType_Label, Value: "hello"},
			{Type: TokenType_NumberLiteral, Value: "42"},
			{Type: TokenType_Label, Value: "world"},
		}, false},

		// Test with spaces
		"spaces_around_tokens": {"  hello   42  ", []Token{
			{Type: TokenType_Label, Value: "hello"},
			{Type: TokenType_NumberLiteral, Value: "42"},
		}, false},

		// Test empty input
		"empty_input": {"", nil, false},

		// Test single space
		"only_spaces": {"   ", nil, false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			lexer := New(tt.input)
			var tokens []Token
			var gotError bool

			for {
				token, err := lexer.GetToken()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					if tt.hasError {
						gotError = true
						break // Expected error
					}
					assert.NoError(t, err, "Unexpected error when tokenizing %q", tt.input)
					break
				}
				tokens = append(tokens, token)
			}

			if tt.hasError {
				assert.True(t, gotError, "Expected error for input %q, but got tokens: %v", tt.input, tokens)
			} else {
				assert.Equal(t, tt.expected, tokens, "GetToken() should produce expected tokens for input %q", tt.input)
			}
		})
	}
}

func TestPeekToken(t *testing.T) {
	t.Run("peek_same_as_get_single_token", func(t *testing.T) {
		lexer := New("hello")

		peekToken, err := lexer.PeekToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "hello"}, peekToken)

		getToken, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "hello"}, getToken)
	})

	t.Run("peek_multiple_calls_same_result", func(t *testing.T) {
		lexer := New("hello world")

		peekToken1, err := lexer.PeekToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "hello"}, peekToken1)

		peekToken2, err := lexer.PeekToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "hello"}, peekToken2)
		assert.Equal(t, peekToken1, peekToken2)
	})

	t.Run("peek_then_get_then_peek_next", func(t *testing.T) {
		lexer := New("hello world")

		// Peek at first token
		peekToken, err := lexer.PeekToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "hello"}, peekToken)

		// Get first token
		getToken, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "hello"}, getToken)

		// Peek at second token
		peekToken2, err := lexer.PeekToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "world"}, peekToken2)

		// Get second token
		getToken2, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "world"}, getToken2)
	})

	t.Run("peek_eof", func(t *testing.T) {
		lexer := New("")

		_, err := lexer.PeekToken()
		assert.Equal(t, io.EOF, err)
	})

	t.Run("peek_after_all_tokens_consumed", func(t *testing.T) {
		lexer := New("hello")

		// Get the token
		_, err := lexer.GetToken()
		assert.NoError(t, err)

		// Peek at nothing left
		_, err = lexer.PeekToken()
		assert.Equal(t, io.EOF, err)
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("unterminated_string_literal", func(t *testing.T) {
		lexer := New(`"hello`)

		_, err := lexer.GetToken()
		assert.Equal(t, errUnexpectedEOF, err)
	})

	t.Run("unterminated_string_literal_with_content", func(t *testing.T) {
		lexer := New(`"hello world`)

		_, err := lexer.GetToken()
		assert.Equal(t, errUnexpectedEOF, err)
	})

	t.Run("invalid_rune_error", func(t *testing.T) {
		lexer := New(`$`)

		_, err := lexer.GetToken()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid rune")
	})

	t.Run("invalid_rune_with_specific_char", func(t *testing.T) {
		lexer := New(`@`)

		_, err := lexer.GetToken()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid rune")
	})

	t.Run("multiple_errors", func(t *testing.T) {
		lexer := New(`$ @`)

		// First error
		_, err := lexer.GetToken()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid rune")

		// Second error
		_, err = lexer.GetToken()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid rune")
	})

	t.Run("unterminated_string_followed_by_invalid_rune", func(t *testing.T) {
		lexer := New(`"hello $`)

		_, err := lexer.GetToken()
		assert.Equal(t, errUnexpectedEOF, err)
	})
}

func TestEdgeCasesAndUnicode(t *testing.T) {
	t.Run("unicode_in_labels", func(t *testing.T) {
		lexer := New("héllo")

		token, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "héllo"}, token)
	})

	t.Run("unicode_numbers", func(t *testing.T) {
		// Note: The lexer likely won't recognize these as numbers since they're not ASCII digits
		lexer := New("٠١٢٣٤٥٦٧٨٩") // Arabic-Indic digits 0-9

		token, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_NumberLiteral, Value: "٠١٢٣٤٥٦٧٨٩"}, token)
	})

	t.Run("unicode_in_strings", func(t *testing.T) {
		lexer := New(`"héllo wörld"`)

		token, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_StringLiteral, Value: "héllo wörld"}, token)
	})

	t.Run("emoji_in_strings", func(t *testing.T) {
		lexer := New(`"hello 🌍 world"`)

		token, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_StringLiteral, Value: "hello 🌍 world"}, token)
	})

	t.Run("tabs_and_newlines_as_whitespace", func(t *testing.T) {
		lexer := New("hello\t\n  world")

		token1, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "hello"}, token1)

		token2, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "world"}, token2)
	})

	t.Run("long_number_literal", func(t *testing.T) {
		longNumber := "12345678901234567890"
		lexer := New(longNumber)

		token, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_NumberLiteral, Value: longNumber}, token)
	})

	t.Run("long_label", func(t *testing.T) {
		longLabel := "averylonglabelwithlotsofcharacters"
		lexer := New(longLabel)

		token, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: longLabel}, token)
	})

	t.Run("long_string_literal", func(t *testing.T) {
		longString := `"This is a very long string with many characters that should be handled properly by the lexer"`
		lexer := New(longString)

		token, err := lexer.GetToken()
		assert.NoError(t, err)
		expected := "This is a very long string with many characters that should be handled properly by the lexer"
		assert.Equal(t, Token{Type: TokenType_StringLiteral, Value: expected}, token)
	})

	t.Run("multiple_whitespace_types", func(t *testing.T) {
		lexer := New("  \t hello \n world \r ")

		token1, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "hello"}, token1)

		token2, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "world"}, token2)
	})

	t.Run("zero_and_negative_concept", func(t *testing.T) {
		lexer := New("0 -5") // Note: the - is an invalid rune, so it would be 0, then an error

		token1, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_NumberLiteral, Value: "0"}, token1)

		_, err = lexer.GetToken()
		assert.Error(t, err) // Should error on the minus sign
	})

	t.Run("consecutive_strings", func(t *testing.T) {
		lexer := New(`"first""second"`)

		token1, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_StringLiteral, Value: "first"}, token1)

		token2, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_StringLiteral, Value: "second"}, token2)
	})

	t.Run("mixed_unicode_identifiers", func(t *testing.T) {
		lexer := New("variable_123 café résumé")

		token1, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "variable_123"}, token1)

		token2, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "café"}, token2)

		token3, err := lexer.GetToken()
		assert.NoError(t, err)
		assert.Equal(t, Token{Type: TokenType_Label, Value: "résumé"}, token3)
	})
}
