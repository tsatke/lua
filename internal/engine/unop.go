package engine

import (
	"fmt"
	"math"

	. "github.com/tsatke/lua/internal/engine/value"
)

func (e *Engine) length(val Value) ([]Value, error) {
	if val.Type() == TypeString {
		return values(NewNumber(float64(len(val.(String).String())))), nil
	}

	// not a string, attempt metamethod
	metaMethod, err := e.metaMethodFunction(val, "__len")
	if err != nil {
		return nil, fmt.Errorf("meta method: %w", err)
	}
	if metaMethod != nil {
		return e.call(metaMethod, val)
	}

	// no metamethod, check if it's a table
	if val.Type() == TypeTable {
		return values(val.(*Table).Length()), nil
	}

	return nil, fmt.Errorf("operand is not a string")
}

func (e *Engine) bitwiseNot(val Value) ([]Value, error) {
	if val.Type() != TypeNumber {
		return nil, fmt.Errorf("operand is not a number")
	}

	floatVal := val.(Number).Value()
	if floatVal != math.Trunc(floatVal) {
		return nil, fmt.Errorf("%v has no integral representation", floatVal)
	}

	return values(NewNumber(float64(^int64(floatVal)))), nil
}
