package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/spf13/afero"
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

	globalScope  *Scope
	currentScope *Scope

	gcrunning bool
	gcpercent int
}

type Scope struct {
	parent    *Scope
	variables map[string]value.Value
}

func newScope() *Scope {
	return newScopeWithParent(nil)
}

func newScopeWithParent(parent *Scope) *Scope {
	return &Scope{
		parent:    parent,
		variables: make(map[string]value.Value),
	}
}

// New creates a new, ready to use Engine, already applying all given options.
// By default, the engine uses os.Stdin as stdin, os.Stdout as stdout and os.Stderr
// as stderr.
func New(opts ...Option) *Engine {
	global := newScope()
	e := &Engine{
		fs: afero.NewOsFs(),

		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
		clock:  sysClock{},

		globalScope:  global,
		currentScope: global,
	}
	for _, opt := range opts {
		opt(e)
	}
	e.initStdlib()
	return e
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

func (e Engine) dumpState() {
	fmt.Printf("clock: %T\n", e.clock)
	fmt.Println("global scope:")
	for name, value := range e.globalScope.variables {
		fmt.Printf("%-15s = %s\n", name, value)
	}
	if e.currentScope != e.globalScope {
		fmt.Println("current scope:")
		for name, value := range e.currentScope.variables {
			fmt.Printf("%-15s = %s\n", name, value)
		}
	}
}

func (e *Engine) assign(scope *Scope, name string, val value.Value) {
	scope.variables[name] = val
}

func (e *Engine) enterNewScope() {
	e.currentScope = newScopeWithParent(e.currentScope)
}

func (e *Engine) leaveScope() {
	e.currentScope = e.currentScope.parent
}

func (e *Engine) variable(name string) (value.Value, bool) {
	for scope := e.currentScope; scope != nil; scope = scope.parent {
		if val, ok := scope.variables[name]; ok {
			return val, true
		}
	}
	return nil, false
}

// isVariableLocal determines whether a variable was declared as 'local'.
// If this method returns true for a variable name, changes of its value must
// take place in the current-scope. If this returns false, the variable either
// doesn't exist or exists, but is defined in the global scope.
func (e *Engine) isVariableLocal(name string) bool {
	if _, ok := e.currentScope.variables[name]; ok {
		return true
	}
	return false
}

func (e *Engine) call(fn *value.Function, args ...value.Value) ([]value.Value, error) {
	e.enterNewScope()
	defer e.leaveScope()

	results, err := fn.Callable(args...)
	if err != nil {
		return nil, fmt.Errorf("error while calling '%s': %w", fn.Name, err)
	}
	return results, nil
}

func toString(val value.Value) string {
	if val.Type() == value.TypeString {
		return string(val.(value.String))
	}
	panic("type " + val.Type().String())
}
