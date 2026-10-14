// Package engine runs terebra source on the mvdan/sh interp engine.
//
// This is the execution core from ADR 0004: bash semantics come from
// mvdan/sh verbatim, and terebra contributes only what mvdan/sh cannot
// provide (see docs/builtin-diff.md):
//
//   - the auger stage (__terebra_auger__), the desugared form of |> ;
//   - the product commands drill, cue, plugin, and state;
//   - externally supplied command execution (the REPL's job control).
//
// Evaluation pipeline: strip shebang, desugar auger pipes, parse as
// bash, brace-expand words, run on a persistent interp.Runner.
package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	mexpand "mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"

	"github.com/qjcg/arcadia/exp/terebra/internal/builtins"
	"github.com/qjcg/arcadia/exp/terebra/internal/cueutil"
	"github.com/qjcg/arcadia/exp/terebra/internal/desugar"
	"github.com/qjcg/arcadia/exp/terebra/internal/drill"
	"github.com/qjcg/arcadia/exp/terebra/internal/plugin"
)

// Options configures an Engine. The zero value uses the process
// standard streams and plain foreground command execution.
type Options struct {
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer

	// External overrides how external commands execute. The REPL
	// supplies its job-control handler here; nil means plain
	// foreground execution via interp's default handler.
	External interp.ExecHandlerFunc
}

// Engine evaluates terebra source on one persistent interpreter.
type Engine struct {
	runner   *interp.Runner
	plugins  *plugin.Registry
	external interp.ExecHandlerFunc
	out      *switchWriter
	bgCalls  atomic.Pointer[map[syntax.Pos]bool]
}

// New builds an engine with its own persistent interpreter.
func New(opts Options) (*Engine, error) {
	if opts.Stdin == nil {
		opts.Stdin = os.Stdin
	}
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	e := &Engine{plugins: plugin.New(), external: opts.External}
	e.out = &switchWriter{dst: opts.Stdout}
	runner, err := interp.New(
		interp.StdIO(opts.Stdin, e.out, opts.Stderr),
		interp.Env(mexpand.ListEnviron(os.Environ()...)),
		interp.ExecHandlers(func(next interp.ExecHandlerFunc) interp.ExecHandlerFunc {
			return e.exec
		}),
	)
	if err != nil {
		return nil, err
	}
	e.runner = runner
	return e, nil
}

// Eval runs one chunk of terebra source to completion.
func (e *Engine) Eval(ctx context.Context, src string) error {
	src = stripShebang(src)
	src = desugar.Source(src)
	prog, err := syntax.NewParser().Parse(strings.NewReader(src), "")
	if err != nil {
		return err
	}
	if err := expandBraces(prog); err != nil {
		return err
	}
	// Record which calls belong to background (&) statements so the
	// job layer can tell foreground jobs (which take the terminal)
	// from background ones (which keep SIGTTIN/SIGTTOU semantics).
	bg := map[syntax.Pos]bool{}
	for _, stmt := range prog.Stmts {
		if stmt.Background {
			syntax.Walk(stmt, func(n syntax.Node) bool {
				if c, ok := n.(*syntax.CallExpr); ok {
					bg[c.Pos()] = true
				}
				return true
			})
		}
	}
	e.bgCalls.Store(&bg)
	return e.runner.Run(ctx, prog)
}

// OutFileIfDefault returns the terminal file behind the engine's
// default stdout when w is that default -- nothing redirected or
// captured. Job handlers pass the file itself to children so isatty
// and terminal ioctls work: ls columns and colors, and interactive
// tools, all key off them. Pipelines and redirects keep their
// writers, where single-column output is bash-correct.
func (e *Engine) OutFileIfDefault(w io.Writer) *os.File {
	if e.out == nil || w != io.Writer(e.out) {
		return nil
	}
	e.out.mu.Lock()
	defer e.out.mu.Unlock()
	f, _ := e.out.dst.(*os.File)
	return f
}

// IsBackground reports whether the command at pos runs inside a
// background (&) statement of the most recent Eval.
func (e *Engine) IsBackground(pos syntax.Pos) bool {
	if m := e.bgCalls.Load(); m != nil {
		return (*m)[pos]
	}
	return false
}

// ExitCode maps an Eval error to a process exit code. The bool
// reports whether err was an explicit exit status, which callers
// should honor without printing an error of their own.
func ExitCode(err error) (int, bool) {
	if err == nil {
		return 0, true
	}
	if code, ok := interp.IsExitStatus(err); ok {
		return int(code), true
	}
	return 1, false
}

// RunScript executes a script file on a throwaway engine.
func RunScript(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot execute %q: %v", path, err)
	}
	return RunScriptFromString(string(data))
}

// RunScriptFromString executes script source on a throwaway engine.
func RunScriptFromString(content string) error {
	e, err := New(Options{})
	if err != nil {
		return err
	}
	return e.Eval(context.Background(), content)
}

