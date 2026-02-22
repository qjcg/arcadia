package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qjcg/arcadia/exp/elbereth/internal/ast"
	"github.com/qjcg/arcadia/exp/elbereth/internal/lexer"
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
	prog := &ast.Program{
		Loc:   p.position(),
		Items: []ast.Node{},
	}

	// Check for #lang at the very beginning
	if p.current.Type == lexer.TokenHashLang {
		p.advance()
		if p.current.Type != lexer.TokenSymbol {
			p.error("expected language name after #lang")
		} else {
			prog.Lang = p.current.Value
			p.advance()
			// If we found a #lang directive, we STOP parsing here
			// and let the caller delegate the rest of the file.
			return prog, nil
		}
	}

	for p.current.Type != lexer.TokenEOF {
		item := p.parseTopLevel()
		if item != nil {
			if pkg, ok := item.(*ast.Package); ok {
				prog.Package = pkg.Name
			}
			prog.Items = append(prog.Items, item)
		}
		if len(p.errors) > 0 {
			break
		}
	}

	if len(p.errors) > 0 {
		return nil, fmt.Errorf("parse errors: %v", p.errors)
	}

	return prog, nil
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
			case "deftest":
				p.advance()
				return p.parseDeftestAfterSymbol(loc)
			case "defbenchmark":
				p.advance()
				return p.parseDefbenchmarkAfterSymbol(loc)
			case "defexample":
				p.advance()
				return p.parseDefexampleAfterSymbol(loc)
			case "package":
				p.advance()
				return p.parsePackageAfterSymbol(loc)
			case "import":
				p.advance()
				return p.parseImportAfterSymbol(loc)
			}

			// If it's a special form that is an Expr, we can use parseExpr logic
			// (but we need to handle the fact that we already consumed LParen)
			name := p.current.Value
			switch name {
			case "fn", "if", "do", "let", "quote", "match", "loop", "recur", "select", "doseq":
				// We need to re-handle these because they are Exprs but can appear at top level
				p.advance() // consume the symbol
				switch name {
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
				case "match":
					return p.parseMatchAfterSymbol(loc)
				case "loop":
					return p.parseLoopAfterSymbol(loc)
				case "recur":
					return p.parseRecurAfterSymbol(loc)
				case "select":
					return p.parseSelectAfterSymbol(loc)
				case "doseq":
					return p.parseDoseqAfterSymbol(loc)
				}
			}
		}

		// Parse as regular function call
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
			case "match":
				return p.parseMatchAfterSymbol(loc)
			case "loop":
				return p.parseLoopAfterSymbol(loc)
			case "recur":
				return p.parseRecurAfterSymbol(loc)
			case "select":
				return p.parseSelectAfterSymbol(loc)
			case "doseq":
				return p.parseDoseqAfterSymbol(loc)
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
		return p.parseBackquote()
	case lexer.TokenUnquote:
		return p.parseUnquote()
	case lexer.TokenUnquoteSplice:
		return p.parseUnquoteSplice()
	default:
		p.error(fmt.Sprintf("unexpected token: %s", p.current.Type))
		loc := p.position()
		p.advance()
		return &ast.NilLit{Loc: loc}
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

func (p *Parser) parseBackquote() ast.Expr {
	loc := p.position()
	p.expect(lexer.TokenBackquote)
	expr := p.parseExpr()
	return &ast.BackquoteExpr{Loc: loc, Expr: expr}
}

func (p *Parser) parseUnquote() ast.Expr {
	loc := p.position()
	p.expect(lexer.TokenUnquote)
	expr := p.parseExpr()
	return &ast.UnquoteExpr{Loc: loc, Expr: expr}
}

