package value

import "strconv"

type Number float64

func (Number) Type() Type       { return TypeNumber }
func (n Number) Value() float64 { return float64(n) }
func (n Number) String() string { return strconv.FormatFloat(float64(n), 'G', -1, 64) }

func NewNumber(value float64) Number {
	return Number(value)
}