// exec is the interp exec handler: the terebra residue runs in
// process, external commands go to the configured handler.
func (e *Engine) exec(ctx context.Context, args []string) error {
	hc := interp.HandlerCtx(ctx)
	code := e.dispatch(ctx, hc, args, hc.Stdin, hc.Stdout, hc.Stderr)
	if code == 0 {
		return nil
	}
	return error(interp.ExitStatus(uint8(code)))
}

// dispatch routes one command, with explicit stdio so the auger stage
// can feed its command formatted input. Returns an exit code.
func (e *Engine) dispatch(
	ctx context.Context,
	hc interp.HandlerContext,
	args []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) int {
	switch args[0] {
	case desugar.AugerStage:
		return e.auger(hc, args[1:], stdin, stdout, stderr)
	case "drill":
		return drill.Handler(args[1:], stdin, stdout, stderr)
	case "cue":
		return builtins.CueHandler()(args[1:], stdin, stdout, stderr)
	case "plugin":
		return plugin.PluginBuiltin(e.plugins)(args[1:], stdin, stdout, stderr)
	case "state":
		return e.stateBuiltin(hc, args[1:], stdin, stdout, stderr)
	}
	var err error
	if _, lookErr := exec.LookPath(args[0]); lookErr != nil {
		// Bash wording: scripts grep for "command not found".
		fmt.Fprintf(stderr, "terebra: %s: command not found\n", args[0])
		return 127
	}
	if e.external != nil {
		err = e.external(ctx, args)
	} else {
		err = interp.DefaultExecHandler(2*time.Second)(ctx, args)
	}
	if err == nil {
		return 0
	}
	if code, ok := interp.IsExitStatus(err); ok {
		return int(code)
	}
	fmt.Fprintln(stderr, err)
	return 1
}

// auger runs the desugared auger pipe stage. With an --enc flag it
// encodes its input; otherwise it formats its input as CUE when valid
// and runs the command with that on stdin (raw passthrough otherwise).
func (e *Engine) auger(
	hc interp.HandlerContext,
	args []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "auger: missing command or encoder")
		return 1
	}
	in, err := io.ReadAll(stdin)
	if err != nil {
		fmt.Fprintf(stderr, "auger: %v\n", err)
		return 1
	}
	trimmed := strings.TrimSpace(string(in))

	if enc, ok := strings.CutPrefix(args[0], "--enc="); ok {
		return e.augerEncode(enc, trimmed, stdout, stderr)
	}

	content := trimmed
	if trimmed != "" {
		ctx := cueutil.NewContext()
		v := cueutil.CompileString(ctx, trimmed)
		if cueutil.Err(v) == nil {
			if formatted, err := cueutil.FormatValue(v); err == nil {
				content = formatted
			}
		}
	}
	return e.runStageCommand(hc, args,
		strings.NewReader(content), stdout, stderr)
}

// augerEncode writes stdin encoded as one of the auger formats.
func (e *Engine) augerEncode(enc, trimmed string, stdout, stderr io.Writer) int {
	if trimmed == "" {
		return 0
	}
	ctx := cueutil.NewContext()
	v := cueutil.CompileString(ctx, trimmed)
	if cueutil.Err(v) != nil {
		fmt.Fprint(stdout, trimmed)
		return 0
	}
	switch enc {
	case "json":
		b, err := cueutil.ToJSON(v)
		if err != nil {
			fmt.Fprint(stdout, trimmed)
			return 0
		}
		fmt.Fprintln(stdout, string(b))
	case "yaml", "cue":
		formatted, err := cueutil.FormatValue(v)
		if err != nil {
			fmt.Fprint(stdout, trimmed)
			return 0
		}
		fmt.Fprintln(stdout, formatted)
	default:
		fmt.Fprintf(stderr, "auger: unknown encoder %q\n", enc)
		return 1
	}
	return 0
}

// runStageCommand runs the auger stage's command with explicit stdin.
func (e *Engine) runStageCommand(
	hc interp.HandlerContext,
	args []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) int {
	switch args[0] {
	case desugar.AugerStage, "drill", "cue", "plugin", "state":
		return e.dispatch(context.Background(), hc, args, stdin, stdout, stderr)
	}
	cmd := exec.Command(args[0], args[1:]...)
	if f := e.OutFileIfDefault(stdout); f != nil {
		stdout = f
	}
	cmd.Stdin, cmd.Stdout, cmd.Stderr = stdin, stdout, stderr
	cmd.Dir = hc.Dir
	cmd.Env = envPairs(hc.Env)
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode()
		}
		fmt.Fprintf(stderr, "%s: %v\n", args[0], err)
		return 127
	}
	return 0
}

