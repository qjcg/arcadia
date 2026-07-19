package script

import (
	"fmt"
	"slices"
	"strings"

	"github.com/qjcg/arcadia/exp/terebra/internal/parser"
)

// Parse parses a script string into a Script AST.
func Parse(input string) (*Script, error) {
	p := &parser_{}
	return p.parse(strings.TrimSpace(input))
}

type parser_ struct {
	input string
	pos   int
}

func (p *parser_) parse(input string) (*Script, error) {
	p.input = input
	p.pos = 0

	script := &Script{}
	for p.pos < len(p.input) {
		p.skipEmptyLines()
		if p.pos >= len(p.input) {
			break
		}

		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		script.Stmts = append(script.Stmts, stmt)
	}
	// Remove nil statements (comments, blank lines)
	filtered := make([]Stmt, 0, len(script.Stmts))
	for _, s := range script.Stmts {
		if s != nil {
			filtered = append(filtered, s)
		}
	}
	script.Stmts = filtered

	return script, nil
}

func (p *parser_) parseStmt() (Stmt, error) {
	line := p.readLine()
	line = strings.TrimSpace(line)

	if line == "" {
		return nil, nil
	}

	// Skip comments
	if strings.HasPrefix(line, "#") {
		return nil, nil
	}

	words := splitWords(line)
	if len(words) == 0 {
		return nil, nil
	}

	keyword := words[0]

	switch keyword {
	case "if":
		return p.parseIf(line, words)
	case "for":
		return p.parseFor(line, words)
	case "while":
		return p.parseWhile(line, words)
	case "until":
		return p.parseUntil(line, words)
	case "function":
		return p.parseFuncDef(line, words)
	case "try":
		return p.parseTry(line, words)
	case "fi", "then", "else", "elif", "do", "done", "end", "catch":
		return nil, fmt.Errorf("unexpected keyword %q", keyword)
	default:
		// Check for function definition with () — e.g., "hello() {"
		if strings.HasSuffix(keyword, "()") && len(words) >= 2 && (words[1] == "{" || words[1] == "()") {
			name := strings.TrimSuffix(keyword, "()")
			if words[1] == "()" {
				// "hello () {" or "hello ()"
				// Combine back
				return p.parseFuncDefAlt(line, append([]string{name, "()"}, words[2:]...))
			}
			return p.parseFuncDefAlt(line, append([]string{name, "()"}, words[1:]...))
		}
		// Regular command
		return p.parseCommand(line)
	}
}

// parseIf parses if/then/elif/else/fi blocks.
func (p *parser_) parseIf(firstLine string, words []string) (*IfStmt, error) {
	// Parse condition: "if <command>; then" or "if <command>\nthen"
	condition, err := p.parseCondition(words[1:], firstLine, []string{"then"})
	if err != nil {
		return nil, fmt.Errorf("if: %v", err)
	}

	stmt := &IfStmt{Condition: condition}

	// Parse then body
	p.skipEmptyLines()
	thenBody, err := p.parseBody([]string{"elif", "else", "fi"})
	if err != nil {
		return nil, fmt.Errorf("if: %v", err)
	}
	stmt.Then = thenBody

	// Parse elif/else/fi
	for {
		p.skipEmptyLines()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("if: expected fi")
		}

		nextLine := p.readLine()
		nextLine = strings.TrimSpace(nextLine)
		nextWords := splitWords(nextLine)

		if len(nextWords) == 0 {
			continue
		}

		switch nextWords[0] {
		case "elif":
			cond, err := p.parseCondition(nextWords[1:], nextLine, []string{"then"})
			if err != nil {
				return nil, fmt.Errorf("elif: %v", err)
			}
			ei := &ElseIfStmt{Condition: cond}
			p.skipEmptyLines()
			body, err := p.parseBody([]string{"elif", "else", "fi"})
			if err != nil {
				return nil, fmt.Errorf("elif: %v", err)
			}
			ei.Body = body
			stmt.ElseIf = append(stmt.ElseIf, ei)

		case "else":
			p.skipEmptyLines()
			elseBody, err := p.parseBody([]string{"fi"})
			if err != nil {
				return nil, fmt.Errorf("else: %v", err)
			}
			stmt.Else = elseBody

		case "fi":
			return stmt, nil

		default:
			return nil, fmt.Errorf("if: expected elif, else, or fi, got %q", nextWords[0])
		}
	}
}

// parseFor parses for/in/do/done loops.
func (p *parser_) parseFor(firstLine string, words []string) (*ForStmt, error) {
	// "for var in word1 word2 ..." or "for var in word1 word2 ...\ndo"
	if len(words) < 3 || words[2] != "in" {
		return nil, fmt.Errorf("for: expected 'for var in ...'")
	}

	stmt := &ForStmt{Var: words[1]}

	// Parse words after "in" until "do" or end of line
	rest := words[3:]
	wordsUntilDo, err := p.parseWordsUntil(firstLine, rest, []string{"do"})
	if err != nil {
		return nil, fmt.Errorf("for: %v", err)
	}
	stmt.Words = wordsUntilDo

	// Parse body until "done"
	p.skipEmptyLines()
	body, err := p.parseBody([]string{"done"})
	if err != nil {
		return nil, fmt.Errorf("for: %v", err)
	}
	stmt.Body = body

	// Consume "done"
	p.skipEmptyLines()
	p.consumeKeyword("done")

	return stmt, nil
}

