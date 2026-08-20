package shell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"cuelang.org/go/cue"
	"github.com/chzyer/readline"
	"github.com/qjcg/arcadia/exp/terebra/internal/cueutil"
	"github.com/qjcg/arcadia/exp/terebra/internal/expand"
	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
	"github.com/qjcg/arcadia/exp/terebra/internal/script"
)

// completer returns a readline.AutoCompleter for tab completion.
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

func (s *Shell) completeCommandName(prefix string) []string {
	var candidates []string

	// Built-in commands
	for _, name := range s.builtins.Names() {
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
				// Check if executable
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

// expandLeadingTilde expands a leading unquoted ~ (bare, ~/, or ~user) in p to
// the corresponding home directory, returning the expanded path. When no
// expansion applies (empty $HOME or unknown user) p is returned unchanged.
func (s *Shell) expandLeadingTilde(p string) string {
	if !strings.HasPrefix(p, "~") {
		return p
	}
	return expand.Tilde(expand.Word{Value: p}).Value
}

func (s *Shell) completeFileOrArg(parts []string, lastWord string) []string {
	// A bare ~ completes to the home directory (with trailing slash), matching
	// bash. ~/ and ~user/... are expanded before directory listing.
	if lastWord == "~" {
		if home := s.expandLeadingTilde("~"); home != "~" {
			return []string{home + "/"}
		}
	}
	// Expand a leading tilde (~, ~/, ~user) to the real path so the directory
	// can be read and the completed path is directly usable.
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
			// Always include the directory prefix if the original path had one
			if dir != "." || strings.Contains(lastWord, "/") {
				name = dir + "/" + name
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

// maskBuilder builds a string alongside a per-byte quoted mask.
type maskBuilder struct {
	str  strings.Builder
	mask []bool
}

func (m *maskBuilder) WriteStringQuoted(s string, quoted bool) {
	m.str.WriteString(s)
	for range s {
		m.mask = append(m.mask, quoted)
	}
}

func (m *maskBuilder) WriteByteQuoted(b byte, quoted bool) {
	m.str.WriteByte(b)
	m.mask = append(m.mask, quoted)
}

func (m *maskBuilder) String() string { return m.str.String() }

func (m *maskBuilder) Mask() []bool { return m.mask }

func (m *maskBuilder) Reset() {
	m.str.Reset()
	m.mask = nil
}

// maskAt returns the quoted flag for byte i of input, or false if mask is nil
// or i is out of range.
func maskAt(mask []bool, i int) bool {
	if mask == nil || i < 0 || i >= len(mask) {
		return false
	}
	return mask[i]
}

// expandVars replaces $VAR, ${VAR}, $((...)), and $? with their values in the
// given string, propagating the quoted mask. Bytes produced by a variable
// expansion inherit the quotedness of the '$' that introduced them.
func (s *Shell) expandVars(input string, mask []bool) (string, []bool) {
	if !strings.ContainsRune(input, '$') && !strings.ContainsRune(input, '`') {
		return input, mask
	}

	var result maskBuilder
	i := 0
	for i < len(input) {
		// Handle backtick command substitution
		if input[i] == '`' {
			i = s.expandBacktick(input, i, &result, maskAt(mask, i))
			continue
		}

		if input[i] != '$' {
			result.WriteByteQuoted(input[i], maskAt(mask, i))
			i++
			continue
		}

		i++ // skip $
		quoted := maskAt(mask, i-1)

		if i >= len(input) {
			result.WriteByteQuoted('$', quoted)
			break
		}

		ch := input[i]

		// $$
		if ch == '$' {
			result.WriteStringQuoted(fmt.Sprintf("%d", os.Getpid()), quoted)
			i++
			continue
		}

		// $?
		if ch == '?' {
			result.WriteStringQuoted(fmt.Sprintf("%d", s.exitCode), quoted)
			i++
			continue
		}

		// $((...)) arithmetic expansion
		if ch == '(' && i+1 < len(input) && input[i+1] == '(' {
			i = s.expandArithmetic(input, i, &result, quoted)
			continue
		}

		// ${VAR} or ${name[idx]} or ${#name[@]}
		if ch == '{' {
			i = s.expandBraced(input, i, &result, quoted)
			continue
		}

		// $VAR - read alphanumeric/underscore name
		start := i
		for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i])) || input[i] == '_') {
			i++
		}
		name := input[start:i]
		if name == "" {
			result.WriteByteQuoted('$', quoted)
			continue
		}

		// Check for $arr[idx] array access (without braces)
		if i < len(input) && input[i] == '[' {
			i++ // skip [
			idxStart := i
			for i < len(input) && input[i] != ']' {
				i++
			}
			index := input[idxStart:i]
			if i < len(input) {
				i++ // skip ]
			}
			result.WriteStringQuoted(s.getArrayVar(name, index), quoted)
			continue
		}

		result.WriteStringQuoted(s.getVar(name), quoted)
	}

	return result.String(), result.Mask()
}

