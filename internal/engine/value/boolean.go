package value

const (
	False = booleanType(0)
	True  = booleanType(1)
)

type booleanType uint8

func (booleanType) Type() Type { return TypeBoolean }
func (b booleanType) String() string {
	if b == 0 {
		return "false"
	}
	return "true"
}
