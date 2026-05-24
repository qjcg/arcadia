package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

// TokenType represents the type of a token
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError

	// Literals
	TokenInt
	TokenFloat
	TokenString
	TokenKeyword
	TokenTrue
	TokenFalse
	TokenNil

	// Identifiers and symbols
	TokenSymbol

	// Delimiters
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenLBrace
	TokenRBrace
	TokenQuote
	TokenBackquote
	TokenUnquote
	TokenUnquoteSplice
	TokenHashLang

	// Special
	TokenNewline
)

// Token represents a lexical token
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// Lexer tokenizes Elbereth source code
type Lexer struct {
	input  string
	pos    int
	line   int
	column int
}

// New creates a new lexer for the given input
func New(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

// Next returns the next token
func (l *Lexer) Next() Token {
	l.skipWhitespaceExceptNewline()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Line: l.line, Column: l.column}
	}

	ch := l.input[l.pos]
	line, col := l.line, l.column

	// Handle newline
	if ch == '\n' {
		l.advance()
		return Token{Type: TokenNewline, Line: line, Column: col}
	}

	// Single-character tokens
	switch ch {
	case '(':
		l.advance()
		return Token{Type: TokenLParen, Value: "(", Line: line, Column: col}
	case ')':
		l.advance()
		return Token{Type: TokenRParen, Value: ")", Line: line, Column: col}
	case '[':
		l.advance()
		return Token{Type: TokenLBracket, Value: "[", Line: line, Column: col}
	case ']':
		l.advance()
		return Token{Type: TokenRBracket, Value: "]", Line: line, Column: col}
	case '{':
		l.advance()
		return Token{Type: TokenLBrace, Value: "{", Line: line, Column: col}
	case '}':
		l.advance()
		return Token{Type: TokenRBrace, Value: "}", Line: line, Column: col}
	case '\'':
		l.advance()
		return Token{Type: TokenQuote, Value: "'", Line: line, Column: col}
	case '`':
		l.advance()
		return Token{Type: TokenBackquote, Value: "`", Line: line, Column: col}
	}

	// Unquote and unquote-splice
	if ch == ',' {
		l.advance()
		if l.pos < len(l.input) && l.input[l.pos] == '@' {
			l.advance()
			return Token{Type: TokenUnquoteSplice, Line: line, Column: col}
		}
		return Token{Type: TokenUnquote, Line: line, Column: col}
	}

	// String literals
	if ch == '"' {
		return l.lexString()
	}

	// Comments
	if ch == ';' {
		return l.lexLineComment()
	}

	// Block comments
	if ch == '#' {
		if l.peekNext() == '|' {
			return l.lexBlockComment()
		}
		// Check for #lang
		if l.input[l.pos:min(l.pos+5, len(l.input))] == "#lang" {
			return l.lexHashLang()
		}
	}

	// Keywords
	if ch == ':' {
		return l.lexKeyword()
	}

	// Numbers or symbols
	if unicode.IsDigit(rune(ch)) || ch == '-' {
		return l.lexNumberOrSymbol()
	}

	// Symbols
	if isSymbolChar(ch) {
		return l.lexSymbol()
	}

	// Unknown
	l.advance()
	return Token{
		Type:   TokenError,
		Value:  string(ch),
		Line:   line,
		Column: col,
	}
}

// NextNonNewline returns the next non-newline token
func (l *Lexer) NextNonNewline() Token {
	for {
		tok := l.Next()
		if tok.Type != TokenNewline {
			return tok
		}
	}
}

// ============================================================================
// Helper methods
// ============================================================================

func (l *Lexer) advance() {
	if l.pos < len(l.input) && l.input[l.pos] == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	l.pos++
}

func (l *Lexer) peek() byte {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) peekNext() byte {
	if l.pos+1 >= len(l.input) {
		return 0
	}
	return l.input[l.pos+1]
}

func (l *Lexer) skipWhitespaceExceptNewline() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
		} else {
			break
		}
	}
}

func (l *Lexer) lexHashLang() Token {
	line, col := l.line, l.column
	l.pos += 5 // skip #lang
	l.column += 5
	return Token{Type: TokenHashLang, Value: "#lang", Line: line, Column: col}
}

func (l *Lexer) lexString() Token {
	startLine, startCol := l.line, l.column
	l.advance() // skip opening quote

	var result string
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '"' {
			l.advance() // skip closing quote
			return Token{
				Type:   TokenString,
				Value:  result,
				Line:   startLine,
				Column: startCol,
			}
		}
		if ch == '\\' && l.pos+1 < len(l.input) {
			l.advance()
			switch l.input[l.pos] {
			case 'n':
				result += "\n"
			case 't':
				result += "\t"
			case 'r':
				result += "\r"
			case '\\':
				result += "\\"
			case '"':
				result += "\""
			default:
				result += string(l.input[l.pos])
			}
			l.advance()
		} else {
			result += string(ch)
			l.advance()
		}
	}

	return Token{
		Type:   TokenError,
		Value:  "unterminated string",
		Line:   startLine,
		Column: startCol,
	}
}

func (l *Lexer) lexLineComment() Token {
	for l.pos < len(l.input) && l.input[l.pos] != '\n' {
		l.advance()
	}
	// Return next token after comment
	return l.Next()
}