// expandBacktick handles a backtick command substitution starting at input[i]
// (which points at the opening backtick). It returns the new index.
func (s *Shell) expandBacktick(input string, i int, result *maskBuilder, quoted bool) int {
	i++ // skip opening backtick
	start := i
	for i < len(input) && input[i] != '`' {
		i++
	}
	cmdStr := input[start:i]
	if i < len(input) {
		i++ // skip closing backtick
	}
	// Execute the command and capture output
	output, err := s.captureCommandOutput(cmdStr)
	if err == nil {
		result.WriteStringQuoted(strings.TrimRight(output, "\n"), quoted)
	}
	return i
}

// expandArithmetic handles a $((...)) arithmetic expansion starting at input[i]
// (which points at the first '(' of the double paren). It returns the new index.
func (s *Shell) expandArithmetic(input string, i int, result *maskBuilder, quoted bool) int {
	i += 2 // skip ((
	start := i
	// depth starts at 2 (for the two (( that were skipped)
	// so we need two )) to close
	depth := 2
	for i < len(input) && depth > 0 {
		if input[i] == '(' {
			depth++
		} else if input[i] == ')' {
			depth--
		}
		if depth > 0 {
			i++
		}
	}
	expr := input[start:i]
	if i < len(input) {
		i++ // skip the final )
	}
	val := s.evalArithmetic(expr)
	result.WriteStringQuoted(fmt.Sprintf("%d", val), quoted)
	return i
}

// expandBraced handles a ${...} expansion starting at input[i] (which points
// at the '{'). It returns the new index.
func (s *Shell) expandBraced(input string, i int, result *maskBuilder, quoted bool) int {
	i++ // skip {
	start := i

	// Check for ${#...} (length prefix)
	isLength := false
	if i < len(input) && input[i] == '#' {
		isLength = true
		i++
		start = i
	}

	// Read until matching }
	depth := 1
	for i < len(input) && depth > 0 {
		if input[i] == '{' {
			depth++
		} else if input[i] == '}' {
			depth--
		}
		if depth > 0 {
			i++
		}
	}
	inner := input[start:i]
	if i < len(input) {
		i++ // skip }
	}

	// Parse inner: name[idx] or name or string operations
	if isLength {
		// ${#var} - length
		if strings.HasSuffix(inner, "[@]") || strings.HasSuffix(inner, "[*]") {
			name := strings.TrimSuffix(strings.TrimSuffix(inner, "[@]"), "[*]")
			val := s.getArrayVar(name, "#")
			result.WriteStringQuoted(val, quoted)
		} else {
			val := s.getVar(inner)
			result.WriteStringQuoted(fmt.Sprintf("%d", len(val)), quoted)
		}
		return i
	}

	// String manipulation operations
	if processed := s.expandStringOp(inner, result, quoted); processed {
		return i
	}

	// Array access: name[idx] or !name[idx] (list keys)
	if before, after, ok := strings.Cut(inner, "["); ok {
		name := before
		// Strip ! prefix for key listing
		listKeys := false
		if strings.HasPrefix(name, "!") {
			listKeys = true
			name = strings.TrimPrefix(name, "!")
		}
		rest := after
		before, _, ok := strings.Cut(rest, "]")
		if ok {
			index := before
			if listKeys {
				result.WriteStringQuoted(s.getArrayVar(name, "!"+index), quoted)
			} else {
				result.WriteStringQuoted(s.getArrayVar(name, index), quoted)
			}
			return i
		}
	}

	// Array name with [@] or [*]
	if strings.HasSuffix(inner, "[@]") || strings.HasSuffix(inner, "[*]") {
		name := strings.TrimSuffix(strings.TrimSuffix(inner, "[@]"), "[*]")
		result.WriteStringQuoted(s.getArrayVar(name, "@"), quoted)
		return i
	}

	// Regular variable
	result.WriteStringQuoted(s.getVar(inner), quoted)
	return i
}

