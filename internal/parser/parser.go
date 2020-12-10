package parser

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/token"
)

// Parser describes a parser that can parse input into a Lua ast.Block.
type Parser interface {
	Parse() (ast.Chunk, bool)
	Errors() []error
}

type namer interface {
	Name() string
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
		input:   input,
	}, nil
}

// Parse parses the input of this parser. If the parsing was successful, true will be returned.
// Otherwise, a potentially incomplete, partial Ast together with false will be returned.
// If this method returns false, obtain the parse errors with Parser.Errors.
func (p *parser) Parse() (ast.Chunk, bool) {
	block := p.block()

	name := "<unknown input>"
	if n, ok := p.input.(namer); ok {
		name = filepath.Base(n.Name())
	}

	return ast.Chunk{
		Block: block,
		Name:  name,
	}, len(p.errors) == 0
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
		if next == nil {
			p.collectError(fmt.Errorf("could not compute token"))
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
	case tk.Is(token.Local):
		next, ok := p.next()
		if !ok {
			p.collectError(fmt.Errorf("unexpected EOF, expected either 'function' or a name"))
			return nil
		}

		p.stash(tk, next)
		switch {
		case next.Is(token.Name):
			return p.local()
		case next.Is(token.Function):
			panic("unsupported")
		}
	case tk.Is(token.Function):
		p.stash(tk)
		fn, ok := p.function()
		if !ok {
			p.collectError(fmt.Errorf("expected function, but got nothing"))
			return nil
		}
		return fn
	case tk.Is(token.End):
		// This is kind of a workaround.
		// If we try to parse ay kind of block, it consists
		// of statements followed by 'end'. 'end' should not give
		// a statement, so stash it and return nil.
		p.stash(tk)
		return nil
	}
	p.collectError(fmt.Errorf("unexpected token %s", tk))
	return nil
}

func (p *parser) function() (ast.Function, bool) {
	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("function"))
		return ast.Function{}, false
	}
	if !next.Is(token.Function) {
		p.collectError(ErrUnexpectedThing("function", next))
		return ast.Function{}, false
	}

	next, ok = p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("name or body"))
		return ast.Function{}, false
	}
	var name *ast.FuncName
	if next.Is(token.Name) {
		p.stash(next)
		fname, ok := p.funcname()
		if !ok {
			p.collectError(ErrUnexpectedEof("name"))
			return ast.Function{}, false
		}
		name = &fname
	}
	fbody, ok := p.funcbody()
	if !ok {
		p.collectError(ErrExpectedSomething("funcbody"))
		return ast.Function{}, false
	}
	return ast.Function{
		FuncName: name,
		FuncBody: fbody,
	}, true
}

func (p *parser) funcname() (ast.FuncName, bool) {
	var list []token.Token
	for name, ok := p.next(); ok; name, ok = p.next() {
		list = append(list, name)
		next, ok := p.next()
		if !ok {
			break
		}
		if !next.Is(token.Dot) {
			p.stash(next)
			break
		}
	}
	next, ok := p.next()
	if !ok {
		return ast.FuncName{
			Name1: list,
		}, true
	}
	if !next.Is(token.Colon) {
		p.stash(next)
		return ast.FuncName{
			Name1: list,
		}, true
	}

	next, ok = p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("name"))
		return ast.FuncName{}, false
	}
	if !next.Is(token.Name) {
		p.collectError(ErrUnexpectedThing("name", next))
		return ast.FuncName{}, false
	}
	return ast.FuncName{
		Name1: list,
		Name2: next,
	}, true
}

