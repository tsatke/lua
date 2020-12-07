package parser

import "fmt"

type MismatchError struct {
	Expected interface{}
	Got      interface{}
}

func (e MismatchError) Error() string {
	return fmt.Sprintf("expected %v, but got %v", e.Expected, e.Got)
}

func ErrUnexpectedEof(expected interface{}) error {
	return ErrUnexpectedThing(expected, "EOF")
}

func ErrExpectedSomething(expected interface{}) error {
	return ErrUnexpectedThing(expected, "nothing")
}

func ErrUnexpectedThing(expected, got interface{}) error {
	return MismatchError{
		Expected: expected,
		Got:      got,
	}
}
