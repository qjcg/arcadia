package shell

import (
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"

	"github.com/chzyer/readline"
)

// replCommands are the REPL-only commands the engine does not serve.
var replCommands = []string{
	"alias", "unalias", "jobs", "fg", "bg", "history", "state",
}

// interpBuiltins are the bash builtins the engine provides. They
// cannot be discovered from a name alone, so completion carries the
// list (mirroring interp.IsBuiltin).
var interpBuiltins = []string{
	".", ":", "[", "alias", "bg", "bind", "break", "builtin", "caller",
	"cd", "command", "compgen", "complete", "compopt", "continue",
	"declare", "dirs", "disown", "echo", "enable", "eval", "exec",
	"exit", "export", "false", "fc", "fg", "getopts", "hash", "help",
	"history", "jobs", "kill", "let", "local", "logout", "mapfile",
	"popd", "printf", "pushd", "pwd", "read", "readarray", "readonly",
	"return", "set", "shift", "shopt", "source", "suspend", "test",
	"time", "times", "trap", "true", "type", "typeset", "ulimit",
	"umask", "unalias", "unset", "wait",
}

func (s *Shell) completer() readline.AutoCompleter {
	return &shellCompleter{shell: s}
}

// shellCompleter implements readline.AutoCompleter.
type shellCompleter struct {
	shell *Shell
}

func (c *shellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	lineStr := string(line[:pos])
	fullCandidates := c.shell.completeCommand(lineStr)
	if len(fullCandidates) == 0 {
		return nil, 0
	}

	// Find the last word being completed
	lastWord := lineStr
	if idx := strings.LastIndexAny(lineStr, " \t"); idx >= 0 {
		lastWord = lineStr[idx+1:]
	}

	// Find the common prefix of the last word and all candidates
	commonLen := len(lastWord)
	for _, cand := range fullCandidates {
		cl := commonPrefixLen(lastWord, cand)
		if cl < commonLen {
			commonLen = cl
		}
	}

	// Build suffixes: strip the common prefix from each candidate
	var suffixes [][]rune
	seen := make(map[string]bool)
	for _, cand := range fullCandidates {
		suffix := cand[commonLen:]
		if !seen[suffix] {
			seen[suffix] = true
			suffixes = append(suffixes, []rune(suffix))
		}
	}

	return suffixes, commonLen
}

// commonPrefixLen returns the length of the common prefix of a and b.
func commonPrefixLen(a, b string) int {
	max := min(len(b), len(a))
	for i := range max {
		if a[i] != b[i] {
			return i
		}
	}
	return max
}

// completeCommand is called by readline for dynamic completion.
func (s *Shell) completeCommand(line string) []string {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}

	// If we're completing the first word (command name)
	if len(line) == len(parts[0]) || (len(line) > len(parts[0]) && line[len(parts[0])] == ' ') {
		prefix := parts[0]
		// Don't complete if we already have a full command with space
		if len(line) > len(parts[0]) {
			return s.completeFileOrArg(parts, strings.TrimSpace(line[len(parts[0]):]))
		}
		// Try command name completion first
		cmds := s.completeCommandName(prefix)
		if len(cmds) > 0 {
			return cmds
		}
		// Fall back to file completion
		return s.completeFileOrArg(parts, prefix)
	}

	// Check if we're completing a partial word
	lastWord := parts[len(parts)-1]
	if !strings.HasSuffix(line, " ") {
		// We're completing the last word
		if len(parts) == 1 {
			return s.completeCommandName(lastWord)
		}
		return s.completeFileOrArg(parts, lastWord)
	}

	// Complete files after a space
	return s.completeFileOrArg(parts, "")
}

// completeCommandName offers REPL commands, shell builtins, aliases,
// functions, and executables from PATH.
func (s *Shell) completeCommandName(prefix string) []string {
	var candidates []string

	for _, name := range replCommands {
		if strings.HasPrefix(name, prefix) {
			candidates = append(candidates, name)
		}
	}
	for _, name := range interpBuiltins {
		if strings.HasPrefix(name, prefix) {
			candidates = append(candidates, name)
		}
	}
	for name := range s.aliases {
		if strings.HasPrefix(name, prefix) {
			candidates = append(candidates, name)
		}
	}
	for name := range s.engine.Funcs() {
		if strings.HasPrefix(name, prefix) {
			candidates = append(candidates, name)
		}
	}

	// PATH commands
	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if strings.HasPrefix(name, prefix) {
				fi, err := e.Info()
				if err != nil {
					continue
				}
				if fi.Mode()&0o111 != 0 {
					candidates = append(candidates, name)
				}
			}
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	unique := candidates[:0]
	for _, c := range candidates {
		if !seen[c] {
			seen[c] = true
			unique = append(unique, c)
		}
	}

	sort.Strings(unique)
	if len(unique) > 100 {
		unique = unique[:100]
	}
	return unique
}

// expandLeadingTilde expands a leading unquoted ~ (bare, ~/, or
// ~user) to the corresponding home directory. When no expansion
// applies, p is returned unchanged.
func (s *Shell) expandLeadingTilde(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	rest := p[1:]
	name := rest
	slash := ""
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		name = rest[:i]
		slash = rest[i:]
	}
	if name == "" {
		home := os.Getenv("HOME")
		if home == "" {
			return p
		}
		return home + slash
	}
	u, err := user.Lookup(name)
	if err != nil {
		return p
	}
	return u.HomeDir + slash
}

// completeFileOrArg completes filesystem paths for the word being
// typed, keeping the ~ form of the typed prefix.
func (s *Shell) completeFileOrArg(parts []string, lastWord string) []string {
	// A bare ~ completes to the home directory (with trailing slash),
	// matching bash.
	if lastWord == "~" {
		if home := s.expandLeadingTilde("~"); home != "~" {
			return []string{home + "/"}
		}
	}
	// Keep the original (unexpanded) directory prefix so completed
	// paths reuse the ~ form instead of an absolute home path.
	origDir := ""
	if idx := strings.LastIndexByte(lastWord, '/'); idx >= 0 {
		origDir = lastWord[:idx+1]
	}

	// Expand a leading tilde so the directory can be read.
	lastWord = s.expandLeadingTilde(lastWord)
	dir := "."
	filePrefix := lastWord

	// Check if the path contains a directory
	if idx := strings.LastIndexAny(lastWord, "/"); idx >= 0 {
		dir = lastWord[:idx]
		filePrefix = lastWord[idx+1:]
		if dir == "" {
			dir = "/"
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var candidates []string
	for _, e := range entries {
		name := e.Name()
		// Skip hidden files unless prefix starts with .
		if !strings.HasPrefix(filePrefix, ".") && strings.HasPrefix(name, ".") {
			continue
		}
		if strings.HasPrefix(name, filePrefix) {
			if e.IsDir() {
				name += "/"
			}
			// Always include the original directory prefix if there was one
			if origDir != "" {
				name = origDir + name
			}
			candidates = append(candidates, name)
		}
	}

	sort.Strings(candidates)
	if len(candidates) > 100 {
		candidates = candidates[:100]
	}
	return candidates
}
