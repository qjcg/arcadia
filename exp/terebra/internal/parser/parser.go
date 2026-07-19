package parser

import (
	"fmt"
	"strings"
)

// Parse parses a shell input line into a single Pipeline.
func Parse(input string) (*Pipeline, error) {
	l := NewLexer(input)
	return parsePipeline(l)
}

// ParseScript parses a shell input line into a Script, supporting &&, ||, ; chaining.
func ParseScript(input string) (*Script, error) {
	l := NewLexer(input)
	return parseScript(l)
}

func parseScript(l *Lexer) (*Script, error) {
	script := &Script{}

	for {
		pipe, err := parsePipeline(l)
		if err != nil {
			return nil, err
		}
		script.Pipelines = append(script.Pipelines, pipe)

		// Check for chaining operators
		tok := l.NextToken()
		switch tok.Type {
		case TokenSemicolon:
			script.Ops = append(script.Ops, ChainingThen)
		case TokenAnd:
			script.Ops = append(script.Ops, ChainingAnd)
		case TokenOr:
			script.Ops = append(script.Ops, ChainingOr)
		case TokenEOF:
			return script, nil
		default:
			return nil, fmt.Errorf("unexpected token %q after command (expected &&, ||, ;, or newline)", tok)
		}
	}
}

func parsePipeline(l *Lexer) (*Pipeline, error) {
	pipe := &Pipeline{}

	for {
		cmd, err := parseCommand(l)
		if err != nil {
			return nil, err
		}

		// Check for background operator
		peek := l.PeekToken()
		if peek.Type == TokenAmpersand {
			l.NextToken()
			cmd.Background = true
		}

		pipe.Commands = append(pipe.Commands, cmd)

		// Check what comes after the command
		peek = l.PeekToken()
		switch peek.Type {
		case TokenPipe, TokenAugerPipe:
			l.NextToken()
			if peek.Type == TokenPipe {
				pipe.Connects = append(pipe.Connects, ConnectPipe)
			} else {
				// Check for encoder pipe: |>json, |>yaml, |>cue
				encPeek := l.PeekToken()
				if encPeek.Type == TokenWord {
					switch encPeek.Value {
					case "json", "yaml", "cue":
						l.NextToken()
						pipe.Encoder = encPeek.Value
						pipe.Connects = append(pipe.Connects, ConnectAuger)
						// If encoder is the last pipe, we're done
						return pipe, nil
					}
				}
				pipe.Connects = append(pipe.Connects, ConnectAuger)
			}
			continue
		case TokenEOF, TokenSemicolon, TokenAnd, TokenOr:
			return pipe, nil
		default:
			return nil, fmt.Errorf("unexpected token %q after command", peek)
		}
	}
}

func parseCommand(l *Lexer) (*Command, error) {
	cmd := &Command{}

	// Skip leading redirects
	for {
		peek := l.PeekToken()
		if isRedirect(peek.Type) {
			tok := l.NextToken()
			redir, err := parseRedirect(tok, l)
			if err != nil {
				return nil, err
			}
			cmd.Redirects = append(cmd.Redirects, redir)
			continue
		}
		if peek.Type == TokenWord {
			break
		}
		return nil, fmt.Errorf("expected command name, got %s", peek.Type)
	}

	// Get command name
	tok := l.NextToken()
	if tok.Type != TokenWord {
		return nil, fmt.Errorf("expected command name, got %s", tok.Type)
	}
	cmd.Name = tok.Value

	// Parse args and redirects, stopping at pipe, &, auger pipe, chaining ops, or eof
	for {
		peek := l.PeekToken()
		if peek.Type == TokenEOF || peek.Type == TokenPipe || peek.Type == TokenAugerPipe || peek.Type == TokenAmpersand || peek.Type == TokenSemicolon || peek.Type == TokenAnd || peek.Type == TokenOr {
			break
		}
		if isRedirect(peek.Type) {
			tok := l.NextToken()
			redir, err := parseRedirect(tok, l)
			if err != nil {
				return nil, err
			}
			cmd.Redirects = append(cmd.Redirects, redir)
			continue
		}
		if peek.Type == TokenWord {
			tok := l.NextToken()
			cmd.Args = append(cmd.Args, tok.Value)
			continue
		}

		return nil, fmt.Errorf("unexpected token %s in command", peek.Type)
	}

	return cmd, nil
}

