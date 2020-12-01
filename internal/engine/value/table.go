package value

type Table struct {
	fields map[string]Value
}

func NewTable() *Table {
	return &Table{
		fields: make(map[string]Value),
	}
}

func (t *Table) Set(key string, value Value) {
	if value == Nil {
		delete(t.fields, key)
	} else {
		t.fields[key] = value
	}
}
