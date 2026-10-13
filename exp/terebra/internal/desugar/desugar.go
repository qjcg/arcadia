// Package desugar rewrites terebra's auger pipes into a reserved
// pipeline stage so that source code parses as pure bash.
//
// This is Mechanism A from ADR 0004. Under bash rules the auger pipe
// is invisible to the grammar, and worse: "cmd1 |> cmd2" parses
// silently as "cmd1 | >cmd2" (a pipe into a file-truncating redirect).
// So before any parsing happens, unquoted auger pipes are rewritten
// into a reserved "__terebra_auger__" stage that the shell treats as a
// command:
//
//	cmd1 |> cmd2    becomes  cmd1 | __terebra_auger__ cmd2
//	cmd1 |>json     becomes  cmd1 | __terebra_auger__ --enc=json
//
// The scanner is deliberately quote-aware only. It never needs to
// understand shell grammar: it skips quoted strings, command
// substitutions, parameter and arithmetic expansions, backticks,
// comments, and heredoc bodies, and rewrites only adjacent unquoted
// "|" and ">" pairs outside of those contexts. Everything it does not
// rewrite is copied through byte for byte.
package desugar

import "strings"

// AugerStage is the reserved command name the auger pipe is rewritten
// into. The shell recognizes it and performs CUE-aware piping.
const AugerStage = "__terebra_auger__"

// Source rewrites unquoted auger pipes into AugerStage pipeline
// stages. Sources without auger pipes are returned byte-identical.
func Source(src string) string {
	var b strings.Builder
	var pending []string // heredoc delimiters awaiting bodies

	i := 0
	atWordStart := true
	for i < len(src) {
		c := src[i]
		switch {
		case c == '\\':
			i = copyEscape(src, i, &b)
			atWordStart = false
		case c == '\'':
			i = copySingleQuoted(src, i, &b)
			atWordStart = false
		case c == '"':
			i = copyDoubleQuoted(src, i, &b)
			atWordStart = false
		case c == '`':
			i = copyBacktick(src, i, &b)
			atWordStart = false
		case c == '$':
			j, ok := copyExpansion(src, i, &b)
			if ok {
				i = j
				atWordStart = false
				continue
			}
			b.WriteByte(c)
			i++
		case c == '#' && atWordStart:
			i = copyComment(src, i, &b)
		case c == '<':
			j, delim, ok := scanHeredocOp(src, i)
			if ok {
				pending = append(pending, delim)
				b.WriteString(src[i:j])
				i = j
				atWordStart = true
				continue
			}
			i = copyOperator(src, i, &b, 1)
			atWordStart = true
		case c == '\n':
			b.WriteByte('\n')
			i++
			atWordStart = true
			for _, d := range pending {
				i = copyHeredocBody(src, i, &b, d)
			}
			pending = nil
		case c == '|':
			i = copyPipe(src, i, &b)
			atWordStart = true
		case c == ' ' || c == '\t' || c == ';' || c == '&' ||
			c == '(' || c == ')' || c == '>':
			b.WriteByte(c)
			i++
			atWordStart = true
		default:
			b.WriteByte(c)
			i++
			atWordStart = false
		}
	}
	return b.String()
}

// copyEscape copies a backslash escape pair verbatim.
func copyEscape(src string, i int, b *strings.Builder) int {
	end := min(i+2, len(src))
	b.WriteString(src[i:end])
	return end
}

// copySingleQuoted copies a single-quoted string verbatim. Nothing
// inside single quotes is special, not even backslashes.
func copySingleQuoted(src string, i int, b *strings.Builder) int {
	if j := strings.IndexByte(src[i+1:], '\''); j >= 0 {
		end := i + 1 + j + 1
		b.WriteString(src[i:end])
		return end
	}
	b.WriteString(src[i:])
	return len(src)
}

// copyDoubleQuoted copies a double-quoted string. Quotes inside are
// inert, but command substitutions, backticks, and parameter
// expansions are live contexts and are handled like anywhere else.
func copyDoubleQuoted(src string, i int, b *strings.Builder) int {
	b.WriteByte(src[i])
	i++
	for i < len(src) {
		switch src[i] {
		case '\\':
			i = copyEscape(src, i, b)
		case '"':
			b.WriteByte(src[i])
			return i + 1
		case '`':
			i = copyBacktick(src, i, b)
		case '$':
			j, ok := copyExpansion(src, i, b)
			if ok {
				i = j
				continue
			}
			b.WriteByte(src[i])
			i++
		default:
			b.WriteByte(src[i])
			i++
		}
	}
	return i
}

// copyBacktick copies a backtick command substitution, recursively
// rewriting its command context.
func copyBacktick(src string, i int, b *strings.Builder) int {
	end := skipBacktick(src, i)
	b.WriteByte('`')
	b.WriteString(Source(src[i+1 : end-1]))
	b.WriteByte('`')
	return end
}

