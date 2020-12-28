package engine

import (
	"fmt"
	. "github.com/tsatke/lua/internal/engine/value"
)

func (e *Engine) cmpEqual(left, right Value) ([]Value, error) {
	if left.Type() == TypeTable && right.Type() == TypeTable {
		if left == right {
			// primitive equal check
			return values(True), nil
		}

		// use metamethod if available
		results, ok, err := e.binaryMetaMethodOperation("__eq", left, right)
		if !ok {
			if err != nil {
				return nil, err
			}
		} else {
			result := results[0]
			if e.valueIsLogicallyTrue(result) {
				return values(True), nil
			}
			return values(False), nil
		}
	}

	if left.Type() != right.Type() {
		return values(False), nil
	}
	switch left.Type() {
	case TypeNil:
		return values(Boolean(left == right)), nil
	case TypeNumber:
		leftNum := left.(Number).Value()
		rightNum := right.(Number).Value()
		return values(Boolean(leftNum == rightNum)), nil
	case TypeString:
		leftStr := left.(String).String()
		rightStr := right.(String).String()
		return values(Boolean(leftStr == rightStr)), nil
	case TypeBoolean:
		leftBool := left.(Boolean)
		rightBool := right.(Boolean)
		return values(Boolean(leftBool == rightBool)), nil
	}
	return nil, fmt.Errorf("cannot compare %s", left.Type())
}

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
