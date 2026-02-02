package lexer

import (
	"testing"
)

func TestLexer(t *testing.T) {
	input := `(defn add [x y] (+ x y))`
	l := New(input)

	expected := []struct {
		typ   TokenType
		value string
	}{
		{TokenLParen, "("},
		{TokenSymbol, "defn"},
		{TokenSymbol, "add"},
		{TokenLBracket, "["},
		{TokenSymbol, "x"},
		{TokenSymbol, "y"},
		{TokenRBracket, "]"},
		{TokenLParen, "("},
		{TokenSymbol, "+"},
		{TokenSymbol, "x"},
		{TokenSymbol, "y"},
		{TokenRParen, ")"},
		{TokenRParen, ")"},
		{TokenEOF, ""},
	}

	for _, exp := range expected {
		tok := l.Next()
		if tok.Type != exp.typ {
			t.Errorf("expected type %v, got %v", exp.typ, tok.Type)
		}
		if exp.value != "" && tok.Value != exp.value {
			t.Errorf("expected value %q, got %q", exp.value, tok.Value)
		}
	}
}

func TestLexStrings(t *testing.T) {
	input := `"hello" "world\n" "with \"quotes\""`
	l := New(input)

	expected := []string{
		"hello",
		"world\n",
		"with \"quotes\"",
	}

	for _, exp := range expected {
		tok := l.Next()
		if tok.Type != TokenString {
			t.Errorf("expected TokenString, got %v", tok.Type)
		}
		if tok.Value != exp {
			t.Errorf("expected value %q, got %q", exp, tok.Value)
		}
	}
}

func TestLexNumbers(t *testing.T) {
	input := `123 45.67 -89 0x1f 0b1010`
	l := New(input)

	expected := []struct {
		typ   TokenType
		value string
	}{
		{TokenInt, "123"},
		{TokenFloat, "45.67"},
		{TokenInt, "-89"},
		{TokenInt, "0x1f"},
		{TokenInt, "0b1010"},
	}

	for _, exp := range expected {
		tok := l.Next()
		if tok.Type != exp.typ {
			t.Errorf("expected type %v, got %v for %s", exp.typ, tok.Type, exp.value)
		}
		if tok.Value != exp.value {
			t.Errorf("expected value %q, got %q", exp.value, tok.Value)
		}
	}
}