func (p *parser) funcbody() (ast.FuncBody, bool) {
	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("'('"))
		return ast.FuncBody{}, false
	}
	if !next.Is(token.ParLeft) {
		p.collectError(ErrUnexpectedThing("'('", next))
		return ast.FuncBody{}, false
	}

	next, ok = p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("parlist or ')'"))
		return ast.FuncBody{}, false
	}
	if !next.Is(token.Name) && !next.Is(token.Ellipsis) && !next.Is(token.ParRight) {
		p.collectError(ErrUnexpectedThing("parlist or ')'", next))
		return ast.FuncBody{}, false
	}

	var parlist ast.ParList
	if !next.Is(token.ParRight) {
		p.stash(next)
		l, ok := p.parlist()
		if !ok {
			p.collectError(ErrExpectedSomething("parlist"))
			return ast.FuncBody{}, false
		}
		parlist = l

		next, ok = p.next()
		if !ok {
			p.collectError(ErrUnexpectedEof("')'"))
			return ast.FuncBody{}, false
		}
		if !next.Is(token.ParRight) {
			p.collectError(ErrUnexpectedThing("')'", next))
			return ast.FuncBody{}, false
		}
	}

	block := p.block()
	if block == nil {
		p.collectError(ErrExpectedSomething("block"))
		return ast.FuncBody{}, false
	}

	next, ok = p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("end"))
		return ast.FuncBody{}, false
	}
	if !next.Is(token.End) {
		p.collectError(ErrUnexpectedThing("end", next))
		return ast.FuncBody{}, false
	}

	return ast.FuncBody{
		ParList: parlist,
		Block:   block,
	}, true
}

func (p *parser) parlist() (ast.ParList, bool) {
	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("name or '...'"))
		return ast.ParList{}, false
	}
	if !next.Is(token.Name) && !next.Is(token.Ellipsis) {
		p.collectError(ErrUnexpectedThing("name or '...'", next))
		return ast.ParList{}, false
	}

	if next.Is(token.Ellipsis) {
		return ast.ParList{
			Ellipsis: true,
		}, true
	}

	// stash the first name of the namelist that we read
	p.stash(next)

	namelist := p.namelist()
	if namelist == nil {
		p.collectError(ErrExpectedSomething("namelist"))
		return ast.ParList{}, false
	}

	next, ok = p.next()
	if !ok {
		return ast.ParList{
			NameList: namelist,
		}, true
	}
	if !next.Is(token.Comma) {
		p.stash(next)
		return ast.ParList{
			NameList: namelist,
		}, true
	}

	next, ok = p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("'...'"))
		return ast.ParList{}, false
	}
	if !next.Is(token.Ellipsis) {
		p.collectError(ErrUnexpectedThing("'...'", next))
		return ast.ParList{}, false
	}
	return ast.ParList{
		NameList: namelist,
		Ellipsis: true,
	}, true
}

func (p *parser) local() ast.Local {
	localKeyword, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected 'local', but got EOF"))
	}
	if !localKeyword.Is(token.Local) {
		p.collectError(fmt.Errorf("expected 'local', but got %s", localKeyword))
	}

	namelist := p.namelist()

	// check if there's a '=' between namelist and explist
	assign, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected '=' followed by explist, but got EOF"))
		return ast.Local{}
	}
	if !assign.Is(token.Assign) {
		p.collectError(fmt.Errorf("expected '=' followed by explist, but got %s", assign))
		return ast.Local{}
	}

	explist := p.explist()

	return ast.Local{
		NameList: namelist,
		ExpList:  explist,
	}
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