// expandStringOp handles string manipulation operations in ${...}.
// Returns true if the operation was handled.
func (s *Shell) expandStringOp(inner string, result *maskBuilder, quoted bool) bool {
	if s.expandSubstring(inner, result, quoted) {
		return true
	}
	if s.expandRemovePrefix(inner, result, quoted) {
		return true
	}
	if s.expandRemoveSuffix(inner, result, quoted) {
		return true
	}
	if s.expandReplace(inner, result, quoted) {
		return true
	}
	if s.expandCase(inner, result, quoted) {
		return true
	}
	return false
}

// expandSubstring handles ${var:offset:length}.
func (s *Shell) expandSubstring(inner string, result *maskBuilder, quoted bool) bool {
	before, after, ok := strings.Cut(inner, ":")
	if !ok {
		return false
	}
	name := before
	rest := after
	val := s.getVar(name)
	parts := strings.SplitN(rest, ":", 2)
	offset := 0
	length := len(val)
	if len(parts[0]) > 0 {
		fmt.Sscanf(parts[0], "%d", &offset)
	}
	if offset < 0 {
		offset = len(val) + offset
	}
	if offset < 0 {
		offset = 0
	}
	if len(parts) > 1 && len(parts[1]) > 0 {
		fmt.Sscanf(parts[1], "%d", &length)
	}
	if offset >= len(val) {
		result.WriteStringQuoted("", quoted)
	} else if offset+length >= len(val) {
		result.WriteStringQuoted(val[offset:], quoted)
	} else {
		result.WriteStringQuoted(val[offset:offset+length], quoted)
	}
	return true
}

// expandRemovePrefix handles ${var#pattern} and ${var##pattern}.
func (s *Shell) expandRemovePrefix(inner string, result *maskBuilder, quoted bool) bool {
	idx := strings.IndexByte(inner, '#')
	if idx < 0 || idx == 0 {
		return false
	}
	if idx+1 < len(inner) && inner[idx+1] == '#' {
		// ${var##pattern} - remove longest prefix
		name := inner[:idx]
		pattern := inner[idx+2:]
		val := s.getVar(name)
		for strings.HasPrefix(val, pattern) {
			val = val[len(pattern):]
		}
		result.WriteStringQuoted(val, quoted)
		return true
	}
	name := inner[:idx]
	pattern := inner[idx+1:]
	val := s.getVar(name)
	if strings.HasPrefix(val, pattern) {
		val = val[len(pattern):]
	}
	result.WriteStringQuoted(val, quoted)
	return true
}

// expandRemoveSuffix handles ${var%pattern} and ${var%%pattern}.
func (s *Shell) expandRemoveSuffix(inner string, result *maskBuilder, quoted bool) bool {
	idx := strings.IndexByte(inner, '%')
	if idx < 0 || idx == 0 {
		return false
	}
	if idx+1 < len(inner) && inner[idx+1] == '%' {
		// ${var%%pattern} - remove longest suffix
		name := inner[:idx]
		pattern := inner[idx+2:]
		val := s.getVar(name)
		for strings.HasSuffix(val, pattern) {
			val = val[:len(val)-len(pattern)]
		}
		result.WriteStringQuoted(val, quoted)
		return true
	}
	name := inner[:idx]
	pattern := inner[idx+1:]
	val := s.getVar(name)
	if strings.HasSuffix(val, pattern) {
		val = val[:len(val)-len(pattern)]
	}
	result.WriteStringQuoted(val, quoted)
	return true
}

