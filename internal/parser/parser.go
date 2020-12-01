package parser

import (
	"fmt"
	"io"

	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/token"
)

// Parser describes a parser that can parse input into a Lua ast.Block.
type Parser interface {
	Parse() (ast.Block, bool)
	Errors() []error
}

type parser struct {
	scanner

	input  io.Reader
	errors []error

	tkstash []token.Token
}

// New creates a new single-use Lua-parser.
func New(input io.Reader) (Parser, error) {
	sc, err := newInMemoryScanner(input)
	if err != nil {
		return nil, fmt.Errorf("in memory scanner: %w", err)
	}
	return &parser{
		scanner: sc,
	}, nil
}

// Parse parses the input of this parser. If the parsing was successful, true will be returned.
// Otherwise, a potentially incomplete, partial Ast together with false will be returned.
// If this method returns false, obtain the parse errors with Parser.Errors.
func (p *parser) Parse() (ast.Block, bool) {
	block := p.block()
	return block, len(p.errors) == 0
}

// Errors returns all the parse errors that may have occurred during the parsing.
func (p *parser) Errors() []error {
	return p.errors
}

func (p *parser) collectError(err error) {
	if err != nil {
		p.errors = append(p.errors, err)
	}
}

func (p *parser) stash(tokens ...token.Token) {
	p.tkstash = append(tokens, p.tkstash...)
}

func (p *parser) next() (token.Token, bool) {
	if len(p.tkstash) > 0 {
		tk := p.tkstash[0]
		p.tkstash = p.tkstash[1:]
		return tk, true
	}
	for {
		next, ok := p.scanner.next()
		if next != nil && next.Is(token.Error) {
			p.collectError(fmt.Errorf("error at %s: %s", next.Pos(), next.Value()))
		}
		if !ok {
			return nil, ok
		}
		return next, ok
	}
}

func (p *parser) block() ast.Block {
	block := ast.Block{}
	for stmt := p.stmt(); stmt != nil; stmt = p.stmt() {
		block = append(block, stmt)
	}
	return block
}

func (p *parser) stmt() ast.Statement {
	tk, ok := p.next()
	if !ok {
		return nil
	}
	switch {
	case tk.Is(token.Name):
		next, ok := p.next()
		if !ok {
			p.collectError(fmt.Errorf("unexpected EOF, expected one of ',', '=', ':' or '('"))
			return nil
		}

		p.stash(tk, next)
		switch {
		case next.Is(token.Comma) || next.Is(token.Assign):
			// varlist -> assignment
			return p.assignment()
		case next.Is(token.Colon) || next.Is(token.ParLeft):
			// prefixexp -> functioncall
			call, ok := p.functionCall()
			if !ok {
				p.collectError(fmt.Errorf("expected function call, but got nothing"))
				return nil
			}
			return call
		}

	}
	return nil
}

func (p *parser) assignment() ast.Assignment {
	varlist := p.varlist()

	// check if there's a '=' between varlist and explist
	assign, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected '=' followed by explist, but got EOF"))
		return ast.Assignment{}
	}
	if !assign.Is(token.Assign) {
		p.collectError(fmt.Errorf("expected '=' followed by explist, but got %s", assign))
		return ast.Assignment{}
	}

	explist := p.explist()

	return ast.Assignment{
		VarList: varlist,
		ExpList: explist,
	}
}

func (p *parser) varlist() []ast.Var {
	var list []ast.Var
	for v, ok := p.var_(); ok; v, ok = p.var_() {
		list = append(list, v)
		next, ok := p.next()
		if !ok {
			break
		}
		if !next.Is(token.Comma) {
			p.stash(next)
			break
		}
	}
	return list
}

func (p *parser) var_() (ast.Var, bool) {
	name, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected a name, but got EOF"))
		return ast.Var{}, false
	}

	if !name.Is(token.Name) {
		p.collectError(fmt.Errorf("expected a name, but got %s", name))
		return ast.Var{}, false
	}

	// if '[' or '.' follows, it's a prefixexp
	lookahead, ok := p.next()
	if ok && (lookahead.Is(token.BracketLeft) || lookahead.Is(token.Dot)) {
		p.stash(name, lookahead)
		prefixexp, ok := p.prefixexp()
		if !ok {
			p.collectError(fmt.Errorf("expected a prefixexp, but got EOF"))
			return ast.Var{}, false
		}
		return ast.Var{
			PrefixExp: prefixexp,
		}, true
	}
	if ok {
		// lookahead not needed, so put back the lookahead
		p.stash(lookahead)
	}

	return ast.Var{
		Name: name,
	}, true
}

