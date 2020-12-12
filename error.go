package lua

import (
	"github.com/tsatke/lua/internal/engine"
	"github.com/tsatke/lua/internal/engine/value"
)

type Error struct {
	Message string
	Stack   []StackFrame
}

type StackFrame struct {
	Name string
}

func errorFromInternal(err engine.Error) Error {
	e := Error{}
	e.Message = err.Message.(value.String).String()
	e.Stack = make([]StackFrame, len(err.Stack))
	for i, frame := range err.Stack {
		e.Stack[i] = StackFrame{
			Name: frame.Name,
		}
	}
	return e
}

func (e Error) Error() string {
	return e.Message
}