func (p *parser) namelist() []token.Token {
	var list []token.Token
	for v, ok := p.next(); ok; v, ok = p.next() {
		if !v.Is(token.Name) {
			p.collectError(fmt.Errorf("expected name, but got %s", v))
			return list
		}
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
		prefixexp := p.prefixexp()
		if prefixexp == nil {
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
		prefixexp := p.prefixexp()
		if prefixexp == nil {
			p.collectError(fmt.Errorf("expected prefixexp, but got nothing"))
			return nil
		}
		return prefixexp
	}
	panic("implement function, tableconstructor, binary operation")
}

func (p *parser) prefixexp() ast.Exp {
	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("prefixexp"))
		return nil
	}

	var var_ ast.Var
	var exp ast.Exp

	switch {
	case next.Is(token.Name):
		p.stash(next)
		v, ok := p.var_()
		if !ok {
			p.collectError(ErrExpectedSomething("var"))
			return nil
		}
		var_ = v
	case next.Is(token.ParLeft):
		p.stash(next)
		e := p.exp()
		if e == nil {
			p.collectError(ErrExpectedSomething("exp"))
			return nil
		}
		rightPar, ok := p.next()
		if !ok {
			p.collectError(ErrUnexpectedEof("')'"))
			return nil
		}
		if !rightPar.Is(token.ParRight) {
			p.collectError(ErrUnexpectedThing("')'", rightPar))
			return nil
		}
		exp = e
	}

	next, ok = p.next()
	if !ok || !(next.Is(token.Colon) || next.Is(token.ParLeft) || next.Is(token.CurlyLeft) || next.Is(token.String)) {
		if ok {
			// only stash next if there is one
			p.stash(next)
		}
		return ast.PrefixExp{
			Var: var_,
			Exp: exp,
		}
	}
	p.stash(next)

	var functionCall ast.Statement
	// first is a flag with which we check if it's the first run of the loop.
	// We use this to check whether it's ok if no :<Name> or <args> follows.
	first := true
	for {
		next, ok = p.next()
		if !ok {
			if first {
				p.collectError(ErrUnexpectedEof("':', '(', '{' or String_"))
				return nil
			}
			break
		}
		if !next.Is(token.ParLeft) {
			p.stash(next)
			break
		}
		first = false

		var name token.Token
		if next.Is(token.Colon) {
			name, ok = p.next()
			if !ok {
				p.collectError(ErrUnexpectedEof("Name"))
				return nil
			}
			if !name.Is(token.Name) {
				p.collectError(ErrUnexpectedThing("Name", name))
				return nil
			}
		}
		p.stash(next)
		args, ok := p.args()
		if !ok {
			p.collectError(ErrExpectedSomething("args"))
			return nil
		}
		functionCall = ast.FunctionCall{
			PrefixExp: ast.PrefixExp{
				Var:          var_,
				Exp:          exp,
				FunctionCall: functionCall,
			},
			Name: name,
			Args: args,
		}
	}
	return ast.PrefixExp{
		FunctionCall: functionCall,
	}
}

func (p *parser) functionCall() (ast.FunctionCall, bool) {
	if next, ok := p.next(); ok {
		// next is either name or '('
		if !next.Is(token.Name) && !next.Is(token.ParLeft) {
			p.collectError(fmt.Errorf("expected either '(' or name, but got %s", next))
			return ast.FunctionCall{}, false
		}

		// functioncall starts with prefixexp, so we need to enter prefixexp anyways
		p.stash(next)
	} else {
		p.collectError(fmt.Errorf("expected prefixexp, but got EOF"))
		return ast.FunctionCall{}, false
	}

	prefixexp := p.prefixexp()
	if prefixexp == nil {
		p.collectError(fmt.Errorf("expected prefixexp, but got nothing"))
		return ast.FunctionCall{}, false
	}
	return prefixexp.(ast.PrefixExp).FunctionCall.(ast.FunctionCall), true
}

func (p *parser) args() (ast.Args, bool) {
	next, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected '(' or a name, but got EOF"))
		return ast.Args{}, false
	}
	switch {
	case next.Is(token.ParLeft):
		lookahead, ok := p.next()
		if !ok {
			p.collectError(fmt.Errorf("expected something after '(', but got EOF"))
			return ast.Args{}, false
		}
		if lookahead.Is(token.ParRight) {
			// no explist can follow, so immediately return an empty one
			return ast.Args{
				ExpList: []ast.Exp{},
			}, true
		}
		p.stash(lookahead)

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
	p.collectError(fmt.Errorf("expected one of '(', '{' or a String_, but got %s", next))
	return ast.Args{}, false
}
