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

	next, ok := p.next()
	if ok {
		// not all tokens consumed
		p.collectError(ErrUnexpectedThing("a statement or eof", next))
		return ast.Chunk{}, false
	}

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

func (p *parser) stmt() (stmt ast.Statement) {
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
		p.stash(tk)
		prefixexp := p.prefixexp()
		if prefixexp == nil {
			p.collectError(ErrExpectedSomething("prefixexp"))
			return nil
		}

		next, ok := p.next()
		if !ok {
			return ast.FunctionCall{
				PrefixExp: prefixexp.(ast.PrefixExp),
			}
		}
		if !(next.Is(token.Assign) || next.Is(token.Comma)) {
			p.stash(next)
			return ast.FunctionCall{
				PrefixExp: prefixexp.(ast.PrefixExp),
			}
		}

		varlist := []ast.Var{
			{
				PrefixExp: prefixexp.(ast.PrefixExp),
			},
		}
		if next.Is(token.Comma) {
			remainingVarList := p.varlist()
			if len(remainingVarList) == 0 {
				p.collectError(ErrExpectedSomething("varlist"))
				return nil
			}
			varlist = append(varlist, remainingVarList...)
		} else {
			p.stash(next)
		}

		assignment := p.assignmentWithVarlist(varlist)
		return assignment
	case tk.Is(token.Local):
		next, ok := p.next()
		if !ok {
			p.collectError(ErrUnexpectedEof("'function' or a name"))
			return nil
		}

		p.stash(tk, next)
		switch {
		case next.Is(token.Name):
			local, ok := p.local()
			if !ok {
				p.collectError(ErrExpectedSomething("local"))
				return nil
			}
			return local
		case next.Is(token.Function):
			localFn, ok := p.localFunction()
			if !ok {
				p.collectError(ErrExpectedSomething("local"))
				return nil
			}
			return localFn
		}
	case tk.Is(token.Function):
		p.stash(tk)
		fn, ok := p.function()
		if !ok {
			p.collectError(ErrExpectedSomething("function"))
			return nil
		}
		return fn
	case tk.Is(token.If):
		p.stash(tk)
		ifBlock, ok := p.if_()
		if !ok {
			p.collectError(ErrExpectedSomething("if block"))
			return nil
		}
		return ifBlock
	case tk.Is(token.Do):
		p.stash(tk)
		doBlock, ok := p.do()
		if !ok {
			p.collectError(ErrExpectedSomething("do block"))
			return nil
		}
		return doBlock
	case tk.Is(token.Repeat):
		p.stash(tk)
		repeatBlock, ok := p.repeat()
		if !ok {
			p.collectError(ErrExpectedSomething("repeat block"))
			return nil
		}
		return repeatBlock
	case tk.Is(token.While):
		p.stash(tk)
		whileBlock, ok := p.while()
		if !ok {
			p.collectError(ErrExpectedSomething("while block"))
			return nil
		}
		return whileBlock
	case tk.Is(token.For):
		l1, ok := p.next()
		if !ok {
			p.collectError(ErrUnexpectedEof("name"))
			return nil
		}
		if !l1.Is(token.Name) {
			p.collectError(ErrUnexpectedThing("name", l1))
			return nil
		}
		l2, ok := p.next()
		if !ok {
			p.collectError(ErrUnexpectedEof("comma or '='"))
			return nil
		}
		if !l2.Is(token.Comma) && !l2.Is(token.In) && !l2.Is(token.Assign) {
			p.collectError(ErrUnexpectedThing("comma, 'in' or '='", l2))
			return nil
		}

		forIn := l2.Is(token.Comma) || l2.Is(token.In)

		p.stash(tk, l1, l2)
		if forIn {
			forInBlock, ok := p.forInBlock()
			if !ok {
				p.collectError(ErrExpectedSomething("for-in block"))
				return nil
			}
			return forInBlock
		}

		forBlock, ok := p.forBlock()
		if !ok {
			p.collectError(ErrExpectedSomething("for block"))
			return nil
		}
		return forBlock
	}
	p.stash(tk)
	return nil
}