func (p *Parser) parseUnquoteSplice() ast.Expr {
	loc := p.position()
	p.expect(lexer.TokenUnquoteSplice)
	expr := p.parseExpr()
	return &ast.UnquoteSpliceExpr{Loc: loc, Expr: expr}
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
	// (deftype Name params? body)
	if p.current.Type != lexer.TokenSymbol {
		p.error("deftype: expected name symbol")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	name := p.current.Value
	p.advance()

	var params []string
	// Optional generic parameters (e.g., :T or [T])
	for p.current.Type == lexer.TokenKeyword || p.current.Type == lexer.TokenSymbol {
		// If it's a bracket or brace or paren, it's the start of the body
		if p.current.Type == lexer.TokenLBrace || p.current.Type == lexer.TokenLParen || p.current.Type == lexer.TokenLBracket {
			break
		}
		params = append(params, p.current.Value)
		p.advance()
	}

	// Body can be a map {f T} or a list of variants (:tag T)
	if p.current.Type == lexer.TokenLBrace {
		p.advance()
		var fields []*ast.Field
		for p.current.Type != lexer.TokenRBrace && p.current.Type != lexer.TokenEOF {
			if p.current.Type != lexer.TokenSymbol {
				p.error("deftype: expected field name")
				p.advance()
				continue
			}
			fieldName := p.current.Value
			p.advance()
			fieldType := p.parseType()
			fields = append(fields, &ast.Field{Name: fieldName, Type: fieldType})
		}
		p.expect(lexer.TokenRBrace)
		p.expect(lexer.TokenRParen)
		return &ast.Deftype{Loc: loc, Name: name, Params: params, Fields: fields}
	} else {
		// Assume variants
		var variants []*ast.Variant
		for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
			if p.current.Type == lexer.TokenLParen {
				p.advance()
				if p.current.Type != lexer.TokenKeyword && p.current.Type != lexer.TokenSymbol {
					p.error("deftype: expected variant tag")
					p.advance()
				} else {
					tagName := p.current.Value
					p.advance()
					var tagType ast.Type
					if p.current.Type != lexer.TokenRParen {
						tagType = p.parseType()
					}
					variants = append(variants, &ast.Variant{Name: tagName, Type: tagType})
				}
				p.expect(lexer.TokenRParen)
			} else if p.current.Type == lexer.TokenKeyword || p.current.Type == lexer.TokenSymbol {
				variants = append(variants, &ast.Variant{Name: p.current.Value})
				p.advance()
			} else {
				p.error("deftype: expected variant definition")
				p.advance()
			}
		}
		p.expect(lexer.TokenRParen)
		return &ast.Deftype{Loc: loc, Name: name, Params: params, Variants: variants}
	}
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

func (p *Parser) parseDeftestAfterSymbol(loc ast.Position) ast.Node {
	// (deftest name body...)
	if p.current.Type != lexer.TokenSymbol {
		p.error("deftest: expected name symbol")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	name := p.current.Value
	p.advance()

	var body []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		body = append(body, p.parseExpr())
	}

	p.expect(lexer.TokenRParen)
	return &ast.Deftest{Loc: loc, Name: name, Body: body}
}

func (p *Parser) parseDefbenchmarkAfterSymbol(loc ast.Position) ast.Node {
	// (defbenchmark name [b] body...)
	if p.current.Type != lexer.TokenSymbol {
		p.error("defbenchmark: expected name symbol")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	name := p.current.Value
	p.advance()

	bParam := "b"
	if p.current.Type == lexer.TokenLBracket {
		params := p.parseParams()
		if len(params) > 0 {
			bParam = params[0].Name
		}
	}

	var body []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		body = append(body, p.parseExpr())
	}

	p.expect(lexer.TokenRParen)
	return &ast.Defbenchmark{Loc: loc, Name: name, BParam: bParam, Body: body}
}

func (p *Parser) parseDefexampleAfterSymbol(loc ast.Position) ast.Node {
	// (defexample name body...)
	if p.current.Type != lexer.TokenSymbol {
		p.error("defexample: expected name symbol")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	name := p.current.Value
	p.advance()

	var body []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		body = append(body, p.parseExpr())
	}

	p.expect(lexer.TokenRParen)
	return &ast.Defexample{Loc: loc, Name: name, Body: body}
}

func (p *Parser) parsePackageAfterSymbol(loc ast.Position) ast.Node {
	// (package name)
	if p.current.Type != lexer.TokenSymbol {
		p.error("package: expected name symbol")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}

	name := p.current.Value
	p.advance()

	p.expect(lexer.TokenRParen)
	return &ast.Package{Loc: loc, Name: name}
}

func (p *Parser) parseImportAfterSymbol(loc ast.Position) ast.Node {
	// (import "path" "path2" [alias "path3"] ...)
	var specs []ast.ImportSpec

	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		var path, alias string

		if p.current.Type == lexer.TokenString {
			path = p.current.Value
			p.advance()
		} else if p.current.Type == lexer.TokenLBracket {
			p.advance() // [
			if p.current.Type != lexer.TokenSymbol {
				p.error("import: expected alias symbol")
			} else {
				alias = p.current.Value
				p.advance()
			}

			if p.current.Type != lexer.TokenString {
				p.error("import: expected path string")
			} else {
				path = p.current.Value
				p.advance()
			}
			p.expect(lexer.TokenRBracket)
		} else {
			p.error("import: expected string or [alias \"path\"]")
			p.advance()
			break
		}
		specs = append(specs, ast.ImportSpec{Path: path, Alias: alias})
	}

	p.expect(lexer.TokenRParen)
	return &ast.Import{Loc: loc, Specs: specs}
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

func (p *Parser) parseMatchAfterSymbol(loc ast.Position) ast.Expr {
	// (match val pattern1 body1 ...)
	val := p.parseExpr()
	var cases []ast.MatchCase
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		pattern := p.parseExpr()
		body := p.parseExpr()
		cases = append(cases, ast.MatchCase{Pattern: pattern, Body: body})
	}
	p.expect(lexer.TokenRParen)
	return &ast.MatchExpr{Loc: loc, Val: val, Cases: cases}
}

