package lua

import (
	"io"
	"os"
	"strings"

	"github.com/tsatke/lua/internal/engine"
)

type Error struct{}

type Engine struct {
	engine *engine.Engine

	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer

	scannerType ScannerType
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

	e.engine = engine.New(
		engine.WithStdin(e.stdin),
		engine.WithStdout(e.stdout),
		engine.WithStderr(e.stderr),
	)

	return e
}

func (e Engine) EvalString(source string) (Values, error) {
	return e.Eval(strings.NewReader(source))
}

// Eval evaluates the bytes in the given reader. If an error occurs while parsing, or something strange
// happens internally, then an error will be returned. However, when Lua's error function is called, the
// error will be of type *Error.
//
// The parsed source will be evaluated as chunk, and all values that the chunk may return are returned
// as Values.
func (e Engine) Eval(source io.Reader) (Values, error) {
	results, err := e.engine.Eval(source)
	if err != nil {
		return nil, err
	}
	return convertFromInternal(results...), nil
}