func (p *parser) forInBlock() (ast.ForInBlock, bool) {
	if !p.requireToken(token.For) {
		return ast.ForInBlock{}, false
	}

	var nameList []token.Token

	for {
		name, ok := p.next()
		if !ok {
			p.collectError(ErrUnexpectedEof("name"))
			return ast.ForInBlock{}, false
		}
		if !name.Is(token.Name) {
			p.collectError(ErrUnexpectedThing("name", name))
			return ast.ForInBlock{}, false
		}
		nameList = append(nameList, name)

		comma, ok := p.next()
		if !ok {
			p.collectError(ErrUnexpectedEof("comma or 'in'"))
			return ast.ForInBlock{}, false
		}
		if !comma.Is(token.Comma) {
			p.stash(comma)
			break
		}
	}

	if !p.requireToken(token.In) {
		return ast.ForInBlock{}, false
	}

	explist := p.explist()
	if explist == nil {
		p.collectError(ErrExpectedSomething("explist"))
		return ast.ForInBlock{}, false
	}

	if !p.requireToken(token.Do) {
		return ast.ForInBlock{}, false
	}

	block := p.block()
	if block == nil {
		p.collectError(ErrExpectedSomething("block"))
		return ast.ForInBlock{}, false
	}

	if !p.requireToken(token.End) {
		return ast.ForInBlock{}, false
	}

	return ast.ForInBlock{
		NameList: nameList,
		In:       explist,
		Do:       block,
	}, true
}

func (p *parser) forBlock() (ast.ForBlock, bool) {
	if !p.requireToken(token.For) {
		return ast.ForBlock{}, false
	}

	name, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("name"))
		return ast.ForBlock{}, false
	}
	if !name.Is(token.Name) {
		p.collectError(ErrUnexpectedThing("name", name))
		return ast.ForBlock{}, false
	}

	if !p.requireToken(token.Assign) {
		return ast.ForBlock{}, false
	}

	from := p.exp()
	if from == nil {
		p.collectError(ErrExpectedSomething("exp (from)"))
		return ast.ForBlock{}, false
	}

	if !p.requireToken(token.Comma) {
		return ast.ForBlock{}, false
	}

	to := p.exp()
	if to == nil {
		p.collectError(ErrExpectedSomething("exp (to)"))
		return ast.ForBlock{}, false
	}

	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("comma or 'do'"))
		return ast.ForBlock{}, false
	}

	var step ast.Exp

	if next.Is(token.Comma) {
		step = p.exp()
		if step == nil {
			p.collectError(ErrUnexpectedEof("exp (step)"))
			return ast.ForBlock{}, false
		}
	} else {
		p.stash(next)
	}

	if !p.requireToken(token.Do) {
		return ast.ForBlock{}, false
	}

	block := p.block()
	if block == nil {
		p.collectError(ErrExpectedSomething("block"))
		return ast.ForBlock{}, false
	}

	if !p.requireToken(token.End) {
		return ast.ForBlock{}, false
	}

	return ast.ForBlock{
		Name: name,
		From: from,
		To:   to,
		Step: step,
		Do:   block,
	}, true
}

func (p *parser) localFunction() (ast.LocalFunction, bool) {
	if !p.requireToken(token.Local) {
		return ast.LocalFunction{}, false
	}
	if !p.requireToken(token.Function) {
		return ast.LocalFunction{}, false
	}

	name, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("name"))
		return ast.LocalFunction{}, false
	}
	if !name.Is(token.Name) {
		p.collectError(ErrUnexpectedThing("name", name))
		return ast.LocalFunction{}, false
	}

	body, ok := p.funcbody()
	if !ok {
		p.collectError(ErrExpectedSomething("funcbody"))
		return ast.LocalFunction{}, false
	}

	return ast.LocalFunction{
		Name:     name,
		FuncBody: body,
	}, true
}