// expandReplace handles ${var/pattern/replacement} and ${var//pattern/replacement}.
func (s *Shell) expandReplace(inner string, result *maskBuilder, quoted bool) bool {
	idx := strings.IndexByte(inner, '/')
	if idx < 0 || idx == 0 {
		return false
	}
	name := inner[:idx]
	rest := inner[idx+1:]
	if len(rest) > 0 && rest[0] == '/' {
		// ${var//pattern/replacement} - replace all
		parts := strings.SplitN(rest[1:], "/", 2)
		if len(parts) == 2 {
			val := s.getVar(name)
			val = strings.ReplaceAll(val, parts[0], parts[1])
			result.WriteStringQuoted(val, quoted)
			return true
		}
	} else {
		parts := strings.SplitN(rest, "/", 2)
		if len(parts) == 2 {
			val := s.getVar(name)
			val = strings.Replace(val, parts[0], parts[1], 1)
			result.WriteStringQuoted(val, quoted)
			return true
		}
	}
	return false
}

// expandCase handles ${var^}, ${var^^}, ${var,}, and ${var,,}.
func (s *Shell) expandCase(inner string, result *maskBuilder, quoted bool) bool {
	if strings.HasSuffix(inner, "^^") {
		name := inner[:len(inner)-2]
		val := s.getVar(name)
		if val != "" {
			result.WriteStringQuoted(strings.ToUpper(val), quoted)
		}
		return true
	}
	if strings.HasSuffix(inner, "^") {
		name := inner[:len(inner)-1]
		val := s.getVar(name)
		if len(val) > 0 {
			val = strings.ToUpper(val[:1]) + val[1:]
		}
		result.WriteStringQuoted(val, quoted)
		return true
	}
	if strings.HasSuffix(inner, ",,") {
		name := inner[:len(inner)-2]
		val := s.getVar(name)
		if val != "" {
			result.WriteStringQuoted(strings.ToLower(val), quoted)
		}
		return true
	}
	if strings.HasSuffix(inner, ",") {
		name := inner[:len(inner)-1]
		val := s.getVar(name)
		if len(val) > 0 {
			val = strings.ToLower(val[:1]) + val[1:]
		}
		result.WriteStringQuoted(val, quoted)
		return true
	}
	return false
}

// evalArithmetic evaluates a simple arithmetic expression.
func (s *Shell) evalArithmetic(expr string) int {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0
	}
	val, _ := s.evalExpr(expr)
	return val
}

// evalExpr evaluates an arithmetic expression (handles + and -).
func (s *Shell) evalExpr(expr string) (int, string) {
	val, rest := s.evalTerm(expr)
	rest = strings.TrimSpace(rest)
	for len(rest) > 0 {
		op := rest[0]
		if op != '+' && op != '-' {
			break
		}
		right, r := s.evalTerm(rest[1:])
		if op == '+' {
			val += right
		} else {
			val -= right
		}
		rest = strings.TrimSpace(r)
	}
	return val, rest
}

// evalTerm evaluates a term (handles *, /, %).
func (s *Shell) evalTerm(expr string) (int, string) {
	val, rest := s.evalFactor(expr)
	rest = strings.TrimSpace(rest)
	for len(rest) > 0 {
		op := rest[0]
		if op != '*' && op != '/' && op != '%' {
			break
		}
		right, r := s.evalFactor(rest[1:])
		if op == '*' {
			val *= right
		} else if op == '/' {
			if right == 0 {
				val = 0
			} else {
				val /= right
			}
		} else {
			if right == 0 {
				val = 0
			} else {
				val %= right
			}
		}
		rest = strings.TrimSpace(r)
	}
	return val, rest
}

// evalFactor evaluates a factor: number, variable, parenthesized expression, or unary +/-.
func (s *Shell) evalFactor(expr string) (int, string) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, ""
	}

	ch := expr[0]

	// Unary minus
	if ch == '-' {
		val, rest := s.evalFactor(expr[1:])
		return -val, rest
	}

	// Unary plus
	if ch == '+' {
		return s.evalFactor(expr[1:])
	}

	// Parenthesized expression
	if ch == '(' {
		val, rest := s.evalExpr(expr[1:])
		rest = strings.TrimSpace(rest)
		if len(rest) > 0 && rest[0] == ')' {
			rest = rest[1:]
		}
		return val, rest
	}

	// Variable reference
	if unicode.IsLetter(rune(ch)) || ch == '_' {
		end := 1
		for end < len(expr) && (unicode.IsLetter(rune(expr[end])) || unicode.IsDigit(rune(expr[end])) || expr[end] == '_') {
			end++
		}
		name := expr[:end]
		valStr := s.getVar(name)
		val := 0
		fmt.Sscanf(valStr, "%d", &val)
		return val, expr[end:]
	}

	// Number
	end := 0
	for end < len(expr) && unicode.IsDigit(rune(expr[end])) {
		end++
	}
	if end == 0 {
		return 0, expr[1:]
	}
	val := 0
	fmt.Sscanf(expr[:end], "%d", &val)
	return val, expr[end:]
}

