package parser

import "strings"

type TokenType int

const (
	TokenWord TokenType = iota
	TokenPipe
	TokenPipeErr // |&
	TokenAugerPipe
	TokenAmpersand
	TokenAnd
	TokenOr
	TokenSemicolon
	TokenRedirectOut
	TokenRedirectAppend
	TokenRedirectIn
	TokenRedirectErr
	TokenRedirectErrAppend
	TokenRedirectErrOut
	TokenRedirectBoth        // &>
	TokenRedirectBothAppend  // &>>
	TokenRedirectHeredoc     // <<
	TokenRedirectHeredocDash // <<-
	TokenRedirectHereString  // <<<
	TokenEOF
	TokenIllegal
)

var tokenNames = map[TokenType]string{
	TokenWord:                "WORD",
	TokenPipe:                "PIPE",
	TokenPipeErr:             "PIPE_ERR",
	TokenAugerPipe:           "AUGER_PIPE",
	TokenAmpersand:           "AMPERSAND",
	TokenAnd:                 "AND",
	TokenOr:                  "OR",
	TokenSemicolon:           "SEMICOLON",
	TokenRedirectOut:         "REDIRECT_OUT",
	TokenRedirectAppend:      "REDIRECT_APPEND",
	TokenRedirectIn:          "REDIRECT_IN",
	TokenRedirectErr:         "REDIRECT_ERR",
	TokenRedirectErrAppend:   "REDIRECT_ERR_APPEND",
	TokenRedirectErrOut:      "REDIRECT_ERR_OUT",
	TokenRedirectBoth:        "REDIRECT_BOTH",
	TokenRedirectBothAppend:  "REDIRECT_BOTH_APPEND",
	TokenRedirectHeredoc:     "REDIRECT_HEREDOC",
	TokenRedirectHeredocDash: "REDIRECT_HEREDOC_DASH",
	TokenRedirectHereString:  "REDIRECT_HERE_STRING",
	TokenEOF:                 "EOF",
	TokenIllegal:             "ILLEGAL",
}

func (tt TokenType) String() string {
	return tokenNames[tt]
}

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	if t.Type == TokenWord || t.Type == TokenIllegal {
		return t.Value
	}
	return t.Type.String()
}

type Lexer struct {
	input string
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF}
	}

	ch := l.input[l.pos]

	// Check for digit followed by > (e.g., 2>, 2>>, 2>&1)
	if ch >= '0' && ch <= '9' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '>' {
		return l.readFdRedirect()
	}

	switch ch {
	case '|':
		return l.readPipe()
	case '&':
		return l.readAmpersand()
	case '>':
		return l.readRedirectOut()
	case '<':
		return l.readRedirectIn()
	case ';':
		l.pos++
		return Token{Type: TokenSemicolon, Value: ";"}
	case '#':
		// Comment — skip to end of line
		for l.pos < len(l.input) && l.input[l.pos] != '\n' {
			l.pos++
		}
		// Skip the newline
		if l.pos < len(l.input) {
			l.pos++
		}
		return l.NextToken()
	default:
		return l.readWord()
	}
}

// readFdRedirect handles a digit followed by > (e.g., 2>, 2>>, 2>&1).
func (l *Lexer) readFdRedirect() Token {
	fd := int(l.input[l.pos] - '0')
	l.pos++ // skip digit, now at '>'

	// Check for >>
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '>' {
		l.pos += 2
		if fd == 2 {
			return Token{Type: TokenRedirectErrAppend, Value: "2>>"}
		}
		return Token{Type: TokenRedirectAppend, Value: ">>"}
	}

	// Check for >&
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '&' {
		l.pos += 2 // skip '>&'
		l.skipWhitespace()
		// Read target fd
		targetStart := l.pos
		for l.pos < len(l.input) && l.input[l.pos] >= '0' && l.input[l.pos] <= '9' {
			l.pos++
		}
		target := l.input[targetStart:l.pos]
		if fd == 2 && target == "1" {
			return Token{Type: TokenRedirectErrOut, Value: "2>&1"}
		}
		return Token{Type: TokenRedirectErrOut, Value: strings.TrimSpace(l.input[l.pos-1:])}
	}

	// Simple > redirect with fd
	l.pos++
	if fd == 2 {
		return Token{Type: TokenRedirectErr, Value: "2>"}
	}
	return Token{Type: TokenRedirectOut, Value: ">"}
}

// readPipe handles the | operator and its multi-char forms.
func (l *Lexer) readPipe() Token {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '&' {
		l.pos += 2
		return Token{Type: TokenPipeErr, Value: "|&"}
	}
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '>' {
		l.pos += 2
		return Token{Type: TokenAugerPipe, Value: "|>"}
	}
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '|' {
		l.pos += 2
		return Token{Type: TokenOr, Value: "||"}
	}
	l.pos++
	return Token{Type: TokenPipe, Value: "|"}
}

