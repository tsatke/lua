package engine

import "io"

type Option func(*Engine)

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
