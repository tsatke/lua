package lua

import (
	"github.com/spf13/afero"
	"io"
	"os"
	"strings"

	"github.com/tsatke/lua/internal/engine"
)

type Engine struct {
	engine *engine.Engine

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	workingDir  string
	scannerType ScannerType
}

func EvalString(in string) error {
	_, err := NewEngine().EvalString(in)
	return err
}

func NewEngine(opts ...Option) Engine {
	e := Engine{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}

	for _, opt := range opts {
		opt(&e)
	}

	if e.workingDir == "" {
		sysWd, _ := os.Getwd()
		e.workingDir = sysWd
	}

	e.engine = engine.New(
		engine.WithStdin(e.stdin),
		engine.WithStdout(e.stdout),
		engine.WithStderr(e.stderr),
		engine.WithFs(afero.NewBasePathFs(afero.NewOsFs(), e.workingDir)),
	)

	return e
}

func (e Engine) EvalString(source string) (Values, error) {
	return e.Eval(strings.NewReader(source))
}

func (e Engine) EvalFile(path string) (Values, error) {
	results, err := e.engine.EvalFile(path)
	if err != nil {
		if luaErr, ok := err.(engine.Error); ok {
			return nil, errorFromInternal(luaErr)
		}
		return nil, err
	}
	return valuesFromInternal(results...), nil
}

// Eval evaluates the bytes in the given reader. If an error occurs while parsing, or something strange
// happens internally, then an error will be returned. However, when Lua's error function is called, the
// error will be of type Error.
//
// The parsed source will be evaluated as chunk, and all values that the chunk may return are returned
// as Values.
func (e Engine) Eval(source io.Reader) (Values, error) {
	results, err := e.engine.Eval(source)
	if err != nil {
		if luaErr, ok := err.(engine.Error); ok {
			return nil, errorFromInternal(luaErr)
		}
		return nil, err
	}
	return valuesFromInternal(results...), nil
}
