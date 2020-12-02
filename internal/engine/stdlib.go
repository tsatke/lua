package engine

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	. "github.com/tsatke/lua/internal/engine/value"
)

func (e *Engine) initStdlib() {
	register := func(fn *Function) {
		e.assign(e.globalScope, fn.Name, fn)
	}
	e.assign(e.globalScope, "_VERSION", NewString("Lua 5.3"))
	register(NewFunction("assert", e.assert))
	register(NewFunction("dofile", e.dofile))
	register(NewFunction("error", e.error))
	register(NewFunction("print", e.print))
	register(NewFunction("tostring", e.tostring))
	register(NewFunction("type", e.type_))
}

func (e *Engine) assert(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need at least one argument to 'assert'")
	}
	if args[0] == Nil || args[0] == False {
		if len(args) > 1 {
			_, _ = e.error(args[1])
			return nil, nil // unreachable
		}
		_, _ = e.error(NewString("assertion failed!"))
		return nil, nil // unreachable
	}
	return args, nil
}

func (e *Engine) collectgarbage(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need at least one argument to 'collectgarbage'")
	}

	opt := args[0]
	if opt.Type() != TypeString {
		return nil, fmt.Errorf("bad argument to 'collectgarbage' (%s expected, got %s)", TypeString, opt.Type())
	}
	switch opt.(String).String() {
	case "collect":
		runtime.GC()
	case "stop":
		e.gcpercent = debug.SetGCPercent(-1)
		e.gcrunning = false
	case "restart":
		debug.SetGCPercent(e.gcpercent)
		e.gcrunning = true
	case "count":
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return values(NewString(fmt.Sprintf("%d", m.HeapAlloc))), nil // TODO this has to be a number
	case "step":
		runtime.GC()
		return values(True), nil
	case "setpause":
	case "setstepmul":
	case "isrunning":
		if e.gcrunning {
			return values(True), nil
		}
		return values(False), nil
	}
	return nil, fmt.Errorf("bad argument to 'collectgarbage' (invalid option '%s')", opt.(String).String())
}

func (e *Engine) dofile(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		// evaluate stdin if no args are given
		results, err := e.Eval(e.stdin)
		if err != nil {
			return nil, fmt.Errorf("eval stdin: %w", err)
		}
		return results, nil
	}

	filename := args[0]
	if filename.Type() != TypeString {
		if filename.Type() == TypeNumber {
			result, err := e.tostring(filename)
			if err != nil {
				return nil, fmt.Errorf("tostring: %w", err)
			}
			filename = result[0]
		} else {
			return nil, fmt.Errorf("bad argument to 'dofile' (%s expected, got %s)", TypeString, filename.Type())
		}
	}

	file, err := os.OpenFile(filename.(String).String(), os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = file.Close() }()

	results, err := e.Eval(file)
	if err != nil {
		return nil, fmt.Errorf("eval: %w", err)
	}
	return results, nil
}

func (e *Engine) error(args ...Value) ([]Value, error) {
	if len(args) > 0 {
		panic(error_{
			message: args[0],
		})
		return nil, nil // unreachable
	}
	panic(error_{})
}

func (e *Engine) print(args ...Value) ([]Value, error) {
	for i := 0; i < len(args); i++ {
		if i != 0 {
			_, _ = e.stdout.Write([]byte("\t"))
		}
		_, _ = e.stdout.Write([]byte(toString(args[i])))
	}
	_, _ = e.stdout.Write([]byte{0x0a})
	return values(Nil), nil
}

func (e *Engine) tostring(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need one argument to 'tostring'")
	}

	value := args[0]
	switch value.Type() {
	case TypeNil:
		return values(NewString("nil")), nil
	case TypeBoolean:
		if value.(Boolean) == 0 {
			return values(NewString("false")), nil
		}
		return values(NewString("true")), nil
	case TypeString:
		return values(value), nil
	case TypeFunction:
		return values(NewString("function " + value.(Function).Name)), nil
	}
	return nil, fmt.Errorf("unsupported type %s", value.Type())
}

func (e *Engine) type_(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need one argument to 'tostring'")
	}

	value := args[0]
	var typeName string
	switch value.Type() {
	case TypeInvalid:
		typeName = "<invalid>"
	case TypeNil:
		typeName = "nil"
	case TypeBoolean:
		typeName = "boolean"
	case TypeNumber:
		typeName = "number"
	case TypeString:
		typeName = "string"
	case TypeFunction:
		typeName = "function"
	case TypeUserdata:
		typeName = "userdata"
	case TypeThread:
		typeName = "thread"
	case TypeTable:
		typeName = "table"
	}
	return values(NewString(typeName)), nil
}

func values(vals ...Value) []Value {
	return vals
}