// parseWhile parses while/do/done loops.
func (p *parser_) parseWhile(firstLine string, words []string) (*WhileStmt, error) {
	condition, err := p.parseCondition(words[1:], firstLine, []string{"do"})
	if err != nil {
		return nil, fmt.Errorf("while: %v", err)
	}

	stmt := &WhileStmt{Condition: condition}

	p.skipEmptyLines()
	body, err := p.parseBody([]string{"done"})
	if err != nil {
		return nil, fmt.Errorf("while: %v", err)
	}
	stmt.Body = body

	// Consume "done"
	p.skipEmptyLines()
	p.consumeKeyword("done")

	return stmt, nil
}

// parseUntil parses until/do/done loops.
func (p *parser_) parseUntil(firstLine string, words []string) (*UntilStmt, error) {
	condition, err := p.parseCondition(words[1:], firstLine, []string{"do"})
	if err != nil {
		return nil, fmt.Errorf("until: %v", err)
	}

	stmt := &UntilStmt{Condition: condition}

	p.skipEmptyLines()
	body, err := p.parseBody([]string{"done"})
	if err != nil {
		return nil, fmt.Errorf("until: %v", err)
	}
	stmt.Body = body

	// Consume "done"
	p.skipEmptyLines()
	p.consumeKeyword("done")

	return stmt, nil
}

// parseFuncDef parses "function name { ... }".
func (p *parser_) parseFuncDef(firstLine string, words []string) (*FuncDef, error) {
	if len(words) < 2 {
		return nil, fmt.Errorf("function: expected name")
	}

	name := words[1]
	stmt := &FuncDef{Name: name}

	// Check for "{" on the same line or next line
	rest := words[2:]
	if len(rest) == 0 || rest[0] != "{" {
		p.skipEmptyLines()
		nextLine := p.readLine()
		nextLine = strings.TrimSpace(nextLine)
		nextWords := splitWords(nextLine)
		if len(nextWords) > 0 && nextWords[0] == "{" {
			rest = nextWords[1:]
		} else {
			return nil, fmt.Errorf("function: expected {")
		}
	} else {
		rest = rest[1:]
	}

	body, err := p.parseBody([]string{"}"})
	if err != nil {
		return nil, fmt.Errorf("function: %v", err)
	}
	stmt.Body = body

	// Consume "}"
	p.skipEmptyLines()
	p.consumeKeyword("}")

	return stmt, nil
}

// parseTry parses try/catch/end blocks.
func (p *parser_) parseTry(firstLine string, words []string) (*TryStmt, error) {
	stmt := &TryStmt{}

	// Parse try body until catch or end
	p.skipEmptyLines()
	tryBody, err := p.parseBody([]string{"catch", "end"})
	if err != nil {
		return nil, fmt.Errorf("try: %v", err)
	}
	stmt.Try = tryBody

	// Check for catch
	p.skipEmptyLines()
	if p.pos < len(p.input) {
		savedPos := p.pos
		nextLine := p.readLine()
		nextLine = strings.TrimSpace(nextLine)
		nextWords := splitWords(nextLine)
		if len(nextWords) > 0 && nextWords[0] == "catch" {
			// Parse catch body
			catchBody, err := p.parseBody([]string{"end"})
			if err != nil {
				return nil, fmt.Errorf("catch: %v", err)
			}
			stmt.Catch = catchBody
		} else {
			p.pos = savedPos
		}
	}

	// Consume "end"
	p.skipEmptyLines()
	p.consumeKeyword("end")

	return stmt, nil
}

// parseFuncDefAlt parses "name() { ... }".
func (p *parser_) parseFuncDefAlt(line string, words []string) (*FuncDef, error) {
	name := words[0]
	stmt := &FuncDef{Name: name}

	rest := words[2:]
	if len(rest) == 0 || rest[0] != "{" {
		p.skipEmptyLines()
		nextLine := p.readLine()
		nextLine = strings.TrimSpace(nextLine)
		nextWords := splitWords(nextLine)
		if len(nextWords) > 0 && nextWords[0] == "{" {
			// Consume the opening brace
		} else {
			return nil, fmt.Errorf("function %s: expected {", name)
		}
	}

	body, err := p.parseBody([]string{"}"})
	if err != nil {
		return nil, fmt.Errorf("function %s: %v", name, err)
	}
	stmt.Body = body

	// Consume "}"
	p.skipEmptyLines()
	p.consumeKeyword("}")

	return stmt, nil
}

// consumeKeyword reads and discards the next line if it equals the expected keyword.
func (p *parser_) consumeKeyword(keyword string) {
	if p.pos >= len(p.input) {
		return
	}
	savedPos := p.pos
	line := p.readLine()
	line = strings.TrimSpace(line)
	if line != keyword {
		p.pos = savedPos
	}
}

