package value

type String string

func (String) Type() Type       { return TypeString }
func (s String) String() string { return string(s) }

func NewString(value string) String {
	return String(value)
}