func (p *parser) explist() []ast.Exp {
	var list []ast.Exp
	for exp := p.exp(); exp != nil; exp = p.exp() {
		list = append(list, exp)
		next, ok := p.next()
		if !ok {
			break
		}
		if !next.Is(token.Comma) {
			p.stash(next)
			break
		}
	}
	return list
}

func (p *parser) exp() ast.Exp {
	next, ok := p.next()
	if !ok {
		return nil
	}
	switch {
	case next.Is(token.Nil):
		return ast.SimpleExp{
			Nil: next,
		}
	case next.Is(token.False):
		return ast.SimpleExp{
			False: next,
		}
	case next.Is(token.True):
		return ast.SimpleExp{
			True: next,
		}
	case next.Is(token.Ellipsis):
		return ast.SimpleExp{
			Ellipsis: next,
		}
	case next.Is(token.Number):
		return ast.SimpleExp{
			Number: next,
		}
	case next.Is(token.String):
		return ast.SimpleExp{
			String: next,
		}
	case next.Is(token.UnaryOperator):
		exp := p.exp()
		if exp == nil {
			p.collectError(fmt.Errorf("expected expression after unary operator %s, but got nothing", next))
			return nil
		}
		return ast.UnopExp{
			Unop: next,
			Exp:  exp,
		}
	case next.Is(token.ParLeft) || next.Is(token.Name):
		p.stash(next)
		prefixexp, ok := p.prefixexp()
		if !ok {
			p.collectError(fmt.Errorf("expected prefixexp, but got nothing"))
			return nil
		}
		return prefixexp
	}
	panic("implement function, tableconstructor, binary operation")
}

func (p *parser) prefixexp() (ast.PrefixExp, bool) {
	next, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected prefixexp, but got EOF"))
		return ast.PrefixExp{}, false
	}
	switch {
	case next.Is(token.ParLeft):
		exp := p.exp()
		if exp == nil {
			p.collectError(fmt.Errorf("expected exp after %s, but got nothing", next))
			return ast.PrefixExp{}, false
		}
		parRight, ok := p.next()
		if !ok {
			p.collectError(fmt.Errorf("expected rightpar after exp, but got EOF"))
		}
		if !parRight.Is(token.ParRight) {
			p.collectError(fmt.Errorf("expected rightpar after exp, but got %s", parRight))
		}
		return ast.PrefixExp{
			Exp: exp,
		}, true
	case next.Is(token.Name):
		p.stash(next)
		v, ok := p.var_()
		if !ok {
			p.collectError(fmt.Errorf("expected var, but got nothing"))
			return ast.PrefixExp{}, false
		}
		return ast.PrefixExp{
			Var: v,
		}, true
	}
	panic("implement functioncall")
}

func (p *parser) functionCall() (ast.FunctionCall, bool) {
	prefixexp, ok := p.prefixexp()
	if !ok {
		p.collectError(fmt.Errorf("expected prefixexp, but got nothing"))
		return ast.FunctionCall{}, false
	}
	colon, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected ':' or args, but got EOF"))
		return ast.FunctionCall{}, false
	}
	var name token.Token
	if colon.Is(token.Colon) {
		next, ok := p.next()
		if !ok {
			p.collectError(fmt.Errorf("expected name, but got EOF"))
			return ast.FunctionCall{}, false
		}
		name = next
	}

	p.stash(colon) // not a colon, put it back for args
	args, ok := p.args()
	if !ok {
		p.collectError(fmt.Errorf("expected args, but got nothing"))
		return ast.FunctionCall{}, false
	}
	return ast.FunctionCall{
		PrefixExp: prefixexp,
		Name:      name,
		Args:      args,
	}, true
}

func (p *parser) args() (ast.Args, bool) {
	next, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected '(' or a name, but got EOF"))
		return ast.Args{}, false
	}
	switch {
	case next.Is(token.ParLeft):
		explist := p.explist()
		rightPar, ok := p.next()
		if !ok {
			p.collectError(fmt.Errorf("expected ')', but got EOF"))
			return ast.Args{}, false
		}
		if !rightPar.Is(token.ParRight) {
			p.collectError(fmt.Errorf("expected ')', but got %s", rightPar))
		}
		return ast.Args{
			ExpList: explist,
		}, true
	case next.Is(token.CurlyLeft):
		p.collectError(fmt.Errorf("table constructors not supported yet"))
		return ast.Args{}, false
	case next.Is(token.String):
		return ast.Args{
			String: next,
		}, true
	}
	p.collectError(fmt.Errorf("expected one of '(', '{' or a String, but got %s", next))
	return ast.Args{}, false
}
