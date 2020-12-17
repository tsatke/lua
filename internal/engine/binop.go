package engine

import (
	"fmt"
	. "github.com/tsatke/lua/internal/engine/value"
	"math"
	"strconv"
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

func (e *Engine) and(left, right Value) ([]Value, error) {
	if !e.valueIsLogicallyTrue(left) {
		return values(left), nil
	}
	return values(right), nil
}

func (e *Engine) bitwiseOr(left, right Value) ([]Value, error) {
	return binaryIntegralOperation(left, right, func(left, right int64) int64 {
		return left | right
	})
}

func (e *Engine) bitwiseAnd(left, right Value) ([]Value, error) {
	return binaryIntegralOperation(left, right, func(left, right int64) int64 {
		return left & right
	})
}

func (e *Engine) bitwiseLeftShift(left, right Value) ([]Value, error) {
	return binaryIntegralOperation(left, right, func(left, right int64) int64 {
		return left << right
	})
}

func (e *Engine) bitwiseRightShift(left, right Value) ([]Value, error) {
	return binaryIntegralOperation(left, right, func(left, right int64) int64 {
		return left >> right
	})
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

func (e *Engine) concatenation(left, right Value) ([]Value, error) {
	if left.Type() != TypeString && left.Type() != TypeNumber {
		return nil, fmt.Errorf("left is not a string")
	}
	if right.Type() != TypeString && right.Type() != TypeNumber {
		return nil, fmt.Errorf("right is not a string")
	}

	var leftVal string
	if left.Type() == TypeString {
		leftVal = left.(String).String()
	} else {
		leftVal = strconv.FormatFloat(left.(Number).Value(), 'G', -1, 64)
	}
	var rightVal string
	if right.Type() == TypeString {
		rightVal = right.(String).String()
	} else {
		rightVal = strconv.FormatFloat(right.(Number).Value(), 'G', -1, 64)
	}

	return values(NewString(leftVal + rightVal)), nil
}

func (e *Engine) modulo(left, right Value) ([]Value, error) {
	if left.Type() != TypeNumber {
		return nil, fmt.Errorf("left is not a number")
	}
	if right.Type() != TypeNumber {
		return nil, fmt.Errorf("right is not a number")
	}

	leftNum := left.(Number).Value()
	rightNum := right.(Number).Value()

	return values(NewNumber(math.Mod(leftNum, rightNum))), nil
}

func binaryIntegralOperation(left, right Value, operator func(left, right int64) int64) ([]Value, error) {
	if left.Type() != TypeNumber && left.Type() != TypeString {
		return nil, fmt.Errorf("left is not a number")
	}
	if right.Type() != TypeNumber && right.Type() != TypeString {
		return nil, fmt.Errorf("right is not a number")
	}

	var leftVal float64
	if left.Type() == TypeNumber {
		leftVal = left.(Number).Value()
	} else {
		res, err := strconv.ParseFloat(left.(String).String(), 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert '%s' to a number", left.(String).String())
		}
		leftVal = res
	}
	var rightVal float64
	if right.Type() == TypeNumber {
		rightVal = right.(Number).Value()
	} else {
		res, err := strconv.ParseFloat(right.(String).String(), 64)
		if err != nil {
			return nil, fmt.Errorf("cannot convert '%s' to a number", right.(String).String())
		}
		rightVal = res
	}

	computationResult := operator(int64(leftVal), int64(rightVal))
	resultValue := NewNumber(float64(computationResult))
	return values(resultValue), nil
}
