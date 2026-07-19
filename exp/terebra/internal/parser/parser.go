package parser

import (
	"fmt"
	"strings"
)

// Parse parses a shell input line into a Pipeline.
func Parse(input string) (*Pipeline, error) {
	l := NewLexer(input)
	return parsePipeline(l)
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
		tok := l.NextToken()
		if tok.Type == TokenPipe {
			pipe.Connects = append(pipe.Connects, ConnectPipe)
			continue
		}
		if tok.Type == TokenAugerPipe {
			pipe.Connects = append(pipe.Connects, ConnectAuger)
			continue
		}
		if tok.Type == TokenEOF {
			break
		}
		return nil, fmt.Errorf("unexpected token %q after command", tok)
	}

	if len(pipe.Commands) == 0 {
		return nil, fmt.Errorf("empty pipeline")
	}

	return pipe, nil
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

	// Parse args and redirects, stopping at pipe, &, auger pipe, or eof
	for {
		peek := l.PeekToken()
		if peek.Type == TokenEOF || peek.Type == TokenPipe || peek.Type == TokenAugerPipe || peek.Type == TokenAmpersand {
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
		TokenRedirectErr, TokenRedirectErrAppend, TokenRedirectErrOut:
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
	default:
		return nil, fmt.Errorf("internal error: not a redirect token: %s", tok)
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