// parseCommand parses a single command line.
func (p *parser_) parseCommand(line string) (*CommandStmt, error) {
	pipe, err := parser.Parse(line)
	if err != nil {
		return nil, err
	}
	return &CommandStmt{Pipeline: pipe}, nil
}

// parseCondition parses a condition that ends with a terminator keyword (e.g., "then", "do").
// The condition may span multiple lines.
func (p *parser_) parseCondition(restWords []string, firstLine string, terminators []string) (*parser.Pipeline, error) {
	// Build the condition text from restWords and subsequent lines
	var conditionWords []string
	conditionWords = append(conditionWords, restWords...)

	// Check if the terminator is in the restWords
	for _, term := range terminators {
		for i, w := range conditionWords {
			if w == term {
				conditionWords = conditionWords[:i]
				// Check for ";" before terminator
				if len(conditionWords) > 0 && conditionWords[len(conditionWords)-1] == ";" {
					conditionWords = conditionWords[:len(conditionWords)-1]
				}
				cond := strings.Join(conditionWords, " ")
				if strings.TrimSpace(cond) == "" {
					return nil, fmt.Errorf("empty condition")
				}
				return parser.Parse(cond)
			}
		}
	}

	// Terminator not on this line, read more lines
	for {
		p.skipEmptyLines()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("expected %s", strings.Join(terminators, " or "))
		}

		nextLine := p.readLine()
		nextLine = strings.TrimSpace(nextLine)
		nextWords := splitWords(nextLine)

		if len(nextWords) == 0 {
			continue
		}

		// Check for terminators
		for _, term := range terminators {
			for i, w := range nextWords {
				if w == term {
					conditionWords = append(conditionWords, nextWords[:i]...)
					cond := strings.Join(conditionWords, " ")
					if strings.TrimSpace(cond) == "" {
						return nil, fmt.Errorf("empty condition")
					}
					return parser.Parse(cond)
				}
			}
		}

		conditionWords = append(conditionWords, nextWords...)
	}
}

// parseBody parses statements until one of the terminators is encountered.
func (p *parser_) parseBody(terminators []string) ([]Stmt, error) {
	var body []Stmt

	for {
		p.skipEmptyLines()
		if p.pos >= len(p.input) {
			return nil, fmt.Errorf("expected %s", strings.Join(terminators, " or "))
		}

		// Peek at the next line to check for terminators
		savedPos := p.pos
		nextLine := p.readLine()
		nextLine = strings.TrimSpace(nextLine)
		nextWords := splitWords(nextLine)

		isTerminator := false
		for _, w := range nextWords {
			if slices.Contains(terminators, w) {
				isTerminator = true
			}
			if isTerminator {
				break
			}
		}

		if isTerminator {
			// Put back the line
			p.pos = savedPos
			return body, nil
		}

		// Not a terminator, parse as a statement
		p.pos = savedPos // reset to parse properly
		stmt, err := p.parseStmt()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			body = append(body, stmt)
		}
	}
}

// parseWordsUntil collects words until a terminator is found.
func (p *parser_) parseWordsUntil(firstLine string, words []string, terminators []string) ([]string, error) {
	var result []string
	result = append(result, words...)

	// Check for terminator in current words
	for _, term := range terminators {
		for i, w := range result {
			if w == term {
				return result[:i], nil
			}
		}
	}

	// Read more lines
	for {
		p.skipEmptyLines()
		if p.pos >= len(p.input) {
			return result, nil
		}

		nextLine := p.readLine()
		nextLine = strings.TrimSpace(nextLine)
		nextWords := splitWords(nextLine)

		if len(nextWords) == 0 {
			continue
		}

		for _, term := range terminators {
			for i, w := range nextWords {
				if w == term {
					result = append(result, nextWords[:i]...)
					return result, nil
				}
			}
		}

		result = append(result, nextWords...)
	}
}

// readLine reads the next line from the input.
func (p *parser_) readLine() string {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '\n' {
		p.pos++
	}
	line := p.input[start:p.pos]
	if p.pos < len(p.input) {
		p.pos++ // skip newline
	}
	return line
}

// skipEmptyLines skips empty lines and lines with only whitespace.
func (p *parser_) skipEmptyLines() {
	for p.pos < len(p.input) {
		start := p.pos
		for p.pos < len(p.input) && p.input[p.pos] != '\n' {
			p.pos++
		}
		line := strings.TrimSpace(p.input[start:p.pos])
		if line != "" {
			p.pos = start
			return
		}
		if p.pos < len(p.input) {
			p.pos++ // skip newline
		}
	}
}

// splitWords splits a line into words, handling basic quoting.
func splitWords(line string) []string {
	return strings.Fields(line)
}

// ParseScript parses a script from a file path.
func ParseScript(path string) (*Script, error) {
	// TODO: read file and parse
	return nil, fmt.Errorf("not implemented")
}
