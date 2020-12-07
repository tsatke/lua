package engine

import "github.com/tsatke/lua/internal/engine/value"

type metaTables struct {
	NilMetaTable      *value.Table
	BooleanMetaTable  *value.Table
	NumberMetaTable   *value.Table
	StringMetaTable   *value.Table
	FunctionMetaTable *value.Table
	UserdataMetaTable *value.Table
	ThreadMetaTable   *value.Table
}

func (e *Engine) initMetatables() {
	e.metaTables.StringMetaTable = value.NewTable()
}

func (t metaTables) Table(typ value.Type) *value.Table {
	switch typ {
	case value.TypeNil:
		return t.NilMetaTable
	case value.TypeBoolean:
		return t.BooleanMetaTable
	case value.TypeNumber:
		return t.NumberMetaTable
	case value.TypeString:
		return t.StringMetaTable
	case value.TypeFunction:
		return t.FunctionMetaTable
	case value.TypeUserdata:
		return t.UserdataMetaTable
	case value.TypeThread:
		return t.ThreadMetaTable
	}
	return nil
}
