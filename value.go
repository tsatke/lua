package lua

import "github.com/tsatke/lua/internal/engine/value"

type Value interface {
	_val()
}

type nilType uint8

func (nilType) _val() {}

type boolType uint8

func (boolType) _val() {}

type Number float64

func (Number) _val() {}

type String string

func (String) _val() {}

var (
	Nil   = nilType(0)
	False = boolType(0)
	True  = boolType(1)
)

type Values []Value

func (v Values) Count() int {
	return len(v)
}

func (v Values) Get(index int) Value {
	if index > len(v) {
		return Nil
	}
	return v[index]
}

func valuesFromInternal(vs ...value.Value) Values {
	var vals Values

	for _, v := range vs {
		var val Value
		switch v.Type() {
		case value.TypeNil:
			val = Nil
		case value.TypeBoolean:
			if !v.(value.Boolean) {
				val = False
			} else {
				val = True
			}
		case value.TypeNumber:
			val = Number(v.(value.Number))
		case value.TypeString:
			val = String(v.(value.String))
		case value.TypeFunction:
			fallthrough
		case value.TypeUserdata:
			fallthrough
		case value.TypeThread:
			fallthrough
		case value.TypeTable:
			panic("unsupported")
		}
		vals = append(vals, val)
	}

	return vals
}
