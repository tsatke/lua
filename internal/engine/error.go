package engine

import (
	"bytes"
	"fmt"
	"github.com/tsatke/lua/internal/engine/value"
)

// Error represents a value originating from Lua's error() function.
type Error struct {
	e       *Engine
	Message value.Value
	Level   value.Value
	Stack   []StackFrame
}

func (e Error) Is(target error) bool {
	_, ok := target.(Error)
	return ok
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

func (e Error) String() string {
	if e.Message == nil {
		return "error called with <nil>"
	}
	res, err := e.e.tostring(e.Message)
	if err != nil {
		panic(err)
	}
	msg := string(res[0].(value.String))

	var buf bytes.Buffer
	var level int
	if e.Level != nil {
		level = int(e.Level.(value.Number))
	}
	if len(e.Stack) > level {
		buf.WriteString(e.Stack[level].Name)
		buf.WriteString(": ")
	}
	buf.WriteString(msg)
	for i, frame := range e.Stack {
		buf.WriteString(fmt.Sprintf("\n\t%d %s", i, frame.String()))
	}
	return buf.String()
}
