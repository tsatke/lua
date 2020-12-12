package ast

import "github.com/tsatke/lua/internal/token"

type (
	// Exp is a Lua expression. This is either a SimpleExp or a ComplexExp.
	Exp interface {
		_exp()
	}

	// SimpleExp is a simple and constant expression.
	SimpleExp struct {
		Nil      token.Token
		False    token.Token
		True     token.Token
		Number   token.Token
		String   token.Token
		Ellipsis token.Token
	}

	// ComplexExp is a not necessarily constant expression.
	ComplexExp interface {
		_exp()
	}

	// Function already declared in statement.go

	// PrefixExp is a prefix expression.
	PrefixExp struct {
		Name token.Token
		Exp  Exp

		Fragments []PrefixExpFragment
	}

	PrefixExpFragment struct {
		Exp  Exp
		Name token.Token
		Args *Args
	}

	// TableConstructor is a table constructor, which consists of
	// a list of Fields.
	TableConstructor struct {
		Fields []Field
	}

	// BinopExp is a binary expression.
	BinopExp struct {
		Left  Exp
		Binop token.Token
		Right Exp
	}

	// UnopExp is a unary expression.
	UnopExp struct {
		Unop token.Token
		Exp  Exp
	}
)

func (SimpleExp) _exp() {}

func (Function) _exp()         {}
func (PrefixExp) _exp()        {}
func (TableConstructor) _exp() {}
func (BinopExp) _exp()         {}
func (UnopExp) _exp()          {}