// getVar returns the value of a variable, checking local vars first, then environment.
func (s *Shell) getVar(name string) string {
	if val, ok := s.vars[name]; ok {
		return val
	}
	return os.Getenv(name)
}

// getArrayVar returns an array element or the whole array.
// name[idx] returns the element at index, name[@] returns all elements as space-separated.
func (s *Shell) getArrayVar(name, idx string) string {
	if idx == "@" || idx == "*" {
		// Return all elements
		if arr, ok := s.arrays[name]; ok {
			return strings.Join(arr, " ")
		}
		if m, ok := s.assoc[name]; ok {
			var vals []string
			for _, v := range m {
				vals = append(vals, v)
			}
			return strings.Join(vals, " ")
		}
		return ""
	}

	// ${!arr[@]} or ${!arr[*]} — list keys
	if idx == "!@" || idx == "!*" {
		if m, ok := s.assoc[name]; ok {
			var keys []string
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return strings.Join(keys, " ")
		}
		return ""
	}

	// Indexed array
	if arr, ok := s.arrays[name]; ok {
		if idx == "#" {
			return fmt.Sprintf("%d", len(arr))
		}
		i := 0
		fmt.Sscanf(idx, "%d", &i)
		if i >= 0 && i < len(arr) {
			return arr[i]
		}
		return ""
	}

	// Associative array
	if m, ok := s.assoc[name]; ok {
		if idx == "#" {
			return fmt.Sprintf("%d", len(m))
		}
		return m[idx]
	}

	return ""
}

// setArray sets an indexed array variable.
func (s *Shell) setArray(name string, values []string) {
	s.arrays[name] = values
}

// tryArrayAssignment checks if the command is an array assignment like arr=(1 2 3).
// Returns true if it was handled as an array assignment.
func (s *Shell) tryArrayAssignment(name string, args []string) bool {
	// Check for arr[idx]=value pattern (single word, no args)
	if len(args) == 0 && strings.Contains(name, "[") && strings.Contains(name, "]") && strings.Contains(name, "=") {
		return s.parseArrayElemAssign(name)
	}

	// Check for arr=(...) pattern where name is "arr=(..." and args contain the rest
	// The parser splits "arr=(1 2 3)" into name="arr=(1", args=["2", "3)"]
	// Also handles arr=([key]=val ...) for associative arrays
	if strings.Contains(name, "=(") {
		return s.parseArrayAssign(name, args)
	}

	if !isIdent(name) {
		return false
	}
	if len(args) == 0 {
		return false
	}

	// Check for arr=( values...) pattern (with space after =()
	first := args[0]
	if strings.HasPrefix(first, "=(") {
		return s.parseArrayAssignSpaced(name, args)
	}

	return false
}

// parseArrayAssign handles the arr=(...) pattern where the opening paren is
// part of the command name (e.g. name="arr=(1", args=["2", "3)"]).
func (s *Shell) parseArrayAssign(name string, args []string) bool {
	before, after, _ := strings.Cut(name, "=(")
	arrName := before
	if !isIdent(arrName) {
		return false
	}
	// The first value might be part of the name after =(
	firstVal := after
	var values []string
	if firstVal != "" {
		values = append(values, firstVal)
	}
	for _, arg := range args {
		v := strings.TrimSuffix(arg, ")")
		values = append(values, v)
	}
	// Remove trailing empty string from last element if it ended with )
	if len(values) > 0 && strings.HasSuffix(args[len(args)-1], ")") {
		last := values[len(values)-1]
		if last == "" {
			values = values[:len(values)-1]
		}
	}
	// Check if values look like [key]=value (associative array)
	isAssoc := false
	for _, v := range values {
		if strings.HasPrefix(v, "[") && strings.Contains(v, "]=") {
			isAssoc = true
			break
		}
	}
	if isAssoc {
		m := make(map[string]string)
		for _, v := range values {
			if after0, ok := strings.CutPrefix(v, "["); ok {
				// [key]=value
				rest := after0
				if before, after, ok := strings.Cut(rest, "]="); ok {
					m[before] = after
				}
			}
		}
		s.assoc[arrName] = m
		delete(s.arrays, arrName)
	} else {
		s.setArray(arrName, values)
	}
	return true
}

