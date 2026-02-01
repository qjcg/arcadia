package parser

import (
	"fmt"
	"strconv"

	"elbereth/internal/ast"
	"elbereth/internal/lexer"
)

// Parser parses Elbereth source code into an AST
type Parser struct {
	lex     *lexer.Lexer
	current lexer.Token
	peeked  lexer.Token
	hasPeek bool
	errors  []string
}

// New creates a new parser
func New(lex *lexer.Lexer) *Parser {
	p := &Parser{
		lex:     lex,
		errors:  []string{},
		hasPeek: false,
	}
	p.advance()
	return p
}

// Parse parses the entire program
func (p *Parser) Parse() (*ast.Program, error) {
	items := []ast.Node{}

	for p.current.Type != lexer.TokenEOF {
		item := p.parseTopLevel()
		if item != nil {
			items = append(items, item)
		}
		if len(p.errors) > 0 {
			break
		}
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.errors)
	}

	return &ast.Program{Items: items}, nil
}

// ============================================================================
// Top-level
// ============================================================================

func (p *Parser) parseTopLevel() ast.Node {
	switch p.current.Type {
	case lexer.TokenEOF:
		return nil
	case lexer.TokenError:
		p.error(fmt.Sprintf("lexer error: %s", p.current.Value))
		p.advance()
		return nil
	case lexer.TokenLParen:
		loc := p.position()
		p.advance() // skip (

		// Check for special forms that produce Nodes
		if p.current.Type == lexer.TokenSymbol {
			switch p.current.Value {
			case "def":
				p.advance()
				return p.parseDefAfterSymbol(loc)
			case "defn":
				p.advance()
				return p.parseDefnAfterSymbol(loc)
			case "deftype":
				p.advance()
				return p.parseDeftypeAfterSymbol(loc)
			case "defmacro":
				p.advance()
				return p.parseDefmacroAfterSymbol(loc)
			}
		}

		// Parse as expression (handles fn, if, do, let, quote, function calls, etc.)
		first := p.parseExprWithoutLParen()
		args := []ast.Expr{}
		for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
			args = append(args, p.parseExpr())
		}
		p.expect(lexer.TokenRParen)
		return &ast.FuncCall{Loc: loc, Func: first, Args: args}

	default:
		expr := p.parseExpr()
		return expr
	}
}

// parseExprWithoutLParen parses an expression assuming we've already consumed the LPAREN
func (p *Parser) parseExprWithoutLParen() ast.Expr {
	switch p.current.Type {
	case lexer.TokenSymbol:
		return p.parseSymbol()
	case lexer.TokenInt:
		return p.parseIntLit()
	case lexer.TokenFloat:
		return p.parseFloatLit()
	case lexer.TokenString:
		return p.parseStringLit()
	case lexer.TokenKeyword:
		return p.parseKeywordLit()
	case lexer.TokenTrue:
		loc := p.position()
		p.advance()
		return &ast.BoolLit{Loc: loc, Value: true}
	case lexer.TokenFalse:
		loc := p.position()
		p.advance()
		return &ast.BoolLit{Loc: loc, Value: false}
	case lexer.TokenNil:
		loc := p.position()
		p.advance()
		return &ast.NilLit{Loc: loc}
	default:
		p.error(fmt.Sprintf("unexpected token in expression: %s", p.current.Type))
		return &ast.NilLit{Loc: p.position()}
	}
}

// ============================================================================
// Expressions
// ============================================================================

