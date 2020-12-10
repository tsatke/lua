package engine

import "github.com/tsatke/lua/internal/engine/value"

type error_ struct {
	e       Engine
	message value.Value
	level   value.Value
	stack   []stackFrame
}

func (e error_) Error() string {
	if e.message == nil {
		return "error called with <nil>"
	}
	res, err := e.e.tostring(e.message)
	if err != nil {
		panic(err)
	}
	return string(res[0].(value.String))
}
