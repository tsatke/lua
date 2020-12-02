package value

type Number float64

func (Number) Type() Type       { return TypeNumber }
func (n Number) Value() float64 { return float64(n) }

func NewNumber(value float64) Number {
	return Number(value)
}
