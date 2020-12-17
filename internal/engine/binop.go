package engine

import (
	"fmt"
	. "github.com/tsatke/lua/internal/engine/value"
	"math"
)

func (e *Engine) add(left, right Value) ([]Value, error) {
	if left.Type() != TypeNumber {
		return nil, fmt.Errorf("left is not a number")
	}
	if right.Type() != TypeNumber {
		return nil, fmt.Errorf("right is not a number")
	}

	leftNum := left.(Number).Value()
	rightNum := right.(Number).Value()
	return values(NewNumber(leftNum + rightNum)), nil
}

func (e *Engine) subtract(left, right Value) ([]Value, error) {
	if left.Type() != TypeNumber {
		return nil, fmt.Errorf("left is not a number")
	}
	if right.Type() != TypeNumber {
		return nil, fmt.Errorf("right is not a number")
	}

	leftNum := left.(Number).Value()
	rightNum := right.(Number).Value()
	return values(NewNumber(leftNum - rightNum)), nil
}

func (e *Engine) multiply(left, right Value) ([]Value, error) {
	if left.Type() != TypeNumber {
		return nil, fmt.Errorf("left is not a number")
	}
	if right.Type() != TypeNumber {
		return nil, fmt.Errorf("right is not a number")
	}

	leftNum := left.(Number).Value()
	rightNum := right.(Number).Value()
	return values(NewNumber(leftNum * rightNum)), nil
}

func (e *Engine) divide(left, right Value) ([]Value, error) {
	if left.Type() != TypeNumber {
		return nil, fmt.Errorf("left is not a number")
	}
	if right.Type() != TypeNumber {
		return nil, fmt.Errorf("right is not a number")
	}

	leftNum := left.(Number).Value()
	rightNum := right.(Number).Value()
	return values(NewNumber(leftNum / rightNum)), nil
}

func (e *Engine) floorDivide(left, right Value) ([]Value, error) {
	if left.Type() != TypeNumber {
		return nil, fmt.Errorf("left is not a number")
	}
	if right.Type() != TypeNumber {
		return nil, fmt.Errorf("right is not a number")
	}

	leftNum := left.(Number).Value()
	rightNum := right.(Number).Value()
	return values(NewNumber(math.Floor(leftNum / rightNum))), nil
}

func (e *Engine) cmpEqual(left, right Value) ([]Value, error) {
	if left.Type() != right.Type() {
		return values(False), nil
	}
	switch left.Type() {
	case TypeNumber:
		leftNum := left.(Number).Value()
		rightNum := right.(Number).Value()
		return values(Boolean(leftNum == rightNum)), nil
	case TypeString:
		leftNum := left.(String).String()
		rightNum := right.(String).String()
		return values(Boolean(leftNum == rightNum)), nil
	}
	return nil, fmt.Errorf("unsupported comparison type: %s", left.Type())
}
