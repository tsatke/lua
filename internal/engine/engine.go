package engine

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/engine/value"
	"github.com/tsatke/lua/internal/parser"
)

type Namer interface {
	Name() string
}

// Engine is an engine that is capable of evaluating Lua code from an io.Reader.
// The engine keeps track of state, so that multiple calls to Eval will build on the
// state that the last Eval call produced.
//
//	engine.Eval(strings.NewReader(`a=5`))
//	engine.Eval(strings.NewReader(`print(a)`)) // prints '5'
//
// If an error occurs during parsing or evaluation, that error will be returned. In case
// of a parse error, the state of the engine will remain unaffected.
type Engine struct {
	fs afero.Fs

	// stdin is the input for any program run by this engine.
	// If a program wants to read from stdin, this is the reader
	// that will be read from.
	stdin io.Reader
	// stdout is the output for any program run by this engine.
	// If a program wants to write to stdout, this is the writer
	// that will be written to.
	stdout io.Writer
	// stderr is the error output for any program run by this engine.
	// If a program wants to write to stderr, this is the writer
	// that will be written to.
	stderr io.Writer

	// clock is the clock that the engine will use if it requires a timestamp.
	clock Clock

	_G     *value.Table
	scopes []*value.Table

	metaTables metaTables

	gcrunning bool
	gcpercent int

	stack *callStack
}

// New creates a new, ready to use Engine, already applying all given options.
// By default, the engine uses os.Stdin as stdin, os.Stdout as stdout and os.Stderr
// as stderr.
func New(opts ...Option) *Engine {
	global := value.NewTable()
	e := &Engine{
		fs: afero.NewOsFs(),

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
		clock:  sysClock{},

		_G:     global,
		scopes: []*value.Table{global},

		stack: newCallStack(),
	}
	for _, opt := range opts {
		opt(e)
	}
	e.initStdlib()
	e.initMetatables()
	return e
}

func (e *Engine) EvalFile(path string) ([]value.Value, error) {
	file, err := e.fs.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()
	return e.Eval(file)
}

func (e *Engine) Eval(source io.Reader) ([]value.Value, error) {
	p, err := parser.New(source)
	if err != nil {
		return nil, fmt.Errorf("create parser: %w", err)
	}
	ast, ok := p.Parse()
	if !ok {
		var errString bytes.Buffer
		errString.WriteString("errors occurred while parsing")
		if namer, ok := source.(Namer); ok {
			errString.WriteString(" " + namer.Name())
		}
		for _, err := range p.Errors() {
			errString.WriteString("\n\t" + err.Error())
		}
		return nil, fmt.Errorf(errString.String())
	}

	results, err := e.evaluateChunk(ast)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (e *Engine) currentScope() *value.Table {
	return e.scopes[0]
}

func (e Engine) dumpState() {
	fmt.Printf("clock: %T\n", e.clock)
	fmt.Println("global scope:")
	for name, value := range e._G.Fields {
		fmt.Printf("%-15s = %s\n", name, value)
	}
	if e.currentScope() != e._G {
		fmt.Println("current scope:")
		for name, value := range e.currentScope().Fields {
			fmt.Printf("%-15s = %s\n", name, value)
		}
	}
}

func (e *Engine) assign(scope *value.Table, name string, val value.Value) {
	scope.Set(value.NewString(name), val)
}

func (e *Engine) enterNewScope() {
	e.scopes = append([]*value.Table{value.NewTable()}, e.scopes...)
}

func (e *Engine) leaveScope() {
	e.scopes[0] = nil
	e.scopes = e.scopes[1:]
}

// variable searches for a variable with the given Name, starting in the current
// scope and always visiting the parent scope if there is no such variable.
func (e *Engine) variable(name string) (value.Value, bool) {
	varName := value.NewString(name)
	for i := 0; i < len(e.scopes); i++ {
		if val, ok := e.scopes[i].Fields[varName]; ok {
			return val, true
		}
	}
	return nil, false
}

func (e *Engine) call(fn *value.Function, args ...value.Value) (vs []value.Value, err error) {
	e.enterNewScope()
	defer e.leaveScope()

	if ok := e.stack.Push(StackFrame{
		Name: fn.Name,
	}); !ok {
		return e.error(value.NewString(fmt.Sprintf("Stack overflow while calling '%s'", fn.Name)))
	}
	defer e.stack.Pop()

	res, err := func() (vs []value.Value, err error) {
		defer func(vs *[]value.Value) {
			if r := recover(); r != nil {
				if ret, ok := r.(Return); ok {
					*vs = ret.Values
				} else {
					panic(r)
				}
			}
		}(&vs)

		results, err := fn.Callable(args...)
		if err != nil {
			var luaErr Error
			if errors.As(err, &luaErr) {
				return nil, luaErr
			}
			return nil, fmt.Errorf("error while calling '%s': %w", fn.Name, err)
		}
		return results, nil
	}()
	return res, err
}

func (e *Engine) createCallable(parameters ast.ParList, block ast.Block) (value.LuaFn, error) {
	return func(args ...value.Value) ([]value.Value, error) {
		// this assumes, that we are already in a separate function scope

		// assign all arguments to the parameters in the current scope
		for i, arg := range args {
			e.assign(e.currentScope(), parameters.NameList[i].Value(), arg)
		}

		results, err := e.evaluateBlock(block)
		if err != nil {
			return nil, fmt.Errorf("block: %w", err)
		}
		return results, nil
	}, nil
}
