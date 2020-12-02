package lua

import "io"

type Option func(*Engine)

type ScannerType uint8

const (
	ScannerTypeInMemory ScannerType = iota
)

func WithScannerType(typ ScannerType) Option {
	return func(e *Engine) {
		e.scannerType = typ
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