func (p *Parser) parseExpr() ast.Expr {
	switch p.current.Type {
	case lexer.TokenInt:
		return p.parseIntLit()
	case lexer.TokenFloat:
		return p.parseFloatLit()
	case lexer.TokenString:
		return p.parseStringLit()
	case lexer.TokenKeyword:
		return p.parseKeywordLit()
	case lexer.TokenTrue:
		loc := p.position()
		p.advance()
		return &ast.BoolLit{Loc: loc, Value: true}
	case lexer.TokenFalse:
		loc := p.position()
		p.advance()
		return &ast.BoolLit{Loc: loc, Value: false}
	case lexer.TokenNil:
		loc := p.position()
		p.advance()
		return &ast.NilLit{Loc: loc}
	case lexer.TokenSymbol:
		return p.parseSymbol()
	case lexer.TokenLParen:
		loc := p.position()
		p.advance() // consume (

		if p.current.Type == lexer.TokenRParen {
			p.advance()
			return &ast.VectorLit{Loc: loc, Elts: []ast.Expr{}}
		}

		// Get the first element
		first := p.parseExpr()

		// Check if it's a special form
		if sym, ok := first.(*ast.Symbol); ok {
			switch sym.Name {
			case "fn":
				return p.parseFnAfterSymbol(loc)
			case "if":
				return p.parseIfAfterSymbol(loc)
			case "do":
				return p.parseDoAfterSymbol(loc)
			case "let":
				return p.parseLetAfterSymbol(loc)
			case "quote":
				return p.parseQuoteAfterSymbol(loc)
			}
		}

		// Regular function call
		args := []ast.Expr{}
		for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
			args = append(args, p.parseExpr())
		}

		p.expect(lexer.TokenRParen)
		return &ast.FuncCall{Loc: loc, Func: first, Args: args}
	case lexer.TokenLBracket:
		return p.parseVector()
	case lexer.TokenLBrace:
		return p.parseMap()
	case lexer.TokenQuote:
		return p.parseQuote()
	case lexer.TokenBackquote:
		p.error("quasiquote not yet implemented")
		p.advance()
		return &ast.NilLit{Loc: p.position()}
	default:
		p.error(fmt.Sprintf("unexpected token: %s", p.current.Type))
		return &ast.NilLit{Loc: p.position()}
	}
}

func (p *Parser) parseIntLit() ast.Expr {
	loc := p.position()
	val, err := strconv.ParseInt(p.current.Value, 0, 64)
	if err != nil {
		p.error(fmt.Sprintf("invalid integer: %s", p.current.Value))
		p.advance()
		return &ast.NilLit{Loc: loc}
	}
	p.advance()
	return &ast.IntLit{Loc: loc, Value: val}
}

func (p *Parser) parseFloatLit() ast.Expr {
	loc := p.position()
	val, err := strconv.ParseFloat(p.current.Value, 64)
	if err != nil {
		p.error(fmt.Sprintf("invalid float: %s", p.current.Value))
		p.advance()
		return &ast.NilLit{Loc: loc}
	}
	p.advance()
	return &ast.FloatLit{Loc: loc, Value: val}
}

func (p *Parser) parseStringLit() ast.Expr {
	loc := p.position()
	value := p.current.Value
	p.advance()
	return &ast.StringLit{Loc: loc, Value: value}
}

func (p *Parser) parseKeywordLit() ast.Expr {
	loc := p.position()
	value := p.current.Value
	p.advance()
	return &ast.KeywordLit{Loc: loc, Value: value}
}

func (p *Parser) parseSymbol() ast.Expr {
	loc := p.position()
	name := p.current.Value
	p.advance()
	return &ast.Symbol{Loc: loc, Name: name}
}

func (p *Parser) parseVector() ast.Expr {
	loc := p.position()
	p.expect(lexer.TokenLBracket)

	var elts []ast.Expr
	for p.current.Type != lexer.TokenRBracket && p.current.Type != lexer.TokenEOF {
		elts = append(elts, p.parseExpr())
	}

	p.expect(lexer.TokenRBracket)
	return &ast.VectorLit{Loc: loc, Elts: elts}
}

func (p *Parser) parseMap() ast.Expr {
	loc := p.position()
	p.expect(lexer.TokenLBrace)

	var pairs []ast.Pair
	for p.current.Type != lexer.TokenRBrace && p.current.Type != lexer.TokenEOF {
		key := p.parseExpr()
		if p.current.Type == lexer.TokenRBrace {
			p.error("map key without value")
			break
		}
		value := p.parseExpr()
		pairs = append(pairs, ast.Pair{Key: key, Value: value})
	}

	p.expect(lexer.TokenRBrace)
	return &ast.MapLit{Loc: loc, Pairs: pairs}
}

func (p *Parser) parseQuote() ast.Expr {
	loc := p.position()
	p.expect(lexer.TokenQuote)
	expr := p.parseExpr()
	return &ast.QuoteExpr{Loc: loc, Expr: expr}
}

// ============================================================================
// Definition parsing
// ============================================================================

