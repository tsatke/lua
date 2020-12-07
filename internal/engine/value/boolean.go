package value

const (
	False = Boolean(false)
	True  = Boolean(true)
)

type Boolean bool

func (Boolean) Type() Type { return TypeBoolean }
func (b Boolean) String() string {
	if !b {
		return "false"
	}
	return "true"
}
