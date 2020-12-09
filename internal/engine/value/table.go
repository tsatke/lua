package value

type Table struct {
	Metatable *Table

	Fields map[string]Value
}

func NewTable() *Table {
	return &Table{
		Fields: make(map[string]Value),
	}
}

func (Table) Type() Type { return TypeTable }

func (t *Table) Set(key string, value Value) {
	if value == Nil {
		delete(t.Fields, key)
	} else {
		t.Fields[key] = value
	}
}

func (t *Table) Get(key string) (Value, bool) {
	val, ok := t.Fields[key]
	return val, ok
}
