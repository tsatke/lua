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
	next, ok := p.next()
	if ok {
		switch {
		case next.Is(token.Return):
			block = append(block, ast.LastStatement{
				ExpList: p.explist(),
			})
		case next.Is(token.Break):
			block = append(block, ast.LastStatement{
				Break: true,
			})
		default:
			p.stash(next)
		}
	}
	return block
}

func (p *parser) stmt() ast.Statement {
	defer func() {
		// optional semicolon after a statement
		next, ok := p.next()
		if !ok {
			return
		}
		if !next.Is(token.SemiColon) {
			p.stash(next)
		}
		return
	}()

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

		var assignment bool
		if next.Is(token.Assign) {
			assignment = true
		} else if next.Is(token.ParLeft) {
			assignment = false
		} else {
			// lookahead for either '=' or '(', and parse an assignment or a functioncall respectively
			var lookahead []token.Token
			for {
				next, ok := p.next()
				if !ok {
					p.collectError(ErrUnexpectedEof("anything"))
					return nil
				}
				lookahead = append(lookahead, next)
				if next.Is(token.Assign) {
					p.stash(lookahead...)
					assignment = true
					break
				} else if next.Is(token.ParLeft) {
					p.stash(lookahead...)
					break
				}
			}
		}

		p.stash(tk, next)
		switch assignment {
		case true:
			return p.assignment()
		case false:
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
	case tk.Is(token.If):
		p.stash(tk)
		ifBlock, ok := p.if_()
		if !ok {
			p.collectError(fmt.Errorf("expected if block, but got nothing"))
			return nil
		}
		return ifBlock
	case tk.Is(token.Do):
		p.stash(tk)
		doBlock, ok := p.do()
		if !ok {
			p.collectError(fmt.Errorf("expected do block, but got nothing"))
			return nil
		}
		return doBlock
	}
	p.stash(tk)
	return nil
}

func (p *parser) do() (ast.DoBlock, bool) {
	if !p.requireToken(token.Do) {
		return ast.DoBlock{}, false
	}

	block := p.block()
	if block == nil {
		p.collectError(ErrExpectedSomething("block"))
		return ast.DoBlock{}, false
	}

	if !p.requireToken(token.End) {
		return ast.DoBlock{}, false
	}

	return ast.DoBlock{
		Do: block,
	}, true
}

func (p *parser) if_() (ast.IfBlock, bool) {
	if !p.requireToken(token.If) {
		return ast.IfBlock{}, false
	}

	exp := p.exp()
	if exp == nil {
		p.collectError(ErrExpectedSomething("exp"))
		return ast.IfBlock{}, false
	}

	if !p.requireToken(token.Then) {
		return ast.IfBlock{}, false
	}

	block := p.block()
	if block == nil {
		p.collectError(ErrExpectedSomething("block"))
		return ast.IfBlock{}, false
	}

	next, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected %s, %s or %s, but got %v", token.Elseif, token.Else, token.End, next))
		return ast.IfBlock{}, false
	}
	if next.Is(token.Elseif) || next.Is(token.Else) {
		p.collectError(fmt.Errorf("elseif and else is not supported yet"))
		return ast.IfBlock{}, false
	}
	if !next.Is(token.End) {
		p.collectError(ErrUnexpectedThing(token.End, next))
		return ast.IfBlock{}, false
	}

	return ast.IfBlock{
		If:   exp,
		Then: block,
	}, true
}