func (p *parser) while() (ast.WhileBlock, bool) {
	if !p.requireToken(token.While) {
		return ast.WhileBlock{}, false
	}

	exp := p.exp()
	if exp == nil {
		p.collectError(ErrExpectedSomething("exp"))
		return ast.WhileBlock{}, false
	}

	if !p.requireToken(token.Do) {
		return ast.WhileBlock{}, false
	}

	block := p.block()
	if block == nil {
		p.collectError(ErrExpectedSomething("block"))
		return ast.WhileBlock{}, false
	}

	if !p.requireToken(token.End) {
		return ast.WhileBlock{}, false
	}

	return ast.WhileBlock{
		While: exp,
		Do:    block,
	}, true
}

func (p *parser) repeat() (ast.RepeatBlock, bool) {
	if !p.requireToken(token.Repeat) {
		return ast.RepeatBlock{}, false
	}

	block := p.block()
	if block == nil {
		p.collectError(ErrExpectedSomething("block"))
		return ast.RepeatBlock{}, false
	}

	if !p.requireToken(token.Until) {
		return ast.RepeatBlock{}, false
	}

	exp := p.exp()
	if exp == nil {
		p.collectError(ErrExpectedSomething("exp"))
		return ast.RepeatBlock{}, false
	}

	return ast.RepeatBlock{
		Repeat: block,
		Until:  exp,
	}, true
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
		p.collectError(ErrUnexpectedEof("elseif, else or end"))
		return ast.IfBlock{}, false
	}
	if next.Is(token.Elseif) {
		p.collectError(fmt.Errorf("elseif is not supported yet"))
		return ast.IfBlock{}, false
	}

	var elseBlock ast.Block
	if next.Is(token.Else) {
		elseBlock = p.block()
	} else {
		p.stash(next)
	}

	next, ok = p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("end"))
		return ast.IfBlock{}, false
	}
	if !next.Is(token.End) {
		p.collectError(ErrUnexpectedThing("end", next))
		return ast.IfBlock{}, false
	}

	return ast.IfBlock{
		If:   exp,
		Then: block,
		Else: elseBlock,
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
	if !p.requireToken(token.Function) {
		return ast.Function{}, false
	}

	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("name or body"))
		return ast.Function{}, false
	}
	var name *ast.FuncName
	p.stash(next)
	if next.Is(token.Name) {
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

func (p *parser) local() (ast.Local, bool) {
	localKeyword, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected 'local', but got EOF"))
		return ast.Local{}, false
	}
	if !localKeyword.Is(token.Local) {
		p.collectError(fmt.Errorf("expected 'local', but got %s", localKeyword))
		return ast.Local{}, false
	}

	namelist := p.namelist()

	// check if there's a '=' between namelist and explist
	assign, ok := p.next()
	if !ok {
		p.collectError(fmt.Errorf("expected '=' followed by explist, but got EOF"))
		return ast.Local{}, false
	}
	if !assign.Is(token.Assign) {
		p.collectError(fmt.Errorf("expected '=' followed by explist, but got %s", assign))
		return ast.Local{}, false
	}

	explist := p.explist()

	return ast.Local{
		NameList: namelist,
		ExpList:  explist,
	}, true
}

