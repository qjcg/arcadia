package parser

import (
	"testing"

	"github.com/qjcg/arcadia/x/elbereth/internal/lexer"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "defn",
			input: `(defn add [x y] (+ x y))`,
		},
		{
			name:  "deftype struct",
			input: `(deftype Person {name string age int})`,
		},
		{
			name: "deftype sum type",
			input: `(deftype Result
  (:ok int)
  (:err string))`,
		},
		{
			name:  "package",
			input: `(package myutils)`,
		},
		{
			name:  "import",
			input: `(import "fmt" "math" [time "time"])`,
		},
		{
			name:    "unclosed paren",
			input:   `(defn add [x y] (+ x y)`,
			wantErr: true,
		},
		{
			name:    "unclosed bracket",
			input:   `(defn add [x y (+ x y))`,
			wantErr: true,
		},
		{
			name:    "missing name",
			input:   `(defn [x y] (+ x y))`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			_, err := p.Parse()
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