// parseArrayAssignSpaced handles the arr=( values...) pattern where the
// opening paren is the first argument (e.g. args[0]="=(").
func (s *Shell) parseArrayAssignSpaced(name string, args []string) bool {
	first := args[0]
	var values []string
	if first == "=(" {
		for _, arg := range args[1:] {
			v := strings.TrimSuffix(arg, ")")
			values = append(values, v)
		}
	} else {
		val := strings.TrimPrefix(first, "=(")
		values = append(values, val)
		for _, arg := range args[1:] {
			v := strings.TrimSuffix(arg, ")")
			values = append(values, v)
		}
	}
	if len(values) > 0 && strings.HasSuffix(args[len(args)-1], ")") {
		last := values[len(values)-1]
		if last == "" {
			values = values[:len(values)-1]
		}
	}
	s.setArray(name, values)
	return true
}

// parseArrayElemAssign parses arr[idx]=value from a command name.
func (s *Shell) parseArrayElemAssign(name string) bool {
	before, after, ok := strings.Cut(name, "=")
	if !ok {
		return false
	}
	lhs := before
	rhs := after
	bracketStart := strings.IndexByte(lhs, '[')
	if bracketStart < 0 {
		return false
	}
	arrName := lhs[:bracketStart]
	index := lhs[bracketStart+1 : len(lhs)-1]
	if !isIdent(arrName) {
		return false
	}
	s.setArrayElem(arrName, index, rhs)
	return true
}

// isIdent checks if a string is a valid identifier.
func isIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}

func (s *Shell) setArrayElem(name, idx, value string) {
	// If already an associative array, use it
	if _, ok := s.assoc[name]; ok {
		s.assoc[name][idx] = value
		return
	}

	// Check if index is numeric — if not, auto-promote to associative array
	if _, err := strconv.Atoi(idx); err != nil {
		if _, ok := s.assoc[name]; !ok {
			s.assoc[name] = make(map[string]string)
		}
		// Remove from indexed arrays if present
		delete(s.arrays, name)
		s.assoc[name][idx] = value
		return
	}

	// Indexed array access
	arr := s.arrays[name]
	if arr == nil {
		arr = make([]string, 0)
	}
	i := 0
	fmt.Sscanf(idx, "%d", &i)
	for len(arr) <= i {
		arr = append(arr, "")
	}
	arr[i] = value
	s.arrays[name] = arr
}

// setVar sets a local shell variable.
func (s *Shell) setVar(name, value string) {
	// Check if variable is readonly
	if s.readonly[name] {
		return
	}
	s.vars[name] = value
}

// exportVar exports a local variable to the environment.
func (s *Shell) exportVar(name string) {
	if val, ok := s.vars[name]; ok {
		os.Setenv(name, val)
	}
}

// unsetVar removes a variable (both local and environment).
func (s *Shell) unsetVar(name string) {
	delete(s.vars, name)
	os.Unsetenv(name)
}

// listVars returns all local variables sorted by name.
func (s *Shell) listVars() []string {
	var result []string
	for k, v := range s.vars {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(result)
	return result
}

// builtinExport handles the export builtin with shell awareness.
func (s *Shell) builtinExport(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		// List all exported variables
		for _, env := range os.Environ() {
			fmt.Fprintln(stdout, env)
		}
		return 0
	}
	for _, arg := range args {
		if before, after, ok := strings.Cut(arg, "="); ok {
			name := before
			value := after
			s.setVar(name, value)
			s.exportVar(name)
		} else {
			// Export existing local var to environment
			s.exportVar(arg)
		}
	}
	return 0
}

// builtinUnset handles the unset builtin with shell awareness.
func (s *Shell) builtinUnset(args []string, stdout, stderr io.Writer) int {
	for _, arg := range args {
		s.unsetVar(arg)
	}
	return 0
}