// stateBuiltin renders interpreter state as CUE. Saving works;
// loading shell state into a running interpreter is owned by the REPL
// layer and reports clearly until then.
func (e *Engine) stateBuiltin(
	hc interp.HandlerContext,
	args []string,
	stdin io.Reader,
	stdout, stderr io.Writer,
) int {
	if len(args) > 0 && args[0] == "load" {
		fmt.Fprintln(stderr,
			"state: load is not supported by the interp engine yet")
		return 1
	}
	vars := map[string]string{}
	hc.Env.Each(func(name string, vr mexpand.Variable) bool {
		vars[name] = vr.String()
		return true
	})
	ctx := cueutil.NewContext()
	state := map[string]any{
		"vars":     vars,
		"exitCode": 0,
		"version":  "0.1.0",
	}
	v := cueutil.EncodeGo(ctx, state)
	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "state: %v\n", err)
		return 1
	}
	schema := `#State: {
	vars: [string]: string
	exitCode: int
	version: "0.1.0"
}`
	schemaV := cueutil.CompileString(ctx, schema)
	if err := cueutil.Err(schemaV); err == nil {
		v = cueutil.Unify(v, schemaV)
	}
	str, err := cueutil.FormatValue(v)
	if err != nil {
		fmt.Fprintf(stderr, "state: %v\n", err)
		return 1
	}
	out := str
	if len(args) > 0 && args[0] == "save" {
		path := "terebra-state.cue"
		if len(args) > 1 {
			path = args[1]
		}
		if err := os.WriteFile(path, []byte(str+"\n"), 0o644); err != nil {
			fmt.Fprintf(stderr, "state: %v\n", err)
			return 1
		}
		return 0
	}
	fmt.Fprintln(stdout, out)
	return 0
}

// envPairs renders the interpreter's exported variables as
// NAME=VALUE pairs for child processes.
func envPairs(env mexpand.Environ) []string {
	var pairs []string
	env.Each(func(name string, vr mexpand.Variable) bool {
		if vr.Exported {
			pairs = append(pairs, name+"="+vr.String())
		}
		return true
	})
	return pairs
}

// expandBraces rewrites every command word through brace expansion,
// the expansion step bash runs first. Quoted parts are untouched.
func expandBraces(file *syntax.File) error {
	var calls []*syntax.CallExpr
	syntax.Walk(file, func(n syntax.Node) bool {
		if c, ok := n.(*syntax.CallExpr); ok {
			calls = append(calls, c)
		}
		return true
	})
	for _, c := range calls {
		args, err := expandWords(c.Args)
		if err != nil {
			return err
		}
		c.Args = args
	}
	return nil
}

func expandWords(words []*syntax.Word) ([]*syntax.Word, error) {
	var out []*syntax.Word
	for _, w := range words {
		if !syntax.SplitBraces(w) {
			out = append(out, w)
			continue
		}
		expanded := false
		for bw, err := range mexpand.BracesSeq(&mexpand.Config{}, w) {
			if err != nil {
				return nil, err
			}
			out = append(out, bw)
			expanded = true
		}
		if !expanded {
			out = append(out, w)
		}
	}
	return out, nil
}

// stripShebang drops a leading interpreter line.
func stripShebang(src string) string {
	if !strings.HasPrefix(src, "#!") {
		return src
	}
	if _, after, ok := strings.Cut(src, "\n"); ok {
		return after
	}
	return ""
}

// --- interpreter introspection and capture ---

// Vars returns the interpreter's live variable table. Assignments
// made by Eval are reflected in it.
func (e *Engine) Vars() map[string]mexpand.Variable {
	return e.runner.Vars
}

// SetVars assigns variables directly into the interpreter. REPL-layer
// features (state load) use this; scripts should use assignments.
func (e *Engine) SetVars(vals map[string]string) {
	for k, v := range vals {
		e.runner.Vars[k] = mexpand.Variable{
			Set:  true,
			Kind: mexpand.String,
			Str:  v,
		}
	}
}

// Funcs returns the interpreter's live function table.
func (e *Engine) Funcs() map[string]*syntax.Stmt {
	return e.runner.Funcs
}

// Exited reports whether the last Eval ran the exit builtin.
func (e *Engine) Exited() bool {
	return e.runner.Exited()
}

// Capture evaluates src with stdout redirected, returning whatever
// the source wrote there. Used for PS1 command substitutions and
// other REPL queries.
func (e *Engine) Capture(ctx context.Context, src string) (string, error) {
	var buf bytes.Buffer
	restore := e.out.swap(&buf)
	defer restore()
	err := e.Eval(ctx, src)
	return buf.String(), err
}

// Incomplete reports whether err says the input was unfinished (an
// unclosed quote or open block), so a REPL can read more lines.
func Incomplete(err error) bool {
	var pe *syntax.ParseError
	return errors.As(err, &pe) && pe.Incomplete
}

// switchWriter routes interpreter output so Capture can temporarily
// divert stdout into a buffer.
type switchWriter struct {
	mu  sync.Mutex
	dst io.Writer
}

func (w *switchWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.dst.Write(p)
}

func (w *switchWriter) swap(dst io.Writer) func() {
	w.mu.Lock()
	defer w.mu.Unlock()
	prev := w.dst
	w.dst = dst
	return func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		w.dst = prev
	}
}