func (p *Parser) parseLoopAfterSymbol(loc ast.Position) ast.Expr {
	// (loop [bindings] body...)
	if p.current.Type != lexer.TokenLBracket {
		p.error("loop: expected bindings vector")
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
	return &ast.LoopExpr{Loc: loc, Bindings: bindings, Body: body}
}

func (p *Parser) parseDoseqAfterSymbol(loc ast.Position) ast.Expr {
	// (doseq [var coll] body...)
	if p.current.Type != lexer.TokenLBracket {
		p.error("doseq: expected [var coll] vector")
		p.advance()
		p.skipToRParen()
		return &ast.NilLit{Loc: loc}
	}
	p.advance() // [

	if p.current.Type != lexer.TokenSymbol {
		p.error("doseq: expected variable name")
	}
	varName := p.current.Value
	p.advance()

	coll := p.parseExpr()
	p.expect(lexer.TokenRBracket)

	var body []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		body = append(body, p.parseExpr())
	}

	p.expect(lexer.TokenRParen)
	return &ast.DoseqExpr{Loc: loc, Var: varName, Coll: coll, Body: body}
}

func (p *Parser) parseRecurAfterSymbol(loc ast.Position) ast.Expr {
	// (recur arg1 arg2 ...)
	var args []ast.Expr
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		args = append(args, p.parseExpr())
	}
	p.expect(lexer.TokenRParen)
	return &ast.RecurExpr{Loc: loc, Args: args}
}

func (p *Parser) parseSelectAfterSymbol(loc ast.Position) ast.Expr {
	// (select [chan val] body ...)
	var cases []ast.SelectCase
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		if p.current.Type != lexer.TokenLBracket {
			p.error("select: expected case vector [chan binding]")
			p.advance()
			continue
		}
		p.advance() // [

		var ch ast.Expr
		var binding string

		if p.current.Type == lexer.TokenSymbol && (p.current.Value == "default" || p.current.Value == ":default") {
			p.advance()
		} else {
			ch = p.parseExpr()
			if p.current.Type == lexer.TokenSymbol {
				binding = p.current.Value
				p.advance()
			}
		}
		p.expect(lexer.TokenRBracket)

		// Parse body (expr until next case or end)
		var body []ast.Expr
		for p.current.Type != lexer.TokenLBracket && p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
			body = append(body, p.parseExpr())
		}

		cases = append(cases, ast.SelectCase{Chan: ch, Binding: binding, Body: body})
	}
	p.expect(lexer.TokenRParen)
	return &ast.SelectExpr{Loc: loc, Cases: cases}
}

