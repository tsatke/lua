package token

import (
	"fmt"
	"strings"
)

// Token describes a Lua token.
// A token may have multiple types, e.g. '-' may be a
// unary and a binary operator.
type Token interface {
	Value() string
	Length() int
	Pos() Position
	Is(Type) bool
	Types() []Type
}

// Position describes the position of something in a file.
// Line and Col is 1-based, Offset is 0-based.
type Position struct {
	Line   int
	Col    int
	Offset int64
}

// New creates a new token from the given arguments.
func New(value string, pos Position, types ...Type) Token {
	return tok{
		value: value,
		pos:   pos,
		types: types,
	}
}

type tok struct {
	value string
	pos   Position
	types []Type
}

// Is determines whether this token has the given type.
func (t tok) Is(typ Type) bool {
	for _, gotTyp := range t.types {
		if gotTyp == typ {
			return true
		}
	}
	return false
}

// Value returns the token value.
func (t tok) Value() string {
	return t.value
}

// Length returns the length of the token Value.
func (t tok) Length() int {
	return len(t.value)
}

func (t tok) Pos() Position {
	return t.pos
}

func (t tok) Types() []Type {
	return t.types
}

func (t tok) String() string {
	return fmt.Sprintf("(%s) %q (types=%v)", t.Pos(), t.Value(), t.Types())
}

func (t tok) GoString() string {
	types := make([]string, len(t.types))
	for i, typ := range t.types {
		types[i] = "token." + typ.String()
	}
	return fmt.Sprintf(`token.New("%s", token.Position{%d, %d, %d}, %s)`, t.value, t.pos.Line, t.pos.Col, t.pos.Offset, strings.Join(types, ", "))
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d,offset=%d", p.Line, p.Col, p.Offset)
}
