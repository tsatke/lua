package engine

import (
	"io"

	"github.com/spf13/afero"
)

type Option func(*Engine)

func WithFs(fs afero.Fs) Option {
	return func(e *Engine) {
		e.fs = fs
	}
}

func WithStdin(stdin io.Reader) Option {
	return func(e *Engine) {
		e.stdin = stdin
	}
}

func WithStdout(stdout io.Writer) Option {
	return func(e *Engine) {
		e.stdout = stdout
	}
}

func WithStderr(stderr io.Writer) Option {
	return func(e *Engine) {
		e.stderr = stderr
	}
}

func WithClock(clock Clock) Option {
	return func(e *Engine) {
		e.clock = clock
	}
}