// ============================================================================
// Helpers
// ============================================================================

func (p *Parser) parseParams() []*ast.Param {
	p.expect(lexer.TokenLBracket)

	var params []*ast.Param
	for p.current.Type != lexer.TokenRBracket && p.current.Type != lexer.TokenEOF {
		isVariadic := false
		if p.current.Type == lexer.TokenSymbol && p.current.Value == "&" {
			isVariadic = true
			p.advance()
		}

		if p.current.Type != lexer.TokenSymbol {
			p.error("expected parameter name")
			p.advance()
			continue
		}

		name := p.current.Value
		p.advance()

		// Check for type annotation (complex types like (chan string))
		var paramType ast.Type
		if p.current.Type == lexer.TokenKeyword || p.current.Type == lexer.TokenSymbol || p.current.Type == lexer.TokenLParen || p.current.Type == lexer.TokenLBracket {
			// If it's a bracket/paren/symbol/keyword, it might be a type
			// But we need to be careful not to consume next parameter name.
			// isTypeSymbol is too restrictive, but p.parseType might be too greedy.
			// Actually Elbereth spec uses :Type syntax for annotations usually.
			if p.current.Type == lexer.TokenKeyword {
				paramType = p.parseType()
			} else if p.current.Type == lexer.TokenLParen || p.current.Type == lexer.TokenLBracket {
				paramType = p.parseType()
			} else if isTypeSymbol(p.current) {
				paramType = p.parseType()
			}
		}

		params = append(params, &ast.Param{Name: name, Type: paramType, Variadic: isVariadic})
	}

	p.expect(lexer.TokenRBracket)
	return params
}

func (p *Parser) parseBindings() []*ast.Binding {
	p.expect(lexer.TokenLBracket)

	var bindings []*ast.Binding
	for p.current.Type != lexer.TokenRBracket && p.current.Type != lexer.TokenEOF {
		var names []string
		if p.current.Type == lexer.TokenSymbol {
			names = append(names, p.current.Value)
			p.advance()
		} else if p.current.Type == lexer.TokenLBracket {
			p.advance() // [
			for p.current.Type != lexer.TokenRBracket && p.current.Type != lexer.TokenEOF {
				if p.current.Type != lexer.TokenSymbol {
					p.error("expected symbol in destructuring")
				} else {
					names = append(names, p.current.Value)
				}
				p.advance()
			}
			p.expect(lexer.TokenRBracket)
		} else {
			p.error("expected binding name or [names]")
			p.advance()
			continue
		}

		// Check for type annotation
		var bindType ast.Type
		if isTypeSymbol(p.current) {
			bindType = &ast.NamedType{Loc: p.position(), Name: p.current.Value}
			p.advance()
		}

		// Parse initializer
		init := p.parseExpr()

		bindings = append(bindings, &ast.Binding{Names: names, Type: bindType, Init: init})
	}

	p.expect(lexer.TokenRBracket)
	return bindings
}

func (p *Parser) parseType() ast.Type {
	loc := p.position()

	if p.current.Type == lexer.TokenSymbol || p.current.Type == lexer.TokenKeyword {
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
	if tok.Type != lexer.TokenSymbol && tok.Type != lexer.TokenKeyword {
		return false
	}
	if strings.HasPrefix(tok.Value, "*") {
		return true
	}
	switch tok.Value {
	case "int", "float", "string", "bool", "byte", "rune":
		return true
	}
	// Capitalized names are often types in Go
	if len(tok.Value) > 0 && tok.Value[0] >= 'A' && tok.Value[0] <= 'Z' {
		return true
	}
	return false
}
