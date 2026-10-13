package shell

// newTestShell builds a REPL shell for tests of the interactive
// machinery. The engine and job control come from New.
func newTestShell() *Shell {
	return New()
}