// builtinSet handles the set builtin with shell awareness.
func (s *Shell) builtinSet(args []string, stdout, stderr io.Writer) int {
	if len(args) > 0 {
		// Handle set -x (debug mode) and set -o vi/emacs
		for _, arg := range args {
			switch arg {
			case "-x":
				s.debug = true
				return 0
			case "+x":
				s.debug = false
				return 0
			case "-o":
				// set -o option
				continue
			case "nullglob":
				s.nullGlob = true
				return 0
			case "dotglob":
				s.dotGlob = true
				return 0
			case "+o":
				// set +o option
				continue
			case "no-nullglob":
				s.nullGlob = false
				return 0
			case "no-dotglob":
				s.dotGlob = false
				return 0
			case "vi":
				if s.rl != nil {
					s.rl.SetVimMode(true)
				}
				return 0
			case "emacs":
				if s.rl != nil {
					s.rl.SetVimMode(false)
				}
				return 0
			default:
				if strings.HasPrefix(arg, "-") {
					fmt.Fprintf(stderr, "set: unknown option %s\n", arg)
					return 1
				}
			}
		}
		return 0
	}
	// List local variables first
	for _, v := range s.listVars() {
		fmt.Fprintln(stdout, v)
	}
	// Then list environment variables
	for _, env := range os.Environ() {
		fmt.Fprintln(stdout, env)
	}
	return 0
}

// builtinState exports shell state as CUE.
func (s *Shell) builtinState(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	// Subcommands: save, load
	if len(args) > 0 {
		switch args[0] {
		case "save":
			return s.builtinStateSave(args[1:], stdout, stderr)
		case "load":
			return s.builtinStateLoad(args[1:], stdout, stderr)
		}
	}

	ctx := cueutil.NewContext()

	// Build state as a Go map
	state := map[string]any{
		"vars":     s.vars,
		"exitCode": s.exitCode,
		"version":  "0.1.0",
	}

	v := cueutil.EncodeGo(ctx, state)
	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "state: %v\n", err)
		return 1
	}

	// Add schema annotation
	stateSchema := `#State: {
	vars: [string]: string
	exitCode: int
	version: "0.1.0"
}`

	schemaV := cueutil.CompileString(ctx, stateSchema)
	if err := cueutil.Err(schemaV); err == nil {
		v = cueutil.Unify(v, schemaV)
	}

	str, err := cueutil.FormatValue(v)
	if err != nil {
		fmt.Fprintf(stderr, "state: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout, str)
	return 0
}

// builtinStateSave saves shell state to a file.
func (s *Shell) builtinStateSave(args []string, stdout, stderr io.Writer) int {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	stateDir := filepath.Join(os.Getenv("HOME"), ".terebra", "states")
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		fmt.Fprintf(stderr, "state save: %v\n", err)
		return 1
	}

	path := filepath.Join(stateDir, name+".cue")
	// Generate state as CUE
	ctx := cueutil.NewContext()
	state := map[string]any{
		"vars":     s.vars,
		"exitCode": s.exitCode,
		"version":  "0.1.0",
		"pwd":      os.Getenv("PWD"),
	}

	v := cueutil.EncodeGo(ctx, state)
	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "state save: %v\n", err)
		return 1
	}

	str, err := cueutil.FormatValue(v)
	if err != nil {
		fmt.Fprintf(stderr, "state save: %v\n", err)
		return 1
	}

	if err := os.WriteFile(path, []byte(str+"\n"), 0o644); err != nil {
		fmt.Fprintf(stderr, "state save: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "state saved to %s\n", path)
	return 0
}

// builtinStateLoad loads shell state from a file.
func (s *Shell) builtinStateLoad(args []string, stdout, stderr io.Writer) int {
	name := "default"
	if len(args) > 0 {
		name = args[0]
	}

	stateDir := filepath.Join(os.Getenv("HOME"), ".terebra", "states")
	path := filepath.Join(stateDir, name+".cue")

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(stderr, "state load: %v\n", err)
		return 1
	}

	ctx := cueutil.NewContext()
	v := cueutil.CompileBytes(ctx, data)
	if err := cueutil.Err(v); err != nil {
		fmt.Fprintf(stderr, "state load: %v\n", err)
		return 1
	}

	// Extract vars from the state
	cueutil.WalkFields(v, func(name string, val cue.Value) bool {
		if name == "vars" {
			cueutil.WalkFields(val, func(fieldName string, fieldVal cue.Value) bool {
				raw, err := cueutil.FormatValueRaw(fieldVal)
				if err == nil && raw != "" {
					s.setVar(fieldName, strings.TrimSpace(raw))
				}
				return true
			})
		}
		return true
	})

	fmt.Fprintf(stdout, "state loaded from %s\n", path)
	return 0
}