// readAmpersand handles the & operator and its multi-char forms.
func (l *Lexer) readAmpersand() Token {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '>' {
		l.pos += 2
		if l.pos < len(l.input) && l.input[l.pos] == '>' {
			l.pos++
			return Token{Type: TokenRedirectBothAppend, Value: "&>>"}
		}
		return Token{Type: TokenRedirectBoth, Value: "&>"}
	}
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '&' {
		l.pos += 2
		return Token{Type: TokenAnd, Value: "&&"}
	}
	l.pos++
	return Token{Type: TokenAmpersand, Value: "&"}
}

// readRedirectOut handles the > operator and its multi-char forms.
func (l *Lexer) readRedirectOut() Token {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '>' {
		l.pos += 2
		return Token{Type: TokenRedirectAppend, Value: ">>"}
	}
	l.pos++
	return Token{Type: TokenRedirectOut, Value: ">"}
}

// readRedirectIn handles the < operator and its multi-char forms.
func (l *Lexer) readRedirectIn() Token {
	if l.pos+1 < len(l.input) && l.input[l.pos+1] == '<' {
		l.pos += 2
		// Check for <<< (here-string)
		if l.pos < len(l.input) && l.input[l.pos] == '<' {
			l.pos++
			return Token{Type: TokenRedirectHereString, Value: "<<<"}
		}
		// Check for <<-
		if l.pos < len(l.input) && l.input[l.pos] == '-' {
			l.pos++
			return Token{Type: TokenRedirectHeredocDash, Value: "<<-"}
		}
		return Token{Type: TokenRedirectHeredoc, Value: "<<"}
	}
	l.pos++
	return Token{Type: TokenRedirectIn, Value: "<"}
}

// PeekToken returns the next token without consuming it.
func (l *Lexer) PeekToken() Token {
	pos := l.pos
	tok := l.NextToken()
	l.pos = pos
	return tok
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch != ' ' && ch != '\t' && ch != '\n' {
			break
		}
		l.pos++
	}
}

func (l *Lexer) readWord() Token {
	var word strings.Builder
	var inSingle, inDouble, escaped bool

	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if escaped {
			word.WriteByte(ch)
			l.pos++
			escaped = false
			continue
		}

		if ch == '\\' && !inSingle {
			escaped = true
			l.pos++
			continue
		}

		if inSingle {
			if ch == '\'' {
				inSingle = false
				l.pos++
				continue
			}
			word.WriteByte(ch)
			l.pos++
			continue
		}

		if inDouble {
			if l.readDoubleQuoted(&word) {
				inDouble = false
			}
			continue
		}

		if ch == '\'' {
			inSingle = true
			l.pos++
			continue
		}

		if ch == '"' {
			inDouble = true
			l.pos++
			continue
		}

		// Operators and whitespace end the word
		if ch == '|' || ch == '>' || ch == '<' || ch == '&' || ch == ';' || ch == '#' || ch == ' ' || ch == '\t' || ch == '\n' {
			break
		}

		// Check for $((...)) arithmetic or $(...) command substitution - don't split on whitespace inside
		if ch == '$' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '(' {
			if l.pos+2 < len(l.input) && l.input[l.pos+2] == '(' {
				// $((...)) arithmetic
				l.readBalanced(&word, "$((", 2)
			} else {
				// $(...) command substitution
				l.readBalanced(&word, "$(", 1)
			}
			continue
		}

		// Check for ${...} variable expansion - don't split on whitespace inside
		if ch == '$' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '{' {
			l.readBalanced(&word, "${", 1)
			continue
		}

		word.WriteByte(ch)
		l.pos++
	}

	return Token{Type: TokenWord, Value: word.String()}
}

// readDoubleQuoted consumes one character while inside a double-quoted string.
// It returns true if the closing quote was consumed (ending the quoted region).
func (l *Lexer) readDoubleQuoted(word *strings.Builder) bool {
	ch := l.input[l.pos]
	if ch == '"' {
		l.pos++
		return true
	}
	if ch == '\\' {
		l.pos++
		if l.pos < len(l.input) {
			next := l.input[l.pos]
			switch next {
			case '"', '\\', '`', '$', '\n':
				word.WriteByte(next)
			default:
				word.WriteByte('\\')
				word.WriteByte(next)
			}
			l.pos++
		}
		return false
	}
	word.WriteByte(ch)
	l.pos++
	return false
}

// readBalanced writes the given opening sequence to word, then reads until the
// matching closing delimiters (depth counts the number of open delimiters).
func (l *Lexer) readBalanced(word *strings.Builder, open string, depth int) {
	word.WriteString(open)
	l.pos += len(open)
	for l.pos < len(l.input) && depth > 0 {
		c := l.input[l.pos]
		if c == '(' || c == '{' {
			depth++
		} else if c == ')' || c == '}' {
			depth--
			if depth == 0 {
				word.WriteByte(c)
				l.pos++
				break
			}
		}
		word.WriteByte(c)
		l.pos++
	}
}
