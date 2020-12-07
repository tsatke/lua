package value

type Table struct {
	Metatable *Table

	fields map[string]Value
}

func NewTable() *Table {
	return &Table{
		fields: make(map[string]Value),
	}
}

func (Table) Type() Type { return TypeTable }

func (t *Table) Set(key string, value Value) {
	if value == Nil {
		delete(t.fields, key)
	} else {
		t.fields[key] = value
	}
}

func (t *Table) Get(key string) (Value, bool) {
	val, ok := t.fields[key]
	return val, ok
}
