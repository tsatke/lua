package engine

import (
	"fmt"
	. "github.com/tsatke/lua/internal/engine/value"
	"math"
	"strconv"
)

func (e *Engine) add(left, right Value) ([]Value, error) {
	return e.binaryFloatingPointOperation("__add", left, right, func(left, right float64) float64 {
		return left + right
	})
}

func (e *Engine) subtract(left, right Value) ([]Value, error) {
	return e.binaryFloatingPointOperation("__sub", left, right, func(left, right float64) float64 {
		return left - right
	})
}

func (e *Engine) multiply(left, right Value) ([]Value, error) {
	return e.binaryFloatingPointOperation("__mul", left, right, func(left, right float64) float64 {
		return left * right
	})
}

func (e *Engine) divide(left, right Value) ([]Value, error) {
	return e.binaryFloatingPointOperation("__div", left, right, func(left, right float64) float64 {
		return left / right
	})
}

func (e *Engine) floorDivide(left, right Value) ([]Value, error) {
	return e.binaryFloatingPointOperation("__idiv", left, right, func(left, right float64) float64 {
		return math.Floor(left / right)
	})
}

func (e *Engine) power(left, right Value) ([]Value, error) {
	return e.binaryFloatingPointOperation("__pow", left, right, func(left, right float64) float64 {
		return math.Pow(left, right)
	})
}

func (e *Engine) and(left, right Value) ([]Value, error) {
	if !e.valueIsLogicallyTrue(left) {
		return values(left), nil
	}
	return values(right), nil
}

func (e *Engine) bitwiseOr(left, right Value) ([]Value, error) {
	return e.binaryIntegralOperation("__bor", left, right, func(left, right int64) int64 {
		return left | right
	})
}

func (e *Engine) bitwiseAnd(left, right Value) ([]Value, error) {
	return e.binaryIntegralOperation("__band", left, right, func(left, right int64) int64 {
		return left & right
	})
}

func (e *Engine) bitwiseXor(left, right Value) ([]Value, error) {
	return e.binaryIntegralOperation("__bxor", left, right, func(left, right int64) int64 {
		return left ^ right
	})
}

func (e *Engine) bitwiseLeftShift(left, right Value) ([]Value, error) {
	return e.binaryIntegralOperation("__shl", left, right, func(left, right int64) int64 {
		return left << right
	})
}

func (e *Engine) bitwiseRightShift(left, right Value) ([]Value, error) {
	return e.binaryIntegralOperation("__shr", left, right, func(left, right int64) int64 {
		return left >> right
	})
}

func (e *Engine) concatenation(left, right Value) ([]Value, error) {
	if e.isNil(left) || e.isNil(right) {
		return e.error(NewString("attempt to concatenate a nil value"))
	}

	if (left.Type() != TypeString && left.Type() != TypeNumber) &&
		(right.Type() != TypeString && right.Type() != TypeNumber) {
		results, ok, err := e.binaryMetaMethodOperation("__concat", left, right)
		if !ok {
			if err != nil {
				return nil, err
			}
			if left.Type() != TypeString && left.Type() != TypeNumber {
				return e.error(NewString("attempt to concatenate a " + left.Type().String() + " value"))
			}
			if right.Type() != TypeString && right.Type() != TypeNumber {
				return e.error(NewString("attempt to concatenate a " + right.Type().String() + " value"))
			}
		}
		return results, nil
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
	return e.binaryFloatingPointOperation("__mod", left, right, math.Mod)
}

func attemptConversionToNumber(left, right Value) (float64, float64, error) {
	if left.Type() != TypeNumber && left.Type() != TypeString {
		return 0, 0, fmt.Errorf("left is not a number")
	}
	if right.Type() != TypeNumber && right.Type() != TypeString {
		return 0, 0, fmt.Errorf("right is not a number")
	}

	var leftVal float64
	if left.Type() == TypeNumber {
		leftVal = left.(Number).Value()
	} else {
		res, err := strconv.ParseFloat(left.(String).String(), 64)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot convert '%s' to a number", left.(String).String())
		}
		leftVal = res
	}
	var rightVal float64
	if right.Type() == TypeNumber {
		rightVal = right.(Number).Value()
	} else {
		res, err := strconv.ParseFloat(right.(String).String(), 64)
		if err != nil {
			return 0, 0, fmt.Errorf("cannot convert '%s' to a number", right.(String).String())
		}
		rightVal = res
	}
	return leftVal, rightVal, nil
}

func (e *Engine) binaryFloatingPointOperation(event string, left, right Value, operator func(left, right float64) float64) ([]Value, error) {
	leftVal, rightVal, err := attemptConversionToNumber(left, right)
	if err != nil {
		results, ok, metaErr := e.binaryMetaMethodOperation(event, left, right)
		if !ok {
			if metaErr != nil {
				return nil, metaErr
			}
			return nil, err
		}
		return results, nil
	}

	computationResult := operator(leftVal, rightVal)
	resultValue := NewNumber(computationResult)
	return values(resultValue), nil
}

func (e *Engine) binaryIntegralOperation(event string, left, right Value, operator func(left, right int64) int64) ([]Value, error) {
	leftVal, rightVal, err := attemptConversionToNumber(left, right)
	if err != nil {
		results, ok, metaErr := e.binaryMetaMethodOperation(event, left, right)
		if !ok {
			if metaErr != nil {
				return nil, metaErr
			}
			return nil, err
		}
		return results, nil
	}

	computationResult := operator(int64(leftVal), int64(rightVal))
	resultValue := NewNumber(float64(computationResult))
	return values(resultValue), nil
}

// binaryMetaMethodOperation attempts to perform the meta method operation for the given event
// according to the language reference. The result of this method is the result values of
// the meta method, a bool which indicates, whether there was a metamethod or an error while obtaining
// it, and an error which would originate from the attempt of obtaining the metamethod.
//
// This means, that this function returns (nil,false,<anything>), if no metamethod could be obtained.
// If in this case, the error is not nil, that means, that there was an error obtaining the meta method.
// If the error is nil, while the bool flag is false, that means, that there was no metamethod that could
// have been called.
//
// This function returns (<not nil>, true, nil) in any case that is not covered by the above paragraph.
func (e *Engine) binaryMetaMethodOperation(event string, left, right Value) ([]Value, bool, error) {
	metaMethod, err := e.metaMethodFunction(left, event)
	if err != nil {
		return nil, false, fmt.Errorf("unable to obtain metamethod from left: %w", err)
	}

	if metaMethod == nil {
		metaMethod, err = e.metaMethodFunction(right, event)
		if err != nil {
			return nil, false, fmt.Errorf("unable to obtain metamethod from right: %w", err)
		}
	}

	if metaMethod == nil {
		return nil, false, err
	}

	results, err := e.call(metaMethod, left, right)
	if err != nil {
		return nil, false, err
	}
	return results, true, nil
}
