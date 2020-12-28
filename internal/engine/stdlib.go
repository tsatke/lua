package engine

import (
	"errors"
	"fmt"
	"math"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"

	. "github.com/tsatke/lua/internal/engine/value"
)

func (e *Engine) initStdlib() {
	register := func(fn *Function) {
		e.assign(e._G, fn.Name, fn)
	}
	e.assign(e._G, "_VERSION", NewString("Lua 5.3"))
	register(NewFunction("assert", e.assert))
	register(NewFunction("collectgarbage", e.collectgarbage))
	register(NewFunction("dofile", e.dofile))
	register(NewFunction("error", e.error))
	register(NewFunction("getmetatable", e.getmetatable))
	register(NewFunction("ipairs", e.ipairs))
	register(NewFunction("pcall", e.pcall))
	register(NewFunction("print", e.print))
	register(NewFunction("rawget", e.rawget))
	register(NewFunction("select", e.select_))
	register(NewFunction("setmetatable", e.setmetatable))
	register(NewFunction("tostring", e.tostring))
	register(NewFunction("type", e.type_))
}

func (e *Engine) assert(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need at least one argument to 'assert'")
	}
	if e.isNil(args[0]) || args[0] == False {
		if len(args) > 1 {
			return e.error(args[1])
		}
		return e.error(NewString("assertion failed!"))
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
		return values(NewNumber(float64(m.HeapAlloc))), nil
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
			if luaErr, ok := err.(Error); ok {
				return nil, luaErr
			}
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

	file, err := e.fs.OpenFile(filename.(String).String(), os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() { _ = file.Close() }()

	results, err := e.Eval(file)
	if err != nil {
		if luaErr, ok := err.(Error); ok {
			return nil, luaErr
		}
		return nil, fmt.Errorf("eval: %w", err)
	}
	return results, nil
}

func (e *Engine) error(args ...Value) ([]Value, error) {
	var message Value
	var level Value
	var stack []StackFrame
	if len(args) > 0 {
		message = args[0]
	}
	if len(args) > 1 {
		level = args[1]
	}
	stack = e.stack.Slice()
	return nil, Error{
		e:       e,
		Message: message,
		Level:   level,
		Stack:   stack,
	}
}

func (e *Engine) getmetatable(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need one argument to 'getmetatable'")
	}

	val := args[0]
	if val.Type() == TypeTable {
		if metatable := val.(*Table).Metatable; metatable != nil {
			return values(metatable), nil
		}
		return values(Nil), nil
	}
	if metatable := e.metaTables.Table(val.Type()); metatable != nil {
		return values(metatable), nil
	}
	return nil, fmt.Errorf("no meta table for type %s", val.Type())
}

func (e *Engine) ipairs(args ...Value) ([]Value, error) {
	/*
		-- From the Lua spec.
		function iter (a, i)
		  i = i + 1
		  local v = a[i]
		  if v then
			return i, v
		  end
		end
	*/
	iter := func(args ...Value) ([]Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("need two arguments to 'iter'")
		}
		if _, ok := args[0].(*Table); !ok {
			return nil, fmt.Errorf("bad argument #1 to 'iter' (%s expected, got %s)", TypeTable, args[0].Type())
		}
		if _, ok := args[1].(Number); !ok {
			return nil, fmt.Errorf("bad argument #2 to 'iter' (%s expected, got %s)", TypeNumber, args[0].Type())
		}
		var a *Table
		var i Number
		a = args[0].(*Table)
		i = args[1].(Number)
		i = NewNumber(i.Value() + 1)
		v, ok := a.Get(i)
		if ok {
			return values(i, v), nil
		}
		return nil, nil
	}
	iterFn := NewFunction("iter", iter)
	return values(iterFn, args[0], NewNumber(0)), nil
}

func (e *Engine) pcall(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need one argument to 'pcall'")
	}

	results, err := func() (res []Value, recoveredErr error) {
		fn := args[0]
		if fn.Type() != TypeFunction {
			return nil, fmt.Errorf("bad argument to 'pcall' (%s expected, got %s)", TypeFunction, fn.Type())
		}
		fnVal := fn.(*Function)
		results, err := e.call(fnVal, args[1:]...)
		if err != nil {
			var luaErr Error
			if errors.As(err, &luaErr) {
				return nil, luaErr
			}

			// this happens if the call fails internally, not if 'error' has been called
			return nil, fmt.Errorf("call: %w", err)
		}
		return results, nil
	}()
	if err != nil {
		if luaErr, ok := err.(Error); ok {
			return values(False, luaErr.Message), nil
		}
		return nil, fmt.Errorf("protected: %w", err)
	}
	return append(values(True), results...), nil
}

