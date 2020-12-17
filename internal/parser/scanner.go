package parser

import "github.com/tsatke/lua/internal/token"

type scanner interface {
	next() (token.Token, bool)
	tkpos() token.Position
}
