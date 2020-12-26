package value

import (
	"math"
	"strconv"
)

type Number float64

func (Number) Type() Type       { return TypeNumber }
func (n Number) Value() float64 { return float64(n) }
func (n Number) String() string { return strconv.FormatFloat(float64(n), 'G', -1, 64) }

func NewNumber(value float64) Number {
	return Number(value)
}

func (n Number) Integral() (int64, bool) {
	trunc := math.Trunc(float64(n))
	return int64(trunc), trunc == float64(n)
}
