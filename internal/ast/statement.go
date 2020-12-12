package ast

import "github.com/tsatke/lua/internal/token"

type (
	// Statement is a Lua statement.
	Statement interface {
		_stmt()
	}

	// Assignment is a Lua assignment. An expression list is assigned to
	// a variable list.
	Assignment struct {
		VarList []Var
		ExpList []Exp
	}

	// FunctionCall is a Lua function call.
	FunctionCall struct {
		PrefixExp PrefixExp
	}

	// DoBlock is a Lua do construct.
	DoBlock struct {
		Do Block
	}

	// WhileBlock is a Lua while construct.
	WhileBlock struct {
		While Exp
		Do    Block
	}

	// RepeatBlock is a Lua repeat construct.
	RepeatBlock struct {
		Repeat Block
		Until  Exp
	}

	// IfBlock is a Lua if construct.
	IfBlock struct {
		If     Exp
		Then   Block
		ElseIf []ElseIf
		Else   Block
	}

	// ElseIf is an IfBlock, that was produced by another
	// grammar production.
	ElseIf IfBlock

	// ForBlock is a Lua for construct, using the production from,to,step.
	ForBlock struct {
		Name token.Token
		From Exp
		To   Exp
		Step Exp
		Do   Block
	}

	// ForInBlock is a Lua for in construct.
	ForInBlock struct {
		NameList []token.Token
		In       []Exp
		Do       Block
	}

	// Function is a Lua function.
	Function struct {
		FuncName *FuncName
		FuncBody FuncBody
	}

	// FuncBody is a Lua function body, including the parameter list.
	FuncBody struct {
		ParList ParList
		Block   Block
	}

	// ParList is a function parameter list.
	ParList struct {
		NameList []token.Token
		Ellipsis bool
	}

	// FuncName is a function name, e.g. a.b.c:d or just foo.bar.
	FuncName struct {
		Name1 []token.Token
		Name2 token.Token
	}

	// LocalFunction is a Lua local function construct.
	LocalFunction struct {
		Name     token.Token
		FuncBody FuncBody
	}

	// Local is a Lua local construct.
	Local struct {
		NameList []token.Token
		ExpList  []Exp
	}

	// LastStatement is a Lua last statement in a Block.
	LastStatement struct {
		ExpList []Exp
	}
)

func (Assignment) _stmt()    {}
func (FunctionCall) _stmt()  {}
func (DoBlock) _stmt()       {}
func (WhileBlock) _stmt()    {}
func (RepeatBlock) _stmt()   {}
func (IfBlock) _stmt()       {}
func (ForBlock) _stmt()      {}
func (ForInBlock) _stmt()    {}
func (Function) _stmt()      {}
func (LocalFunction) _stmt() {}
func (Local) _stmt()         {}
func (LastStatement) _stmt() {}
