package engine

import (
	"fmt"
	"github.com/tsatke/lua/internal/engine/value"
)

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

func (e *Engine) metaMethodFunction(object value.Value, event string) (*value.Function, error) {
	val, err := e.metaMethod(object, event)
	if err != nil {
		return nil, err
	}
	if e.isNil(val) {
		return nil, nil
	}
	return val.(*value.Function), nil
}

func (e *Engine) metaMethodTable(object value.Value, event string) (*value.Table, error) {
	val, err := e.metaMethod(object, event)
	if err != nil {
		return nil, err
	}
	if e.isNil(val) {
		return nil, nil
	}
	return val.(*value.Table), nil
}

func (e *Engine) metaMethod(object value.Value, event string) (value.Value, error) {
	var metaTable *value.Table

	switch object.Type() {
	case value.TypeTable:
		results, err := e.getmetatable(object)
		if err != nil {
			return nil, fmt.Errorf("getmetatable: %w", err)
		}
		if results[0] == nil || results[0] == value.Nil {
			return nil, nil
		}
		metaTable = results[0].(*value.Table)
	default:
		metaTable = e.metaTables.Table(object.Type())
	}

	if metaTable == nil {
		return nil, nil
	}

	eventString := value.NewString(event)
	metaMethods, err := e.rawget(metaTable, eventString)
	if err != nil {
		return nil, fmt.Errorf("rawget: %w", err)
	}

	return metaMethods[0], nil
}
