package value

const (
	False = Boolean(0)
	True  = Boolean(1)
)

type Boolean uint8

func (Boolean) Type() Type { return TypeBoolean }
func (b Boolean) String() string {
	if b == 0 {
		return "false"
	}
	return "true"
}