func (p *Parser) parseDefAfterSymbol(loc ast.Position) ast.Node {
	// (def name value) or (def name :type value)
	if p.current.Type != lexer.TokenSymbol {
		p.error("def: expected name symbol")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	name := p.current.Value
	p.advance()

	var typeAnnot ast.Type
	value := p.parseExpr()

	// Check if next is a type annotation (peek ahead by trying to parse as type)
	if sym, ok := value.(*ast.Symbol); ok && (sym.Name == "int" || sym.Name == "string" || sym.Name == "float" || sym.Name == "bool") {
		typeAnnot = &ast.NamedType{Loc: value.Pos(), Name: sym.Name}
		if p.current.Type == lexer.TokenRParen {
			p.error("def: type annotation without value")
			p.advance()
			return &ast.Def{Loc: loc, Name: name, Type: typeAnnot}
		}
		value = p.parseExpr()
	}

	p.expect(lexer.TokenRParen)
	return &ast.Def{Loc: loc, Name: name, Type: typeAnnot, Value: value}
}

func (p *Parser) parseDefnAfterSymbol(loc ast.Position) ast.Node {
	// (defn name [params] body...)
	if p.current.Type != lexer.TokenSymbol {
		p.error("defn: expected name symbol")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	name := p.current.Value
	p.advance()

	if p.current.Type != lexer.TokenLBracket {
		p.error("defn: expected parameter vector")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	params := p.parseParams()

	var returnType ast.Type
	// Check for return type annotation
	if p.current.Type != lexer.TokenRParen && isTypeSymbol(p.current) {
		returnType = &ast.NamedType{Loc: p.position(), Name: p.current.Value}
		p.advance()
	}

	var body []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		body = append(body, p.parseExpr())
	}

	p.expect(lexer.TokenRParen)
	return &ast.Defn{Loc: loc, Name: name, Params: params, ReturnType: returnType, Body: body}
}

func (p *Parser) parseDeftypeAfterSymbol(loc ast.Position) ast.Node {
	// (deftype Name {field Type ...})
	if p.current.Type != lexer.TokenSymbol {
		p.error("deftype: expected name symbol")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	name := p.current.Value
	p.advance()

	if p.current.Type != lexer.TokenLBrace {
		p.error("deftype: expected field map")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	p.expect(lexer.TokenLBrace)

	var fields []*ast.Field
	for p.current.Type != lexer.TokenRBrace && p.current.Type != lexer.TokenEOF {
		if p.current.Type != lexer.TokenSymbol {
			p.error("deftype: expected field name")
			p.advance()
			continue
		}

		fieldName := p.current.Value
		p.advance()

		// Expect type
		if p.current.Type == lexer.TokenRParen {
			p.error("deftype: expected field type")
			break
		}

		fieldType := p.parseType()

		fields = append(fields, &ast.Field{Name: fieldName, Type: fieldType})
	}

	p.expect(lexer.TokenRBrace)
	p.expect(lexer.TokenRParen)

	return &ast.Deftype{Loc: loc, Name: name, Fields: fields}
}

func (p *Parser) parseDefmacroAfterSymbol(loc ast.Position) ast.Node {
	// (defmacro name [params] body...)
	if p.current.Type != lexer.TokenSymbol {
		p.error("defmacro: expected name symbol")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	name := p.current.Value
	p.advance()

	if p.current.Type != lexer.TokenLBracket {
		p.error("defmacro: expected parameter vector")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	params := p.parseParams()

	var body []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		body = append(body, p.parseExpr())
	}

	p.expect(lexer.TokenRParen)
	return &ast.Defmacro{Loc: loc, Name: name, Params: params, Body: body}
}

func (p *Parser) parseFnAfterSymbol(loc ast.Position) ast.Expr {
	// (fn [params] body...)
	if p.current.Type != lexer.TokenLBracket {
		p.error("fn: expected parameter vector")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	params := p.parseParams()

	var body []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		body = append(body, p.parseExpr())
	}

	p.expect(lexer.TokenRParen)
	return &ast.FuncLit{Loc: loc, Params: params, Body: body}
}

func (p *Parser) parseIfAfterSymbol(loc ast.Position) ast.Expr {
	// (if cond then else?)
	cond := p.parseExpr()
	then := p.parseExpr()

	var elseExpr ast.Expr
	if p.current.Type != lexer.TokenRParen {
		elseExpr = p.parseExpr()
	}

	p.expect(lexer.TokenRParen)
	return &ast.IfExpr{Loc: loc, Cond: cond, Then: then, Else: elseExpr}
}

func (p *Parser) parseDoAfterSymbol(loc ast.Position) ast.Expr {
	// (do expr1 expr2 ...)
	var exprs []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		exprs = append(exprs, p.parseExpr())
	}

	p.expect(lexer.TokenRParen)
	return &ast.DoExpr{Loc: loc, Exprs: exprs}
}

func (p *Parser) parseLetAfterSymbol(loc ast.Position) ast.Expr {
	// (let [x 1 y 2] expr1 expr2 ...)
	if p.current.Type != lexer.TokenLBracket {
		p.error("let: expected bindings vector")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	bindings := p.parseBindings()

	var body []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		body = append(body, p.parseExpr())
	}

	p.expect(lexer.TokenRParen)
	return &ast.LetExpr{Loc: loc, Bindings: bindings, Body: body}
}

func (p *Parser) parseQuoteAfterSymbol(loc ast.Position) ast.Expr {
	// (quote expr)
	expr := p.parseExpr()
	p.expect(lexer.TokenRParen)
	return &ast.QuoteExpr{Loc: loc, Expr: expr}
}

// ============================================================================
// Helpers
// ============================================================================

func (p *Parser) parseParams() []*ast.Param {
	p.expect(lexer.TokenLBracket)

	var params []*ast.Param
	for p.current.Type != lexer.TokenRBracket && p.current.Type != lexer.TokenEOF {
		if p.current.Type != lexer.TokenSymbol {
			p.error("expected parameter name")
			p.advance()
			continue
		}

		name := p.current.Value
		p.advance()

		// Check for type annotation
		var paramType ast.Type
		if isTypeSymbol(p.current) {
			paramType = &ast.NamedType{Loc: p.position(), Name: p.current.Value}
			p.advance()
		}

		params = append(params, &ast.Param{Name: name, Type: paramType})
	}

	p.expect(lexer.TokenRBracket)
	return params
}

func (p *Parser) parseBindings() []*ast.Binding {
	p.expect(lexer.TokenLBracket)

	var bindings []*ast.Binding
	for p.current.Type != lexer.TokenRBracket && p.current.Type != lexer.TokenEOF {
		if p.current.Type != lexer.TokenSymbol {
			p.error("expected binding name")
			p.advance()
			continue
		}

		name := p.current.Value
		p.advance()

		// Check for type annotation
		var bindType ast.Type
		if isTypeSymbol(p.current) {
			bindType = &ast.NamedType{Loc: p.position(), Name: p.current.Value}
			p.advance()
		}

		// Parse initializer
		init := p.parseExpr()

		bindings = append(bindings, &ast.Binding{Name: name, Type: bindType, Init: init})
	}

	p.expect(lexer.TokenRBracket)
	return bindings
}

func (p *Parser) parseType() ast.Type {
	loc := p.position()

	if p.current.Type == lexer.TokenSymbol {
		name := p.current.Value
		p.advance()
		return &ast.NamedType{Loc: loc, Name: name}
	}

	if p.current.Type == lexer.TokenLBracket {
		p.advance()
		elemType := p.parseType()
		p.expect(lexer.TokenRBracket)
		return &ast.SliceType{Loc: loc, EltType: elemType}
	}

	if p.current.Type == lexer.TokenLParen {
		p.advance()
		if p.current.Type == lexer.TokenSymbol && p.current.Value == "chan" {
			p.advance()
			elemType := p.parseType()
			var buffer int64
			if p.current.Type == lexer.TokenInt {
				if val, err := strconv.ParseInt(p.current.Value, 0, 64); err == nil {
					buffer = val
				}
				p.advance()
			}
			p.expect(lexer.TokenRParen)
			return &ast.ChanType{Loc: loc, EltType: elemType, Buffer: buffer}
		}
		p.error("unexpected type")
		p.skipToRParen()
	}

	p.error(fmt.Sprintf("unexpected type token: %s", p.current.Type))
	return &ast.NamedType{Loc: loc, Name: "unknown"}
}

// ============================================================================
// Utilities
// ============================================================================

func (p *Parser) advance() {
	if p.hasPeek {
		p.current = p.peeked
		p.hasPeek = false
	} else {
		p.current = p.lex.NextNonNewline()
	}
}

func (p *Parser) peek() lexer.Token {
	if !p.hasPeek {
		p.peeked = p.lex.NextNonNewline()
		p.hasPeek = true
	}
	return p.peeked
}

func (p *Parser) expect(tt lexer.TokenType) bool {
	if p.current.Type != tt {
		p.error(fmt.Sprintf("expected %s, got %s", tt, p.current.Type))
		return false
	}
	p.advance()
	return true
}

func (p *Parser) position() ast.Position {
	return ast.Position{Line: p.current.Line, Column: p.current.Column}
}

func (p *Parser) error(msg string) {
	p.errors = append(p.errors, fmt.Sprintf("line %d: %s", p.current.Line, msg))
}

func (p *Parser) skipToRParen() {
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		p.advance()
	}
	if p.current.Type == lexer.TokenRParen {
		p.advance()
	}
}

func isTypeSymbol(tok lexer.Token) bool {
	if tok.Type != lexer.TokenSymbol {
		return false
	}
	switch tok.Value {
	case "int", "float", "string", "bool", "byte", "rune":
		return true
	}
	return false
}