// copyExpansion copies "$" expansions. Command substitution bodies
// are command contexts and are recursively rewritten; arithmetic and
// parameter expansions are copied verbatim. Reports whether it handled
// a bracketed form.
func copyExpansion(src string, i int, b *strings.Builder) (int, bool) {
	switch {
	case strings.HasPrefix(src[i:], "$(("):
		end := skipBalancedParens(src, i+2)
		b.WriteString(src[i:end])
		return end, true
	case strings.HasPrefix(src[i:], "$("):
		end := skipParenCmd(src, i)
		b.WriteString("$(")
		b.WriteString(Source(src[i+2 : end-1]))
		b.WriteByte(')')
		return end, true
	case strings.HasPrefix(src[i:], "${"):
		end := skipBrace(src, i)
		b.WriteString(src[i:end])
		return end, true
	}
	return i, false
}

// copyComment copies a comment through end of line verbatim.
func copyComment(src string, i int, b *strings.Builder) int {
	if j := strings.IndexByte(src[i:], '\n'); j >= 0 {
		b.WriteString(src[i : i+j])
		return i + j
	}
	b.WriteString(src[i:])
	return len(src)
}

// copyOperator copies n operator bytes verbatim.
func copyOperator(src string, i int, b *strings.Builder, n int) int {
	end := min(i+n, len(src))
	b.WriteString(src[i:end])
	return end
}

// copyPipe handles "|" and its multi-char forms, mirroring the old
// terebra lexer's precedence: "|&", "|>", "||", "|".
func copyPipe(src string, i int, b *strings.Builder) int {
	switch {
	case strings.HasPrefix(src[i:], "|&"):
		return copyOperator(src, i, b, 2)
	case strings.HasPrefix(src[i:], "|>"):
		return copyAuger(src, i, b)
	case strings.HasPrefix(src[i:], "||"):
		return copyOperator(src, i, b, 2)
	}
	return copyOperator(src, i, b, 1)
}

// copyAuger rewrites one auger pipe into the reserved auger stage.
// An encoder keyword (json, yaml, cue) directly after the pipe is
// recognized as the encoder form only when it ends the command; when
// followed by arguments it is the stage command itself.
func copyAuger(src string, i int, b *strings.Builder) int {
	j := i + 2
	k := j
	for k < len(src) && (src[k] == ' ' || src[k] == '\t') {
		k++
	}
	wordEnd := k
	for wordEnd < len(src) && !isWordBreak(src[wordEnd]) {
		wordEnd++
	}
	word := src[k:wordEnd]
	if isEncoder(word) && endsCommand(src, wordEnd) {
		b.WriteString("| " + AugerStage + " --enc=" + word)
		return wordEnd
	}
	b.WriteString("| " + AugerStage)
	if j < len(src) && src[j] != ' ' && src[j] != '\t' {
		b.WriteByte(' ')
	}
	return j
}

// isEncoder reports whether word auger-encodes its input.
func isEncoder(word string) bool {
	switch word {
	case "json", "yaml", "cue":
		return true
	}
	return false
}

// endsCommand reports whether the position right after a word is the
// end of its command: end of input, or a terminator possibly after
// horizontal whitespace. A "#" there starts a comment, so the command
// ended too.
func endsCommand(src string, i int) bool {
	for i < len(src) && (src[i] == ' ' || src[i] == '\t') {
		i++
	}
	if i >= len(src) {
		return true
	}
	switch src[i] {
	case '\n', ';', '&', '|', ')', '>', '<', '#':
		return true
	}
	return false
}

// isWordBreak reports whether a byte ends a shell word.
func isWordBreak(c byte) bool {
	switch c {
	case ' ', '\t', '\n', ';', '&', '|', '<', '>',
		'(', ')', '\'', '"', '`', '$', '\\':
		return true
	}
	return false
}

// scanHeredocOp recognizes a heredoc redirection operator and its
// delimiter word, returning the index just after the delimiter and
// the unquoted delimiter itself. Here-strings ("<<<") are not
// heredocs and report ok=false.
func scanHeredocOp(src string, i int) (int, string, bool) {
	if !strings.HasPrefix(src[i:], "<<") || strings.HasPrefix(src[i:], "<<<") {
		return i, "", false
	}
	j := i + 2
	if j < len(src) && src[j] == '-' {
		j++
	}
	for j < len(src) && (src[j] == ' ' || src[j] == '\t') {
		j++
	}
	end := j
	for end < len(src) && !isDelimBreak(src[end]) {
		end++
	}
	delim := strings.Map(func(r rune) rune {
		switch r {
		case '\'', '"', '`':
			return -1
		}
		return r
	}, src[j:end])
	if delim == "" {
		return i, "", false
	}
	return end, delim, true
}