func (p *parser) assignmentWithVarlist(varlist []ast.Var) ast.Assignment {
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

func (p *parser) assignment() ast.Assignment {
	varlist := p.varlist()
	return p.assignmentWithVarlist(varlist)
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

func (p *parser) exp() ast.Exp {
	return p.expPrecedence(p.expAtomic(), precedence0)
}

func (p *parser) expAtomic() (exp ast.Exp) {
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
		exp = p.exp()
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
	case next.Is(token.Function):
		p.stash(next)
		fn, ok := p.function()
		if !ok {
			p.collectError(ErrExpectedSomething("function"))
			return nil
		}
		if fn.FuncName != nil {
			p.collectError(fmt.Errorf("function cannot have a name if it's an expression"))
			return nil
		}
		exp = fn
	case next.Is(token.CurlyLeft):
		p.stash(next)
		tbl, ok := p.tableconstructor()
		if !ok {
			p.collectError(ErrExpectedSomething("table"))
			return nil
		}
		exp = tbl
	}
	if exp == nil {
		p.collectError(ErrUnexpectedThing("either 'nil', 'false', 'true', '...', a number, a string, a unary operator, a name, 'function', '(' or '{'", next))
		return
	}
	return
}

func (p *parser) expPrecedence(lhs ast.Exp, minPrecedence precedence) ast.Exp {
	lookahead, ok := p.next()
	if !ok {
		return lhs
	}
	p.stash(lookahead)

	for lookahead != nil && lookahead.Is(token.BinaryOperator) && precedenceOf(lookahead.Value()) >= minPrecedence {
		op := lookahead

		_, ok = p.next()
		if !ok {
			p.collectError(ErrUnexpectedEof("right hand side of expression"))
			return nil
		}

		rhs := p.expAtomic()

		lookahead, ok = p.next()
		if !ok {
			return ast.BinopExp{
				Left:  lhs,
				Binop: op,
				Right: rhs,
			}
		}
		p.stash(lookahead)

		for lookahead.Is(token.BinaryOperator) && ((precedenceOf(lookahead.Value()) > precedenceOf(op.Value())) ||
			(isRightAssociative(lookahead.Value()) && (precedenceOf(lookahead.Value()) == precedenceOf(op.Value())))) {
			rhs = p.expPrecedence(rhs, precedenceOf(lookahead.Value()))
			lookahead, ok = p.next()
			if !ok {
				break
			}
			p.stash(lookahead)
		}
		lhs = ast.BinopExp{
			Left:  lhs,
			Binop: op,
			Right: rhs,
		}
	}

	return lhs
}

func (p *parser) tableconstructor() (ast.TableConstructor, bool) {
	if !p.requireToken(token.CurlyLeft) {
		return ast.TableConstructor{}, false
	}

	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("field or '}'"))
		return ast.TableConstructor{}, false
	}
	if next.Is(token.CurlyRight) {
		return ast.TableConstructor{}, true
	}
	p.stash(next)

	fields := p.fieldlist()

	if !p.requireToken(token.CurlyRight) {
		return ast.TableConstructor{}, false
	}

	return ast.TableConstructor{
		Fields: fields,
	}, true
}

func (p *parser) fieldlist() []ast.Field {
	var fields []ast.Field

	for {
		field, ok := p.field()
		if !ok {
			if len(fields) == 0 {
				break
			}
			p.collectError(ErrExpectedSomething("field"))
			return nil
		}
		fields = append(fields, field)

		next, ok := p.next()
		if !ok {
			break
		}
		if !(next.Is(token.Comma) || next.Is(token.SemiColon)) {
			p.stash(next)
			break
		}
	}

	return fields
}

func (p *parser) field() (ast.Field, bool) {
	next, ok := p.next()
	if !ok {
		p.collectError(ErrUnexpectedEof("name, '[' or exp"))
		return ast.Field{}, false
	}

	var name token.Token
	var leftExp ast.Exp

	switch {
	case next.Is(token.Name):
		name = next
	case next.Is(token.BracketLeft):
		leftExp = p.exp()
		if leftExp == nil {
			p.collectError(ErrExpectedSomething("exp"))
			return ast.Field{}, false
		}

		if !p.requireToken(token.BracketRight) {
			return ast.Field{}, false
		}
	default:
		p.stash(next)
	}

	if name != nil || leftExp != nil {
		if !p.requireToken(token.Assign) {
			return ast.Field{}, false
		}
	}

	rightExp := p.exp()
	if rightExp == nil {
		p.collectError(ErrExpectedSomething("exp"))
		return ast.Field{}, false
	}

	return ast.Field{
		LeftExp:  leftExp,
		LeftName: name,
		RightExp: rightExp,
	}, true
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