func (l *Lexer) lexBlockComment() Token {
	startLine, startCol := l.line, l.column
	l.advance() // skip #
	l.advance() // skip |

	for l.pos < len(l.input) {
		if l.input[l.pos] == '|' && l.peekNext() == '#' {
			l.advance() // skip |
			l.advance() // skip #
			return l.Next()
		}
		l.advance()
	}

	return Token{
		Type:   TokenError,
		Value:  "unterminated block comment",
		Line:   startLine,
		Column: startCol,
	}
}

func (l *Lexer) lexKeyword() Token {
	startLine, startCol := l.line, l.column
	l.advance() // skip :

	var name strings.Builder
	for l.pos < len(l.input) && isSymbolChar(l.input[l.pos]) {
		name.WriteString(string(l.input[l.pos]))
		l.advance()
	}

	if name.String() == "" {
		return Token{
			Type:   TokenError,
			Value:  "empty keyword",
			Line:   startLine,
			Column: startCol,
		}
	}

	return Token{
		Type:   TokenKeyword,
		Value:  name.String(),
		Line:   startLine,
		Column: startCol,
	}
}

func (l *Lexer) lexNumberOrSymbol() Token {
	startPos := l.pos
	startLine, startCol := l.line, l.column

	// Check for hex, binary, or negative numbers
	if l.peek() == '-' {
		if l.peekNext() != 0 && !unicode.IsDigit(rune(l.peekNext())) {
			return l.lexSymbolFrom(startPos, startLine, startCol)
		}
		l.advance()
	}

	// Check for hex (0x) or binary (0b)
	if l.peek() == '0' && l.peekNext() != 0 {
		next := l.peekNext()
		if next == 'x' || next == 'X' {
			return l.lexHex(startLine, startCol)
		}
		if next == 'b' || next == 'B' {
			return l.lexBinary(startLine, startCol)
		}
	}

	// Try to lex as number
	hasDecimal := false
	for l.pos < len(l.input) {
		ch := l.peek()
		if unicode.IsDigit(rune(ch)) {
			l.advance()
		} else if ch == '.' && !hasDecimal {
			hasDecimal = true
			l.advance()
		} else {
			break
		}
	}

	value := l.input[startPos:l.pos]

	if hasDecimal {
		return Token{
			Type:   TokenFloat,
			Value:  value,
			Line:   startLine,
			Column: startCol,
		}
	}

	return Token{
		Type:   TokenInt,
		Value:  value,
		Line:   startLine,
		Column: startCol,
	}
}

func (l *Lexer) lexHex(startLine, startCol int) Token {
	l.advance() // skip 0
	l.advance() // skip x

	startPos := l.pos
	for l.pos < len(l.input) {
		ch := l.peek()
		if (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F') {
			l.advance()
		} else {
			break
		}
	}

	return Token{
		Type:   TokenInt,
		Value:  "0x" + l.input[startPos:l.pos],
		Line:   startLine,
		Column: startCol,
	}
}

func (l *Lexer) lexBinary(startLine, startCol int) Token {
	l.advance() // skip 0
	l.advance() // skip b

	startPos := l.pos
	for l.pos < len(l.input) {
		ch := l.peek()
		if ch == '0' || ch == '1' {
			l.advance()
		} else {
			break
		}
	}

	return Token{
		Type:   TokenInt,
		Value:  "0b" + l.input[startPos:l.pos],
		Line:   startLine,
		Column: startCol + 2,
	}
}

func (l *Lexer) lexSymbol() Token {
	return l.lexSymbolFrom(l.pos, l.line, l.column)
}

func (l *Lexer) lexSymbolFrom(startPos int, startLine, startCol int) Token {
	for l.pos < len(l.input) && isSymbolChar(l.input[l.pos]) {
		l.advance()
	}

	value := l.input[startPos:l.pos]

	// Check for reserved words
	switch value {
	case "true":
		return Token{Type: TokenTrue, Line: startLine, Column: startCol}
	case "false":
		return Token{Type: TokenFalse, Line: startLine, Column: startCol}
	case "nil":
		return Token{Type: TokenNil, Line: startLine, Column: startCol}
	}

	return Token{
		Type:   TokenSymbol,
		Value:  value,
		Line:   startLine,
		Column: startCol,
	}
}

func isSymbolChar(ch byte) bool {
	if unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) {
		return true
	}
	switch ch {
	case '-', '_', '+', '*', '/', '%', '=', '!', '<', '>', '&', '|', '?', '.', ':', '@', '$':
		return true
	}
	return false
}

// TokenTypeString returns a string representation of a token type
func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "ERROR"
	case TokenInt:
		return "INT"
	case TokenFloat:
		return "FLOAT"
	case TokenString:
		return "STRING"
	case TokenKeyword:
		return "KEYWORD"
	case TokenTrue:
		return "TRUE"
	case TokenFalse:
		return "FALSE"
	case TokenNil:
		return "NIL"
	case TokenSymbol:
		return "SYMBOL"
	case TokenLParen:
		return "LPAREN"
	case TokenRParen:
		return "RPAREN"
	case TokenLBracket:
		return "LBRACKET"
	case TokenRBracket:
		return "RBRACKET"
	case TokenLBrace:
		return "LBRACE"
	case TokenRBrace:
		return "RBRACE"
	case TokenQuote:
		return "QUOTE"
	case TokenBackquote:
		return "BACKQUOTE"
	case TokenUnquote:
		return "UNQUOTE"
	case TokenUnquoteSplice:
		return "UNQUOTE_SPLICE"
	case TokenHashLang:
		return "HASHLANG"
	case TokenNewline:
		return "NEWLINE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t)
	}
}
