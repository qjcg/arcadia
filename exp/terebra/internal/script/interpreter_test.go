package script

import (
	"errors"
	"io"
	"testing"

	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
)

// fakeExec is a script.Executor that records calls and returns scripted results.
type fakeExec struct {
	pipelineErr error
	runErr      error
	vars        map[string]string
	funcs       map[string][]Stmt
	pipelines   []*parser.Pipeline
	// callErrors, when set, returns the error for the i-th ExecutePipeline call
	// (0-based); calls beyond the slice return nil.
	callErrors []error
}

func newFakeExec() *fakeExec {
	return &fakeExec{
		vars:  map[string]string{},
		funcs: map[string][]Stmt{},
	}
}

func (f *fakeExec) ExecutePipeline(pipe *parser.Pipeline) error {
	f.pipelines = append(f.pipelines, pipe)
	if f.callErrors != nil {
		i := len(f.pipelines) - 1
		if i < len(f.callErrors) {
			return f.callErrors[i]
		}
		return nil
	}
	return f.pipelineErr
}

func (f *fakeExec) RunCommand(name string, args []string, stdin io.Reader, stdout io.Writer) error {
	return f.runErr
}

func (f *fakeExec) SetVar(name, value string) { f.vars[name] = value }
func (f *fakeExec) GetVar(name string) string { return f.vars[name] }
func (f *fakeExec) FuncDefs() map[string][]Stmt {
	return f.funcs
}

func (f *fakeExec) SetFuncDef(name string, body []Stmt) {
	f.funcs[name] = body
}

func TestExecStmtCommand(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	stmt := &CommandStmt{Pipeline: &parser.Pipeline{}}
	if err := interp.ExecStmt(stmt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec.pipelines) != 1 {
		t.Fatalf("expected 1 pipeline, got %d", len(exec.pipelines))
	}
}

func TestExecStmtUnknownType(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	err := interp.ExecStmt(nil)
	if err == nil {
		t.Fatal("expected error for unknown statement type")
	}
}

func TestExecIfConditionTrue(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	s := &IfStmt{
		Condition: &parser.Pipeline{},
		Then:      []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}},
	}
	if err := interp.execIf(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// condition + then body
	if len(exec.pipelines) != 2 {
		t.Fatalf("expected 2 pipelines, got %d", len(exec.pipelines))
	}
}

func TestExecIfConditionFalseElse(t *testing.T) {
	exec := newFakeExec()
	// condition fails, else body succeeds
	exec.callErrors = []error{errors.New("fail"), nil}
	interp := NewInterpreter(exec)
	s := &IfStmt{
		Condition: &parser.Pipeline{},
		Else:      []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}},
	}
	if err := interp.execIf(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// condition + else body
	if len(exec.pipelines) != 2 {
		t.Fatalf("expected 2 pipelines, got %d", len(exec.pipelines))
	}
}

func TestExecIfElseIf(t *testing.T) {
	exec := newFakeExec()
	// condition fails, elif condition succeeds, elif body succeeds
	exec.callErrors = []error{errors.New("fail"), nil, nil}
	interp := NewInterpreter(exec)
	s := &IfStmt{
		Condition: &parser.Pipeline{},
		ElseIf: []*ElseIfStmt{
			{Condition: &parser.Pipeline{}, Body: []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}}},
		},
	}
	if err := interp.execIf(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// condition + elif condition + elif body
	if len(exec.pipelines) != 3 {
		t.Fatalf("expected 3 pipelines, got %d", len(exec.pipelines))
	}
}

func TestExecFor(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	s := &ForStmt{
		Var:   "x",
		Words: []string{"a", "b", "c"},
		Body:  []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}},
	}
	if err := interp.execFor(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exec.vars["x"] != "c" {
		t.Fatalf("expected last var value 'c', got %q", exec.vars["x"])
	}
	if len(exec.pipelines) != 3 {
		t.Fatalf("expected 3 body pipelines, got %d", len(exec.pipelines))
	}
}

func TestExecWhile(t *testing.T) {
	exec := newFakeExec()
	// call1=condition(nil), call2=body(nil), call3=condition(stop)
	exec.callErrors = []error{nil, nil, errors.New("stop")}
	interp := NewInterpreter(exec)
	s := &WhileStmt{
		Condition: &parser.Pipeline{},
		Body:      []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}},
	}
	if err := interp.execWhile(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec.pipelines) != 3 {
		t.Fatalf("expected 3 pipeline calls, got %d", len(exec.pipelines))
	}
}

func TestExecUntil(t *testing.T) {
	exec := newFakeExec()
	// call1=condition(fail), call2=body(nil), call3=condition(nil)
	exec.callErrors = []error{errors.New("fail"), nil, nil}
	interp := NewInterpreter(exec)
	s := &UntilStmt{
		Condition: &parser.Pipeline{},
		Body:      []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}},
	}
	if err := interp.execUntil(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec.pipelines) != 3 {
		t.Fatalf("expected 3 pipeline calls, got %d", len(exec.pipelines))
	}
}

func TestExecTryNoError(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	s := &TryStmt{
		Try:   []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}},
		Catch: []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}},
	}
	if err := interp.execTry(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// only try body ran
	if len(exec.pipelines) != 1 {
		t.Fatalf("expected 1 pipeline, got %d", len(exec.pipelines))
	}
}

func TestExecTryWithError(t *testing.T) {
	exec := newFakeExec()
	// try body fails, catch body succeeds
	exec.callErrors = []error{errors.New("boom"), nil}
	interp := NewInterpreter(exec)
	s := &TryStmt{
		Try:   []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}},
		Catch: []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}},
	}
	if err := interp.execTry(s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// try + catch body
	if len(exec.pipelines) != 2 {
		t.Fatalf("expected 2 pipelines, got %d", len(exec.pipelines))
	}
}

func TestExecStmtFuncDef(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	body := []Stmt{&CommandStmt{Pipeline: &parser.Pipeline{}}}
	if err := interp.ExecStmt(&FuncDef{Name: "foo", Body: body}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exec.funcs["foo"]) != 1 {
		t.Fatalf("expected func def stored, got %d", len(exec.funcs["foo"]))
	}
}

func TestExecLineEmpty(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	consumed, err := ExecLine("   ", interp)
	if err != nil || consumed {
		t.Fatalf("expected (false, nil), got (%v, %v)", consumed, err)
	}
}

func TestExecLineControlFlow(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	consumed, err := ExecLine("if true\nthen\n  echo hi\nfi", interp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !consumed {
		t.Fatal("expected control flow line to be consumed")
	}
}

func TestExecLinePlainCommand(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	consumed, err := ExecLine("echo hello", interp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if consumed {
		t.Fatal("expected plain command not to be consumed")
	}
}

func TestIsCommand(t *testing.T) {
	if IsCommand("echo") != true {
		t.Fatal("expected echo to be a command")
	}
	for _, kw := range []string{"if", "then", "else", "fi", "for", "in", "do", "done", "while", "until", "function"} {
		if IsCommand(kw) != false {
			t.Errorf("expected %q to not be a command", kw)
		}
	}
}

func TestParseAndExec(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	if err := interp.ParseAndExec("echo hi\n"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseAndExecInvalid(t *testing.T) {
	exec := newFakeExec()
	interp := NewInterpreter(exec)
	if err := interp.ParseAndExec("if then"); err == nil {
		t.Fatal("expected error for invalid script")
	}
}
