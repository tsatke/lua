package engine

import "github.com/tsatke/lua/internal/engine/value"

type error_ struct {
	message value.Value
	level   value.Value
}
