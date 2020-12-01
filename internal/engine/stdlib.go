package engine

import "github.com/tsatke/lua/internal/engine/value"

func (e *Engine) initStdlib() {
	e.registerFunction(e.globalScope, value.NewFunction("print", e.print))
	e.registerFunction(e.globalScope, value.NewFunction("error", e.error))
}

func (e *Engine) registerFunction(scope *Scope, fn *value.Function) {
	e.assign(scope, fn.Name, fn)
}

func (e *Engine) print(args ...value.Value) (value.Value, error) {
	for i := 0; i < len(args); i++ {
		if i != 0 {
			_, _ = e.stdout.Write([]byte("\t"))
		}
		_, _ = e.stdout.Write([]byte(toString(args[i])))
	}
	_, _ = e.stdout.Write([]byte{0x0a})
	return value.Nil, nil
}

func (e *Engine) error(args ...value.Value) (value.Value, error) {
	if len(args) > 0 {
		panic(error_{
			message: args[0],
		})
		return nil, nil // unreachable
	}
	panic(error_{})
}