func isRedirect(tt TokenType) bool {
	switch tt {
	case TokenRedirectOut, TokenRedirectAppend, TokenRedirectIn,
		TokenRedirectErr, TokenRedirectErrAppend, TokenRedirectErrOut,
		TokenRedirectHeredoc, TokenRedirectHeredocDash:
		return true
	}
	return false
}

func parseRedirect(tok Token, l *Lexer) (*Redirect, error) {
	r := &Redirect{}

	switch tok.Type {
	case TokenRedirectOut:
		r.Type = RedirectStdout
	case TokenRedirectErr:
		r.Type = RedirectStderr
	case TokenRedirectAppend:
		r.Type = RedirectAppend
	case TokenRedirectErrAppend:
		r.Type = RedirectStderrAppend
	case TokenRedirectIn:
		r.Type = RedirectStdin
	case TokenRedirectErrOut:
		r.Type = RedirectStderrToStdout
		r.File = "1"
		return r, nil
	case TokenRedirectHeredoc:
		r.Type = RedirectHeredoc
	case TokenRedirectHeredocDash:
		r.Type = RedirectHeredocDash
	default:
		return nil, fmt.Errorf("internal error: not a redirect token: %s", tok)
	}

	// For heredocs, read the delimiter and then the heredoc content
	if r.Type == RedirectHeredoc || r.Type == RedirectHeredocDash {
		next := l.NextToken()
		if next.Type != TokenWord {
			return nil, fmt.Errorf("expected heredoc delimiter after %s, got %s", tok.Type, next)
		}
		delimiter := next.Value
		// Strip quotes from delimiter to determine if expansion is needed
		quoted := false
		if (len(delimiter) >= 2 && delimiter[0] == '\'' && delimiter[len(delimiter)-1] == '\'') ||
			(len(delimiter) >= 2 && delimiter[0] == '"' && delimiter[len(delimiter)-1] == '"') {
			quoted = true
			delimiter = delimiter[1 : len(delimiter)-1]
		}
		r.Quoted = quoted
		r.File = delimiter
		// Read heredoc content from remaining input
		var content strings.Builder
		for l.pos < len(l.input) {
			// Skip to next line
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.pos++
			}
			if l.pos < len(l.input) {
				l.pos++ // skip \n
			}
			// Read the next line
			lineStart := l.pos
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.pos++
			}
			line := l.input[lineStart:l.pos]
			// Check for delimiter
			trimmed := line
			if r.Type == RedirectHeredocDash {
				trimmed = strings.TrimLeft(line, "\t")
			}
			if trimmed == delimiter {
				break
			}
			if content.Len() > 0 {
				content.WriteByte('\n')
			}
			content.WriteString(line)
		}
		// Add trailing newline (the newline before the delimiter)
		if content.Len() > 0 {
			content.WriteByte('\n')
		}
		r.Content = content.String()
		return r, nil
	}

	next := l.NextToken()
	if next.Type != TokenWord {
		return nil, fmt.Errorf("expected filename after redirect %s, got %s", tok.Type, next)
	}
	r.File = next.Value

	return r, nil
}

// LexTokens is a helper for debugging and testing.
func LexTokens(input string) []Token {
	l := NewLexer(input)
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF || tok.Type == TokenIllegal {
			break
		}
	}
	return tokens
}

// IsUnquotedString checks if a string is a simple unquoted word.
func IsUnquotedString(s string) bool {
	return !strings.ContainsAny(s, "'\"\\ \t\n")
}
