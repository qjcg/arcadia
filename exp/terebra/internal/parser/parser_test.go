package parser

import (
	"testing"
)

func tokenize(t *testing.T, input string) []Token {
	t.Helper()
	return LexTokens(input)
}

func TestLexerSimpleWords(t *testing.T) {
	tokens := tokenize(t, "echo hello world")
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Type != TokenWord || tokens[0].Value != "echo" {
		t.Errorf("expected WORD(echo), got %s(%s)", tokens[0].Type, tokens[0].Value)
	}
	if tokens[1].Type != TokenWord || tokens[1].Value != "hello" {
		t.Errorf("expected WORD(hello), got %s(%s)", tokens[1].Type, tokens[1].Value)
	}
	if tokens[2].Type != TokenWord || tokens[2].Value != "world" {
		t.Errorf("expected WORD(world), got %s(%s)", tokens[2].Type, tokens[2].Value)
	}
	if tokens[3].Type != TokenEOF {
		t.Errorf("expected EOF, got %s", tokens[3].Type)
	}
}

func TestLexerPipe(t *testing.T) {
	tokens := tokenize(t, "ls | wc")
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Value != "ls" {
		t.Errorf("expected ls, got %s", tokens[0].Value)
	}
	if tokens[1].Type != TokenPipe {
		t.Errorf("expected PIPE, got %s", tokens[1].Type)
	}
	if tokens[2].Value != "wc" {
		t.Errorf("expected wc, got %s", tokens[2].Value)
	}
}

func TestLexerRedirects(t *testing.T) {
	tests := []struct {
		input string
		types []TokenType
		vals  []string
	}{
		{"echo > file", []TokenType{TokenWord, TokenRedirectOut, TokenWord, TokenEOF}, []string{"echo", ">", "file", ""}},
		{"echo >> file", []TokenType{TokenWord, TokenRedirectAppend, TokenWord, TokenEOF}, []string{"echo", ">>", "file", ""}},
		{"cat < file", []TokenType{TokenWord, TokenRedirectIn, TokenWord, TokenEOF}, []string{"cat", "<", "file", ""}},
		{"cmd 2> file", []TokenType{TokenWord, TokenRedirectErr, TokenWord, TokenEOF}, []string{"cmd", "2>", "file", ""}},
		{"cmd 2>> file", []TokenType{TokenWord, TokenRedirectErrAppend, TokenWord, TokenEOF}, []string{"cmd", "2>>", "file", ""}},
		{"cmd 2>&1", []TokenType{TokenWord, TokenRedirectErrOut, TokenEOF}, []string{"cmd", "2>&1", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenize(t, tt.input)
			if len(tokens) != len(tt.types) {
				t.Fatalf("expected %d tokens, got %d: %v", len(tt.types), len(tokens), tokens)
			}
			for i := range tt.types {
				if tokens[i].Type != tt.types[i] {
					t.Errorf("token %d: expected type %s, got %s", i, tt.types[i], tokens[i].Type)
				}
				if tokens[i].Value != tt.vals[i] {
					t.Errorf("token %d: expected value %q, got %q", i, tt.vals[i], tokens[i].Value)
				}
			}
		})
	}
}

func TestLexerQuoting(t *testing.T) {
	tests := []struct {
		input string
		vals  []string
	}{
		{"echo 'hello world'", []string{"echo", "hello world"}},
		{`echo "hello world"`, []string{"echo", "hello world"}},
		{`echo "hello's"`, []string{"echo", "hello's"}},
		{`echo 'say "hi"'`, []string{"echo", `say "hi"`}},
		{`echo hello\ world`, []string{"echo", "hello world"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			tokens := tokenize(t, tt.input)
			wordCount := 0
			for _, tok := range tokens {
				if tok.Type == TokenWord {
					if wordCount < len(tt.vals) && tok.Value != tt.vals[wordCount] {
						t.Errorf("word %d: expected %q, got %q", wordCount, tt.vals[wordCount], tok.Value)
					}
					wordCount++
				}
			}
			if wordCount != len(tt.vals) {
				t.Errorf("expected %d words, got %d", len(tt.vals), wordCount)
			}
		})
	}
}

func TestLexerComments(t *testing.T) {
	tokens := tokenize(t, "echo hello # this is a comment")
	// Should only have echo, hello, EOF
	if len(tokens) != 3 {
		t.Fatalf("expected 3 tokens (word, word, eof), got %d: %v", len(tokens), tokens)
	}
	if tokens[0].Value != "echo" {
		t.Errorf("expected echo, got %s", tokens[0].Value)
	}
}

