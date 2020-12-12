package engine

import "github.com/tsatke/lua/internal/engine/value"

type Error struct {
	e       Engine
	Message value.Value
	level   value.Value
	Stack   []StackFrame
}

func (e Error) Error() string {
	if e.Message == nil {
		return "error called with <nil>"
	}
	res, err := e.e.tostring(e.Message)
	if err != nil {
		panic(err)
	}
	return string(res[0].(value.String))
}
