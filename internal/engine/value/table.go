package value

type Table struct {
	Metatable *Table

	Fields map[Value]Value
}

func NewTable() *Table {
	return &Table{
		Fields: make(map[Value]Value),
	}
}

func (Table) Type() Type { return TypeTable }

func (t *Table) Set(key Value, value Value) {
	if value == Nil || value == nil {
		delete(t.Fields, key)
	} else {
		t.Fields[key] = value
	}
}

func (t *Table) Get(key Value) (Value, bool) {
	val, ok := t.Fields[key]
	return val, ok
}

func (t *Table) Length() Value {
	return NewNumber(float64(len(t.Fields)))
}