func TestLexerAmpersand(t *testing.T) {
	tokens := tokenize(t, "sleep 1 &")
	if len(tokens) != 4 {
		t.Fatalf("expected 4 tokens, got %d: %v", len(tokens), tokens)
	}
	if tokens[2].Type != TokenAmpersand {
		t.Errorf("expected AMPERSAND, got %s", tokens[2].Type)
	}
}

func TestParseSimpleCommand(t *testing.T) {
	pipe, err := Parse("echo hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipe.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(pipe.Commands))
	}
	cmd := pipe.Commands[0]
	if cmd.Name != "echo" {
		t.Errorf("expected name echo, got %s", cmd.Name)
	}
	if len(cmd.Args) != 2 || cmd.Args[0] != "hello" || cmd.Args[1] != "world" {
		t.Errorf("expected args [hello world], got %v", cmd.Args)
	}
	if len(cmd.Redirects) != 0 {
		t.Errorf("expected 0 redirects, got %d", len(cmd.Redirects))
	}
}

func TestParsePipeline(t *testing.T) {
	pipe, err := Parse("ls -la | grep go | wc -l")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipe.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(pipe.Commands))
	}
	if pipe.Commands[0].Name != "ls" {
		t.Errorf("expected ls, got %s", pipe.Commands[0].Name)
	}
	if pipe.Commands[1].Name != "grep" {
		t.Errorf("expected grep, got %s", pipe.Commands[1].Name)
	}
	if pipe.Commands[2].Name != "wc" {
		t.Errorf("expected wc, got %s", pipe.Commands[2].Name)
	}
}

func TestParseRedirects(t *testing.T) {
	tests := []struct {
		input    string
		redirCnt int
		redirTp  RedirectType
		redirF   string
	}{
		{"echo hi > out.txt", 1, RedirectStdout, "out.txt"},
		{"echo hi >> out.txt", 1, RedirectAppend, "out.txt"},
		{"cat < in.txt", 1, RedirectStdin, "in.txt"},
		{"cmd 2> err.txt", 1, RedirectStderr, "err.txt"},
		{"cmd 2>> err.txt", 1, RedirectStderrAppend, "err.txt"},
		{"cmd 2>&1", 1, RedirectStderrToStdout, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pipe, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			cmd := pipe.Commands[0]
			if len(cmd.Redirects) != tt.redirCnt {
				t.Fatalf("expected %d redirects, got %d: %v", tt.redirCnt, len(cmd.Redirects), cmd.Redirects)
			}
			if cmd.Redirects[0].Type != tt.redirTp {
				t.Errorf("expected redirect type %s, got %s", tt.redirTp, cmd.Redirects[0].Type)
			}
			if cmd.Redirects[0].File != tt.redirF {
				t.Errorf("expected redirect file %q, got %q", tt.redirF, cmd.Redirects[0].File)
			}
		})
	}
}

func TestParseMultipleRedirects(t *testing.T) {
	pipe, err := Parse("cmd > out.txt 2> err.txt < in.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cmd := pipe.Commands[0]
	if len(cmd.Redirects) != 3 {
		t.Fatalf("expected 3 redirects, got %d", len(cmd.Redirects))
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestParseOnlyWhitespace(t *testing.T) {
	_, err := Parse("   ")
	if err == nil {
		t.Fatal("expected error for whitespace-only input")
	}
}

func TestParseComment(t *testing.T) {
	_, err := Parse("# just a comment")
	if err == nil {
		t.Fatal("expected error for comment-only input")
	}
}

func TestParseBackground(t *testing.T) {
	pipe, err := Parse("sleep 10 &")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pipe.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(pipe.Commands))
	}
	if !pipe.Commands[0].Background {
		t.Errorf("expected Background=true")
	}
	if pipe.Commands[0].Name != "sleep" {
		t.Errorf("expected sleep, got %s", pipe.Commands[0].Name)
	}
}

func TestLexerIntegrity(t *testing.T) {
	// Spot check: ensure all token types are accounted for
	inputs := []string{
		"a",
		"a|b",
		"a>b",
		"a>>b",
		"a<b",
		"a2>b",
		"a2>>b",
		"a2>&1",
	}
	for _, input := range inputs {
		tokens := LexTokens(input)
		if len(tokens) < 2 || tokens[len(tokens)-1].Type != TokenEOF {
			t.Errorf("%q: missing EOF token", input)
		}
		for _, tok := range tokens[:len(tokens)-1] {
			if tok.Type == TokenIllegal {
				t.Errorf("%q: illegal token: %q", input, tok.Value)
			}
		}
	}
}
