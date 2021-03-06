package lua

import (
	"io"
	"os"
	"strings"
)

func ExampleEngine_EvalString() {
	source := `print("Hello, World!")`
	e := NewEngine(
		WithStdout(os.Stdout),
	)
	_, err := e.EvalString(source)
	if err != nil {
		panic(err)
	}
	// Output: Hello, World!
}

func ExampleEngine_Eval() {
	// this can be every file from os.Open or (afero.Fs).Open, which all implement the
	// io.Reader interface
	var sourceFile io.Reader = strings.NewReader(`print("Hello, World!")`)
	e := NewEngine(
		WithStdout(os.Stdout),
	)
	_, err := e.Eval(sourceFile)
	if err != nil {
		panic(err)
	}
	// Output: Hello, World!
}
