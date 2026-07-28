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

func (s *Shell) completeFileOrArg(parts []string, lastWord string) []string {
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

// expandVars replaces $VAR, ${VAR}, $((...)), and $? with their values in the given string.
func (s *Shell) expandVars(input string) string {
	if !strings.ContainsRune(input, '$') && !strings.ContainsRune(input, '`') {
		return input
	}

	var result strings.Builder
	i := 0
	for i < len(input) {
		// Handle backtick command substitution
		if input[i] == '`' {
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
				result.WriteString(strings.TrimRight(output, "\n"))
			}
			continue
		}

		if input[i] != '$' {
			result.WriteByte(input[i])
			i++
			continue
		}

		i++ // skip $

		if i >= len(input) {
			result.WriteByte('$')
			break
		}

		ch := input[i]

		// $$
		if ch == '$' {
			result.WriteString(fmt.Sprintf("%d", os.Getpid()))
			i++
			continue
		}

		// $?
		if ch == '?' {
			result.WriteString(fmt.Sprintf("%d", s.exitCode))
			i++
			continue
		}

		// $((...)) arithmetic expansion
		if ch == '(' && i+1 < len(input) && input[i+1] == '(' {
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
			result.WriteString(fmt.Sprintf("%d", val))
			continue
		}

		// ${VAR} or ${name[idx]} or ${#name[@]}
		if ch == '{' {
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
					result.WriteString(val)
				} else {
					val := s.getVar(inner)
					result.WriteString(fmt.Sprintf("%d", len(val)))
				}
				continue
			}

			// String manipulation operations
			if processed := s.expandStringOp(inner, &result); processed {
				continue
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
						val := s.getArrayVar(name, "!"+index)
						result.WriteString(val)
					} else {
						val := s.getArrayVar(name, index)
						result.WriteString(val)
					}
					continue
				}
			}

			// Array name with [@] or [*]
			if strings.HasSuffix(inner, "[@]") || strings.HasSuffix(inner, "[*]") {
				name := strings.TrimSuffix(strings.TrimSuffix(inner, "[@]"), "[*]")
				val := s.getArrayVar(name, "@")
				result.WriteString(val)
				continue
			}

			// Regular variable
			result.WriteString(s.getVar(inner))
			continue
		}

		// $VAR - read alphanumeric/underscore name
		start := i
		for i < len(input) && (unicode.IsLetter(rune(input[i])) || unicode.IsDigit(rune(input[i])) || input[i] == '_') {
			i++
		}
		name := input[start:i]
		if name == "" {
			result.WriteByte('$')
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
			result.WriteString(s.getArrayVar(name, index))
			continue
		}

		result.WriteString(s.getVar(name))
	}

	return result.String()
}

// expandStringOp handles string manipulation operations in ${...}.
// Returns true if the operation was handled.
func (s *Shell) expandStringOp(inner string, result *strings.Builder) bool {
	// ${var:offset:length} - substring
	if before, after, ok := strings.Cut(inner, ":"); ok {
		// Only if ':' is not part of a valid variable name
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
			result.WriteString("")
		} else if offset+length >= len(val) {
			result.WriteString(val[offset:])
		} else {
			result.WriteString(val[offset : offset+length])
		}
		return true
	}

	// ${var#pattern} - remove shortest prefix
	if idx := strings.IndexByte(inner, '#'); idx >= 0 && idx > 0 {
		if idx+1 < len(inner) && inner[idx+1] == '#' {
			// ${var##pattern} - remove longest prefix
			name := inner[:idx]
			pattern := inner[idx+2:]
			val := s.getVar(name)
			for strings.HasPrefix(val, pattern) {
				val = val[len(pattern):]
			}
			result.WriteString(val)
			return true
		}
		name := inner[:idx]
		pattern := inner[idx+1:]
		val := s.getVar(name)
		if strings.HasPrefix(val, pattern) {
			val = val[len(pattern):]
		}
		result.WriteString(val)
		return true
	}

	// ${var%pattern} - remove shortest suffix
	if idx := strings.IndexByte(inner, '%'); idx >= 0 && idx > 0 {
		if idx+1 < len(inner) && inner[idx+1] == '%' {
			// ${var%%pattern} - remove longest suffix
			name := inner[:idx]
			pattern := inner[idx+2:]
			val := s.getVar(name)
			for strings.HasSuffix(val, pattern) {
				val = val[:len(val)-len(pattern)]
			}
			result.WriteString(val)
			return true
		}
		name := inner[:idx]
		pattern := inner[idx+1:]
		val := s.getVar(name)
		if strings.HasSuffix(val, pattern) {
			val = val[:len(val)-len(pattern)]
		}
		result.WriteString(val)
		return true
	}

	// ${var/pattern/replacement} - replace first
	if idx := strings.IndexByte(inner, '/'); idx >= 0 && idx > 0 {
		name := inner[:idx]
		rest := inner[idx+1:]
		if len(rest) > 0 && rest[0] == '/' {
			// ${var//pattern/replacement} - replace all
			parts := strings.SplitN(rest[1:], "/", 2)
			if len(parts) == 2 {
				val := s.getVar(name)
				val = strings.ReplaceAll(val, parts[0], parts[1])
				result.WriteString(val)
				return true
			}
		} else {
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) == 2 {
				val := s.getVar(name)
				val = strings.Replace(val, parts[0], parts[1], 1)
				result.WriteString(val)
				return true
			}
		}
	}

	// ${var^} - uppercase first, ${var^^} - uppercase all
	if strings.HasSuffix(inner, "^^") {
		name := inner[:len(inner)-2]
		val := s.getVar(name)
		if val != "" {
			result.WriteString(strings.ToUpper(val))
		}
		return true
	}
	if strings.HasSuffix(inner, "^") {
		name := inner[:len(inner)-1]
		val := s.getVar(name)
		if len(val) > 0 {
			val = strings.ToUpper(val[:1]) + val[1:]
		}
		result.WriteString(val)
		return true
	}

	// ${var,} - lowercase first, ${var,,} - lowercase all
	if strings.HasSuffix(inner, ",,") {
		name := inner[:len(inner)-2]
		val := s.getVar(name)
		if val != "" {
			result.WriteString(strings.ToLower(val))
		}
		return true
	}
	if strings.HasSuffix(inner, ",") {
		name := inner[:len(inner)-1]
		val := s.getVar(name)
		if len(val) > 0 {
			val = strings.ToLower(val[:1]) + val[1:]
		}
		result.WriteString(val)
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

	if !isIdent(name) {
		return false
	}
	if len(args) == 0 {
		return false
	}

	// Check for arr=( values...) pattern (with space after =()
	first := args[0]
	if strings.HasPrefix(first, "=(") {
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

	return false
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
