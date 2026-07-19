package script

import "github.com/qjcg/arcadia/exp/terebra/internal/parser"

// Stmt is a statement in the scripting language.
type Stmt interface {
	stmt()
}

// CommandStmt is a single command or pipeline.
type CommandStmt struct {
	Pipeline *parser.Pipeline
}

func (s *CommandStmt) stmt() {}

// IfStmt represents an if/then/elif/else/fi block.
type IfStmt struct {
	Condition *parser.Pipeline
	Then      []Stmt
	ElseIf    []*ElseIfStmt
	Else      []Stmt
}

func (s *IfStmt) stmt() {}

// ElseIfStmt represents an elif block.
type ElseIfStmt struct {
	Condition *parser.Pipeline
	Body      []Stmt
}

// ForStmt represents a for/in/do/done loop.
type ForStmt struct {
	Var   string
	Words []string
	Body  []Stmt
}

func (s *ForStmt) stmt() {}

// WhileStmt represents a while/do/done loop.
type WhileStmt struct {
	Condition *parser.Pipeline
	Body      []Stmt
}

func (s *WhileStmt) stmt() {}

// UntilStmt represents an until/do/done loop.
type UntilStmt struct {
	Condition *parser.Pipeline
	Body      []Stmt
}

func (s *UntilStmt) stmt() {}

// FuncDef represents a function definition.
type FuncDef struct {
	Name string
	Body []Stmt
}

func (s *FuncDef) stmt() {}

// TryStmt represents a try/catch/end block.
type TryStmt struct {
	Try   []Stmt
	Catch []Stmt
}

func (s *TryStmt) stmt() {}

type Script struct {
	Stmts []Stmt
}
