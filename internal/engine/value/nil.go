package value

const (
	// Nil is the constant value nil.
	Nil = nilValue(0)
)

var _ Value = (*nilValue)(nil)

type nilValue uint8

func (nilValue) Type() Type     { return TypeNil }
func (nilValue) String() string { return "nil" }
