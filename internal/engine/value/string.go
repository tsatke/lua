package value

type String string

func (String) Type() Type { return TypeString }

func NewString(value string) String {
	return String(value)
}