// isDelimBreak reports whether a byte ends a heredoc delimiter word.
// Quote characters are part of the raw word and stripped afterwards,
// so quoted delimiters like <<'EOF' and <<E"OF" work.
func isDelimBreak(c byte) bool {
	switch c {
	case ' ', '\t', '\n', ';', '&', '|', '<', '>', '(', ')', '$', '\\':
		return true
	}
	return false
}

// copyHeredocBody copies one heredoc body verbatim through its
// terminator line. Bodies are never rewritten. Leading tabs are
// ignored when matching the terminator, covering the "<<-" form.
func copyHeredocBody(src string, i int, b *strings.Builder, delim string) int {
	for i < len(src) {
		nl := strings.IndexByte(src[i:], '\n')
		lineEnd := len(src)
		next := len(src)
		if nl >= 0 {
			lineEnd = i + nl
			next = lineEnd + 1
		}
		line := src[i:lineEnd]
		b.WriteString(src[i:next])
		i = next
		if line == delim || strings.TrimLeft(line, "\t") == delim {
			return i
		}
		if nl < 0 {
			return i
		}
	}
	return i
}

// skipBacktick returns the index just after the backtick closing the
// one at i.
func skipBacktick(src string, i int) int {
	j := i + 1
	for j < len(src) {
		switch src[j] {
		case '\\':
			j += 2
		case '`':
			return j + 1
		case '\'':
			j = skipSingleQuoted(src, j)
		case '"':
			j = skipDoubleQuoted(src, j)
		default:
			j++
		}
	}
	return len(src)
}

// skipSingleQuoted returns the index just after the single quote
// closing the one at i.
func skipSingleQuoted(src string, i int) int {
	if j := strings.IndexByte(src[i+1:], '\''); j >= 0 {
		return i + 1 + j + 1
	}
	return len(src)
}

// skipDoubleQuoted returns the index just after the double quote
// closing the one at i, stepping over nested live contexts.
func skipDoubleQuoted(src string, i int) int {
	j := i + 1
	for j < len(src) {
		switch {
		case src[j] == '\\':
			j += 2
		case src[j] == '"':
			return j + 1
		case src[j] == '`':
			j = skipBacktick(src, j)
		case src[j] == '$' && strings.HasPrefix(src[j:], "$(("):
			j = skipBalancedParens(src, j+2)
		case src[j] == '$' && strings.HasPrefix(src[j:], "$("):
			j = skipParenCmd(src, j)
		case src[j] == '$' && strings.HasPrefix(src[j:], "${"):
			j = skipBrace(src, j)
		default:
			j++
		}
	}
	return len(src)
}

// skipParenCmd returns the index just after the ")" closing the "$("
// command substitution at i.
func skipParenCmd(src string, i int) int {
	j := i + 2
	depth := 1
	for j < len(src) {
		switch {
		case src[j] == '\\':
			j += 2
		case src[j] == '\'':
			j = skipSingleQuoted(src, j)
		case src[j] == '"':
			j = skipDoubleQuoted(src, j)
		case src[j] == '`':
			j = skipBacktick(src, j)
		case strings.HasPrefix(src[j:], "$(("):
			j = skipBalancedParens(src, j+2)
		case strings.HasPrefix(src[j:], "$("):
			j = skipParenCmd(src, j)
		case strings.HasPrefix(src[j:], "${"):
			j = skipBrace(src, j)
		case src[j] == '(':
			depth++
			j++
		case src[j] == ')':
			depth--
			j++
			if depth == 0 {
				return j
			}
		default:
			j++
		}
	}
	return len(src)
}

// skipBrace returns the index just after the "}" closing the "${"
// parameter expansion at i.
func skipBrace(src string, i int) int {
	j := i + 2
	depth := 1
	for j < len(src) {
		switch {
		case src[j] == '\\':
			j += 2
		case src[j] == '\'':
			j = skipSingleQuoted(src, j)
		case src[j] == '"':
			j = skipDoubleQuoted(src, j)
		case strings.HasPrefix(src[j:], "${"):
			j = skipBrace(src, j)
		case src[j] == '{':
			depth++
			j++
		case src[j] == '}':
			depth--
			j++
			if depth == 0 {
				return j
			}
		default:
			j++
		}
	}
	return len(src)
}

// skipBalancedParens returns the index just after the ")" balancing
// the "(" at i, stepping over quoted regions.
func skipBalancedParens(src string, i int) int {
	j := i
	depth := 0
	for j < len(src) {
		switch {
		case src[j] == '\\':
			j += 2
		case src[j] == '\'':
			j = skipSingleQuoted(src, j)
		case src[j] == '"':
			j = skipDoubleQuoted(src, j)
		case src[j] == '`':
			j = skipBacktick(src, j)
		case src[j] == '(':
			depth++
			j++
		case src[j] == ')':
			depth--
			j++
			if depth == 0 {
				return j
			}
		default:
			j++
		}
	}
	return len(src)
}
