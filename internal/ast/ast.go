package ast

import "github.com/tsatke/lua/internal/token"

type (
	// Chunk is just a Block, but is expected to run in a separate scope.
	Chunk struct {
		Name string
		Block
	}

	// Block is a list of statements. At least one statement is present.
	// The last Statement may be of type LastStatement.
	Block []Statement

	// Var is a variable declaration without assignment.
	Var struct {
		Name      token.Token
		Exp       Exp
		PrefixExp Exp
	}

	// Field is a field in a table constructor.
	Field struct {
		LeftExp  Exp
		LeftName token.Token
		RightExp Exp
	}

	// Args is an argument list when calling a function.
	Args struct {
		ExpList          []Exp
		TableConstructor ComplexExp
		String           token.Token
	}
)

// StatementsWithoutLast returns all statements in this Block that are not a
// LastStatement.
func (b Block) StatementsWithoutLast() []Statement {
	if _, hasLast := b.LastStatement(); hasLast {
		return b[:len(b)-1]
	}
	return b
}

// LastStatement returns the LastStatement in this Block, or false, if there
// is no LastStatement in this Block.
func (b Block) LastStatement() (Statement, bool) {
	if len(b) > 0 {
		if lastStatement, ok := b[len(b)-1].(LastStatement); ok {
			return lastStatement, true
		}
	}
	return nil, false
}