func (e *Engine) print(args ...Value) ([]Value, error) {
	for i := 0; i < len(args); i++ {
		if i != 0 {
			_, _ = e.stdout.Write([]byte("\t"))
		}
		strs, err := e.tostring(args[i])
		if err != nil {
			return nil, fmt.Errorf("tostring: %w", err)
		}
		_, _ = e.stdout.Write([]byte(strs[0].(String)))
	}
	_, _ = e.stdout.Write([]byte{0x0a})
	return nil, nil
}

func (e *Engine) rawget(args ...Value) ([]Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("need exactly two arguments to 'rawget'")
	}

	var table *Table
	index := args[1]

	if tbl, ok := args[0].(*Table); ok {
		table = tbl
	} else {
		return nil, fmt.Errorf("bad argument #1 to 'rawget' (%s expected, got %s)", TypeTable, args[0].Type())
	}

	val, ok := table.Get(index)
	if !ok {
		return values(Nil), nil
	}
	return values(val), nil
}

func (e *Engine) select_(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need at least one argument to 'select'")
	}
	if str, ok := args[0].(String); ok && str == "#" {
		return values(NewNumber(float64(len(args) - 1))), nil
	} else if ok {
		return nil, fmt.Errorf("if the first argument to 'select' is a string, it must be the string '#'")
	}

	num, ok := args[0].(Number)
	if !ok {
		return nil, fmt.Errorf("bad argument #1 to 'select' (%s expected, got %s", TypeNumber, args[0].Type())
	}
	if float64(num) != math.Trunc(float64(num)) {
		return nil, fmt.Errorf("number %f has no integral representation", float64(num))
	}

	selection := int(num)
	if selection < 0 {
		return args[len(args)-selection:], nil
	}
	fromIndex := selection + 1
	if fromIndex > len(args) {
		fromIndex = len(args)
	}
	return args[fromIndex:], nil
}

func (e *Engine) setmetatable(args ...Value) ([]Value, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("need two arguments to 'setmetatable'")
	}

	if args[0].Type() != TypeTable {
		return nil, fmt.Errorf("bad argument #1 to 'setmetatable' (%s expected, got %s)", TypeTable, args[0].Type())
	}
	if args[1].Type() != TypeTable && args[1].Type() != TypeNil {
		return nil, fmt.Errorf("bad argument #2 to 'setmetatable' (%s expected, got %s)", TypeTable, args[1].Type())
	}
	metatable := args[0].(*Table).Metatable
	if metatable != nil {
		if _, ok := metatable.Get(NewString("__metatable")); ok {
			_, _ = e.error(NewString("original metatable has a __metatable field"))
		}
	}
	if args[1] == Nil {
		args[0].(*Table).Metatable = nil
	} else {
		args[0].(*Table).Metatable = args[1].(*Table)
	}

	return values(args[0]), nil
}

func (e *Engine) tostring(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need one argument to 'tostring'")
	}

	if args[0] == nil {
		return values(NewString("nil")), nil
	}

	value := args[0]
	switch value.Type() {
	case TypeNil:
		return values(NewString("nil")), nil
	case TypeBoolean:
		if !value.(Boolean) {
			return values(NewString("false")), nil
		}
		return values(NewString("true")), nil
	case TypeString:
		return values(value), nil
	case TypeFunction:
		return values(NewString("function " + value.(*Function).Name)), nil
	case TypeNumber:
		return values(NewString(strconv.FormatFloat(float64(value.(Number)), 'G', -1, 64))), nil
	}
	return nil, fmt.Errorf("unsupported type %s to 'tostring'", value.Type())
}

func (e *Engine) tonumber(args ...Value) ([]Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("need one argument to 'tonumber'")
	}

	if args[0] == nil {
		return values(Nil), nil
	}

	value := args[0]
	switch value.Type() {
	case TypeNumber:
		return values(value), nil
	case TypeString:
		num, err := strconv.ParseFloat(value.(String).String(), 64)
		if err != nil {
			return values(Nil), nil
		}
		return values(NewNumber(num)), nil
	}
	return values(Nil), nil
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
