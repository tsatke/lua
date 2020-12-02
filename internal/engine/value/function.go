package value

type LuaFn func(...Value) ([]Value, error)

type Function struct {
	Name     string
	Callable LuaFn
}

func NewFunction(name string, callable LuaFn) *Function {
	return &Function{
		Name:     name,
		Callable: callable,
	}
}

func (Function) Type() Type { return TypeFunction }

func (f Function) String() string {
	return "function " + f.Name
}