// Shell methods implementing the script.Executor interface.

func (s *Shell) RunCommand(name string, args []string, stdin io.Reader, stdout io.Writer) error {
	cmd := &parser.Command{
		Name: name,
		Args: args,
	}
	return s.ExecuteCommand(cmd, stdin, stdout, nil)
}

func (s *Shell) FuncDefs() map[string][]script.Stmt {
	return s.funcs
}

func (s *Shell) SetFuncDef(name string, body []script.Stmt) {
	s.funcs[name] = body
}

// GetVar returns the value of a variable.
func (s *Shell) GetVar(name string) string {
	return s.getVar(name)
}

// SetVar sets a local shell variable.
func (s *Shell) SetVar(name, value string) {
	s.setVar(name, value)
}

// builtinSource executes a script file.
func (s *Shell) builtinSource(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "source: expected filename")
		return 1
	}

	data, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(stderr, "source: %v\n", err)
		return 1
	}

	if err := s.interp.ParseAndExec(string(data)); err != nil {
		fmt.Fprintf(stderr, "source: %v\n", err)
		return 1
	}
	return 0
}

// captureCommandOutput executes a command string and returns its stdout.
func (s *Shell) captureCommandOutput(cmdStr string) (string, error) {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return "", nil
	}
	path, err := exec.LookPath(parts[0])
	if err != nil {
		return "", err
	}
	cmd := exec.Command(path, parts[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// builtinAlias handles the alias builtin.
func (s *Shell) builtinAlias(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		// List all aliases
		names := make([]string, 0, len(s.aliases))
		for n := range s.aliases {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			fmt.Fprintf(stdout, "%s=%s\n", n, s.aliases[n])
		}
		return 0
	}
	for _, arg := range args {
		if before, after, ok := strings.Cut(arg, "="); ok {
			s.aliases[before] = after
		} else {
			// Show alias value
			if val, ok := s.aliases[arg]; ok {
				fmt.Fprintf(stdout, "%s=%s\n", arg, val)
			} else {
				fmt.Fprintf(stderr, "alias: %s: not found\n", arg)
				return 1
			}
		}
	}
	return 0
}

// builtinUnalias handles the unalias builtin.
func (s *Shell) builtinUnalias(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "unalias: expected alias name")
		return 1
	}
	for _, arg := range args {
		delete(s.aliases, arg)
	}
	return 0
}

// builtinHistory handles the history builtin.
func (s *Shell) builtinHistory(args []string, stdout, stderr io.Writer) int {
	// Read history from the history file
	home := os.Getenv("HOME")
	if home == "" {
		return 0
	}
	histPath := filepath.Join(home, historyFile)
	data, err := os.ReadFile(histPath)
	if err != nil {
		return 0
	}
	lines := strings.Split(string(data), "\n")
	// Remove trailing empty line
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	n := 0
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &n)
	}
	start := 0
	if n > 0 && len(lines) > n {
		start = len(lines) - n
	}
	for i := start; i < len(lines); i++ {
		fmt.Fprintf(stdout, "%5d  %s\n", i+1, lines[i])
	}
	return 0
}

// builtinReadonly handles the readonly builtin.
func (s *Shell) builtinReadonly(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		// List all readonly variables
		names := make([]string, 0, len(s.readonly))
		for n := range s.readonly {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			val := s.getVar(n)
			fmt.Fprintf(stdout, "%s=%s\n", n, val)
		}
		return 0
	}
	for _, arg := range args {
		if before, after, ok := strings.Cut(arg, "="); ok {
			s.setVar(before, after)
			s.readonly[before] = true
		} else {
			s.readonly[arg] = true
		}
	}
	return 0
}
