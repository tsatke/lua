package value

//go:generate stringer -type=Type

type Type uint8

const (
	TypeInvalid Type = iota
	TypeNil
	TypeBoolean
	TypeNumber
	TypeString
	TypeFunction
	TypeUserdata
	TypeThread
	TypeTable
)
