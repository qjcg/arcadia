package script

import (
	"fmt"
	"io"
	"strings"

	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
)

// Executor interface abstracts the shell execution methods needed by the interpreter.
type Executor interface {
	ExecutePipeline(pipe *parser.Pipeline) error
	RunCommand(name string, args []string, stdin io.Reader, stdout io.Writer) error
	SetVar(name, value string)
	GetVar(name string) string
	FuncDefs() map[string][]Stmt
	SetFuncDef(name string, body []Stmt)
}

// Interpreter executes script AST nodes.
type Interpreter struct {
	exec Executor
}

// NewInterpreter creates a new script interpreter.
func NewInterpreter(exec Executor) *Interpreter {
	return &Interpreter{exec: exec}
}

// ExecScript executes all statements in a script.
func (interp *Interpreter) ExecScript(script *Script) error {
	for _, stmt := range script.Stmts {
		if err := interp.ExecStmt(stmt); err != nil {
			return err
		}
	}
	return nil
}

// ExecStmt executes a single statement.
func (interp *Interpreter) ExecStmt(stmt Stmt) error {
	switch s := stmt.(type) {
	case *CommandStmt:
		return interp.exec.ExecutePipeline(s.Pipeline)

	case *IfStmt:
		return interp.execIf(s)

	case *ForStmt:
		return interp.execFor(s)

	case *WhileStmt:
		return interp.execWhile(s)

	case *UntilStmt:
		return interp.execUntil(s)

	case *TryStmt:
		return interp.execTry(s)

	case *FuncDef:
		interp.exec.SetFuncDef(s.Name, s.Body)
		return nil

	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

func (interp *Interpreter) execIf(s *IfStmt) error {
	// Check if condition succeeds (exit code 0)
	err := interp.exec.ExecutePipeline(s.Condition)
	if err == nil {
		// Condition succeeded, execute then body
		for _, stmt := range s.Then {
			if err := interp.ExecStmt(stmt); err != nil {
				return err
			}
		}
		return nil
	}

	// Check elif branches
	for _, ei := range s.ElseIf {
		err := interp.exec.ExecutePipeline(ei.Condition)
		if err == nil {
			for _, stmt := range ei.Body {
				if err := interp.ExecStmt(stmt); err != nil {
					return err
				}
			}
			return nil
		}
	}

	// Execute else branch
	for _, stmt := range s.Else {
		if err := interp.ExecStmt(stmt); err != nil {
			return err
		}
	}

	return nil
}

func (interp *Interpreter) execFor(s *ForStmt) error {
	for _, word := range s.Words {
		interp.exec.SetVar(s.Var, word)
		for _, stmt := range s.Body {
			if err := interp.ExecStmt(stmt); err != nil {
				return err
			}
		}
	}
	return nil
}

func (interp *Interpreter) execWhile(s *WhileStmt) error {
	for {
		err := interp.exec.ExecutePipeline(s.Condition)
		if err != nil {
			return nil
		}
		for _, stmt := range s.Body {
			if err := interp.ExecStmt(stmt); err != nil {
				return err
			}
		}
	}
}

func (interp *Interpreter) execUntil(s *UntilStmt) error {
	for {
		err := interp.exec.ExecutePipeline(s.Condition)
		if err == nil {
			return nil
		}
		for _, stmt := range s.Body {
			if err := interp.ExecStmt(stmt); err != nil {
				return err
			}
		}
	}
}

func (interp *Interpreter) execTry(s *TryStmt) error {
	for _, stmt := range s.Try {
		err := interp.ExecStmt(stmt)
		if err != nil {
			// Execute catch block
			for _, catchStmt := range s.Catch {
				if cerr := interp.ExecStmt(catchStmt); cerr != nil {
					return cerr
				}
			}
			return nil
		}
	}
	return nil
}

// ParseAndExec parses and executes a script string.
func (interp *Interpreter) ParseAndExec(input string) error {
	script, err := Parse(input)
	if err != nil {
		return err
	}
	return interp.ExecScript(script)
}

// ExecLine executes a single line of script, with support for control flow keywords.
// Returns true if the line was consumed as a control flow keyword.
func ExecLine(line string, interp *Interpreter) (bool, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return false, nil
	}

	words := strings.Fields(line)
	if len(words) == 0 {
		return false, nil
	}

	// Check for control flow keywords
	switch words[0] {
	case "if", "for", "while", "until", "function":
		script, err := Parse(line + "\n")
		if err != nil {
			return true, err
		}
		return true, interp.ExecScript(script)
	}

	return false, nil
}

// Helper functions for the interpreter

// IsCommand checks if a word is a command (not a keyword).
func IsCommand(word string) bool {
	switch word {
	case "if", "then", "else", "elif", "fi",
		"for", "in", "do", "done",
		"while", "until",
		"function":
		return false
	}
	return true
}
