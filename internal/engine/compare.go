package engine

import (
	"fmt"
	. "github.com/tsatke/lua/internal/engine/value"
)

func (e *Engine) less(left, right Value) (bool, error) {
	if left.Type() != right.Type() {
		return false, nil
	}
	switch left.Type() {
	case TypeNumber:
		leftVal, rightVal := left.(Number).Value(), right.(Number).Value()
		return leftVal < rightVal, nil
	case TypeString:
		leftVal, rightVal := left.(String).String(), right.(String).String()
		return leftVal < rightVal, nil
	}
	return false, fmt.Errorf("%s is not comparable", left.Type())
}

func (e *Engine) lessEqual(left, right Value) (bool, error) {
	if left.Type() != right.Type() {
		return false, nil
	}
	switch left.Type() {
	case TypeNumber:
		leftVal, rightVal := left.(Number).Value(), right.(Number).Value()
		return leftVal <= rightVal, nil
	case TypeString:
		leftVal, rightVal := left.(String).String(), right.(String).String()
		return leftVal <= rightVal, nil
	}
	return false, fmt.Errorf("%s is not comparable", left.Type())
}

func (e *Engine) equal(left, right Value) (bool, error) {
	if left.Type() != right.Type() {
		return false, nil
	}
	switch left.Type() {
	case TypeNumber,
		TypeString,
		TypeBoolean:
		return left == right, nil
	}
	return false, fmt.Errorf("%s can not be checked for equality", left.Type())
}