// requireToken obtains the next token (using next()) and checks its type
// against the given type. If they are equal, true is returned.
// Otherwise, an error is collected and false is returned.
//
// THE OFFENDING TOKEN IS NOT STASHED.
func (p *parser) requireToken(typ token.Type) bool {
	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof(typ))
		return false
	}
	if !next.Is(typ) {
		p.collectError(ErrUnexpectedThing(typ, next))
		return false
	}
	return true
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
	prefixexp := p.prefixexp()
	if prefixexp == nil {
		p.collectError(ErrExpectedSomething("prefixexp"))
		return ast.Var{}, false
	}
	return ast.Var{
		PrefixExp: prefixexp.(ast.PrefixExp),
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

func (p *parser) exp() (exp ast.Exp) {
	next, ok := p.next()
	if !ok {
		return nil
	}
	switch {
	case next.Is(token.Nil):
		exp = ast.SimpleExp{
			Nil: next,
		}
	case next.Is(token.False):
		exp = ast.SimpleExp{
			False: next,
		}
	case next.Is(token.True):
		exp = ast.SimpleExp{
			True: next,
		}
	case next.Is(token.Ellipsis):
		exp = ast.SimpleExp{
			Ellipsis: next,
		}
	case next.Is(token.Number):
		exp = ast.SimpleExp{
			Number: next,
		}
	case next.Is(token.String):
		exp = ast.SimpleExp{
			String: next,
		}
	case next.Is(token.UnaryOperator):
		exp := p.exp()
		if exp == nil {
			p.collectError(fmt.Errorf("expected expression after unary operator %s, but got nothing", next))
			return nil
		}
		exp = ast.UnopExp{
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
		exp = prefixexp
	}
	if exp == nil {
		p.collectError(fmt.Errorf("function, prefixexp and tableconstructor are not supported yet (%s)", next))
		return
	}

	// check lookahead for binary operation expression

	lookahead, ok := p.next()
	if !ok {
		return
	}
	if lookahead.Is(token.BinaryOperator) {
		nextExp := p.exp()
		if nextExp == nil {
			p.collectError(ErrExpectedSomething("exp after binary operator"))
			return nil
		}
		exp = ast.BinopExp{
			Left:  exp,
			Binop: lookahead,
			Right: nextExp,
		}
	} else {
		p.stash(lookahead)
	}
	return
}

func (p *parser) prefixexp() ast.Exp {
	var name1 token.Token
	var exp1 ast.Exp

	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("name or '('"))
		return nil
	}
	switch {
	case next.Is(token.Name):
		name1 = next
	case next.Is(token.ParLeft):
		exp := p.exp()
		if exp == nil {
			p.collectError(ErrExpectedSomething("exp"))
			return nil
		}
		if !p.requireToken(token.ParRight) {
			return nil
		}
		exp1 = exp
	}

	next, ok = p.next()
	if !ok || !(next.Is(token.Colon) || next.Is(token.ParLeft) || next.Is(token.CurlyLeft) || next.Is(token.String) || next.Is(token.Dot) || next.Is(token.BracketLeft)) {
		if ok {
			// only stash next if there is one
			p.stash(next)
		}
		return ast.PrefixExp{
			Name: name1,
			Exp:  exp1,
		}
	}
	p.stash(next)

	var fragments []ast.PrefixExpFragment

fragments:
	for {
		next, ok := p.next()
		if !ok {
			break
		}
		switch {
		case next.Is(token.Dot):
			name, ok := p.next()
			if !ok {
				p.collectError(ErrUnexpectedEof("name"))
				return nil
			}
			if !name.Is(token.Name) {
				p.collectError(ErrUnexpectedThing("name", name))
				return nil
			}
			fragments = append(fragments, ast.PrefixExpFragment{
				Name: name,
			})
		case next.Is(token.BracketLeft):
			exp := p.exp()
			if exp == nil {
				p.collectError(ErrExpectedSomething("exp"))
				return nil
			}
			if !p.requireToken(token.BracketRight) {
				return nil
			}
			fragments = append(fragments, ast.PrefixExpFragment{
				Exp: exp,
			})
		case next.Is(token.Colon):
			name, ok := p.next()
			if !ok {
				p.collectError(ErrUnexpectedEof("name"))
				return nil
			}
			args, ok := p.args()
			if !ok {
				p.collectError(ErrExpectedSomething("args"))
				return nil
			}
			fragments = append(fragments, ast.PrefixExpFragment{
				Name: name,
				Args: &args,
			})
		case next.Is(token.ParLeft):
			p.stash(next)
			args, ok := p.args()
			if !ok {
				p.collectError(ErrExpectedSomething("args"))
				return nil
			}
			fragments = append(fragments, ast.PrefixExpFragment{
				Args: &args,
			})
		default:
			p.stash(next)
			break fragments
		}
	}
	return ast.PrefixExp{
		Name:      name1,
		Exp:       exp1,
		Fragments: fragments,
	}
}

func (p *parser) functionCall() (ast.FunctionCall, bool) {
	prefixexp := p.prefixexp()
	if prefixexp == nil {
		p.collectError(fmt.Errorf("expected prefixexp, but got nothing"))
		return ast.FunctionCall{}, false
	}
	return ast.FunctionCall{
		PrefixExp: prefixexp.(ast.PrefixExp),
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
