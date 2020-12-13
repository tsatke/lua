package engine

import (
	"fmt"

	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/engine/value"
	"github.com/tsatke/lua/internal/token"
)

func (e *Engine) protect(errToSet *error) {
	if r := recover(); r != nil {
		if luaErr, ok := r.(Error); ok {
			*errToSet = luaErr
		} else {
			panic(r)
		}
	}
}

func (e *Engine) evaluateChunk(chunk ast.Chunk) (vs []value.Value, err error) {
	defer e.protect(&err) // chunks are run just as blocks, but protected

	if ok := e.stack.Push(StackFrame{
		Name: chunk.Name,
	}); !ok {
		_, _ = e.error(value.NewString("Stack overflow while evaluating chunk"))
		return nil, fmt.Errorf("Stack overflow")
	}
	defer e.stack.Pop()

	return e.evaluateBlock(chunk.Block)
}

func (e *Engine) evaluateBlock(block ast.Block) ([]value.Value, error) {
	e.enterNewScope()
	defer e.leaveScope()

	for _, stmt := range block.StatementsWithoutLast() {
		if _, err := e.evaluateStatement(stmt); err != nil {
			return nil, fmt.Errorf("statement %T: %w", stmt, err)
		}
	}

	lastStatement, ok := block.LastStatement()
	if !ok {
		return nil, nil
	}
	results, err := e.evaluateStatement(lastStatement)
	if err != nil {
		return nil, fmt.Errorf("last statement %T: %w", lastStatement, err)
	}
	return results, nil
}

func (e *Engine) evaluateStatement(stmt ast.Statement) ([]value.Value, error) {
	switch s := stmt.(type) {
	case ast.Assignment:
		return nil, e.evaluateAssignment(s)
	case ast.Local:
		return nil, e.evaluateLocal(s)
	case ast.FunctionCall:
		return e.evaluateFunctionCall(s)
	case ast.Function:
		return e.evaluateFunction(s)
	case ast.IfBlock:
		return e.evaluateIfBlock(s)
	case ast.DoBlock:
		return e.evaluateDoBlock(s)
	}
	return nil, fmt.Errorf("%T unsupported", stmt)
}

func (e *Engine) evaluateIfBlock(block ast.IfBlock) ([]value.Value, error) {
	// if
	ifConds, err := e.evaluateExpression(block.If)
	if err != nil {
		return nil, fmt.Errorf("expression: %w", err)
	}
	var ifCond value.Value
	if len(ifConds) > 0 {
		ifCond = ifConds[0]
	}
	if !(ifCond == nil || ifCond == value.False || ifCond == value.Nil) {
		return e.evaluateBlock(block.Then)
	}

	// elseif (all of them)
	for i, elseIf := range block.ElseIf {
		conds, err := e.evaluateExpression(elseIf.If)
		if err != nil {
			return nil, fmt.Errorf("elseif[%d] expression: %w", i, err)
		}
		var cond value.Value
		if len(conds) > 0 {
			cond = conds[0]
		}
		if !(cond == nil || cond == value.False || cond == value.Nil) {
			return e.evaluateBlock(elseIf.Then)
		}
	}

	// else
	if len(block.ElseIf) > 0 {
		panic("elseif is not supported yet")
	}
	if block.Else != nil {
		return e.evaluateBlock(block.Else)
	}
	return nil, nil
}

func (e *Engine) evaluateDoBlock(block ast.DoBlock) ([]value.Value, error) {
	return e.evaluateBlock(block.Do)
}

func (e *Engine) evaluateFunction(decl ast.Function) ([]value.Value, error) {
	if decl.FuncName.Name2 != nil {
		return nil, fmt.Errorf("function with ':' not supported")
	}
	if len(decl.FuncName.Name1) != 1 {
		return nil, fmt.Errorf("only plain functions supported")
	}

	fnName := decl.FuncName.Name1[0].Value()

	luaFn, err := e.createCallable(decl.FuncBody.ParList, decl.FuncBody.Block)
	if err != nil {
		return nil, fmt.Errorf("create callable: %w", err)
	}
	e.assign(e._G, fnName, value.NewFunction(fnName, luaFn)) // change the function Name when we support more than just one Name fragment
	return nil, nil
}

func (e *Engine) evaluateLocal(local ast.Local) error {
	nameAmount, expAmount := len(local.NameList), len(local.ExpList)
	amount := nameAmount
	if expAmount < nameAmount {
		amount = expAmount
	}

	for i := 0; i < amount; i++ {
		if err := e.evaluateAssignLocal(local.NameList[i], local.ExpList[i]); err != nil {
			return fmt.Errorf("assign: %w", err)
		}
	}

	return nil
}

func (e *Engine) evaluateFunctionCall(call ast.FunctionCall) ([]value.Value, error) {
	prefixExp := call.PrefixExp
	if prefixExp.Exp != nil {
		return nil, fmt.Errorf("expression calls are not supported yet")
	}

	results, err := e.evaluatePrefixExpression(prefixExp)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (e *Engine) evaluateArgs(args ast.Args) ([]value.Value, error) {
	if args.TableConstructor != nil {
		return nil, fmt.Errorf("table constructors unsupported")
	}

	if args.String != nil {
		return []value.Value{value.NewString(stringTokenToString(args.String))}, nil
	}
	if args.ExpList != nil {
		vals, err := e.evaluateExpList(args.ExpList)
		if err != nil {
			return nil, fmt.Errorf("explist: %w", err)
		}
		return vals, nil
	}
	return nil, nil
}

func (e *Engine) evaluateExpList(explist []ast.Exp) ([]value.Value, error) {
	var vals []value.Value
	for _, exp := range explist {
		results, err := e.evaluateExpression(exp)
		if err != nil {
			return nil, fmt.Errorf("evaluateChunk expression: %w", err)
		}
		vals = append(vals, results...)
	}
	return vals, nil
}

func (e *Engine) evaluateAssignment(assignment ast.Assignment) error {
	varAmount, expAmount := len(assignment.VarList), len(assignment.ExpList)
	amount := varAmount
	if expAmount < varAmount {
		amount = expAmount
	}

	for i := 0; i < amount; i++ {
		if err := e.evaluateAssign(assignment.VarList[i], assignment.ExpList[i]); err != nil {
			return fmt.Errorf("assign: %w", err)
		}
	}

	return nil
}

func (e *Engine) evaluateAssign(v ast.Var, exp ast.Exp) error {
	if v.Name == nil {
		return fmt.Errorf("can only assign to simple variable")
	}

	name := v.Name.Value()
	vals, err := e.evaluateExpression(exp)
	if err != nil {
		return fmt.Errorf("expression: %w", err)
	}
	scope := e._G
	// if the variable is already declared in the current scope (either
	// we are currently in the global scope, or the variable has been declared
	// with 'local'), assign in the current scope
	if _, ok := e.currentScope().Fields[name]; ok {
		scope = e.currentScope()
	}
	e.assign(scope, name, vals[0])
	return nil
}

func (e *Engine) evaluateAssignLocal(tk token.Token, exp ast.Exp) error {
	if !tk.Is(token.Name) {
		return fmt.Errorf("expected a Name token, but got %s (parser broken?)", tk)
	}

	name := tk.Value()

	val, err := e.evaluateExpression(exp)
	if err != nil {
		return fmt.Errorf("expression: %w", err)
	}
	// assign in current scope, since it is a local assignment
	e.assign(e.currentScope(), name, val[0])
	return nil
}

func (e *Engine) evaluateExpression(exp ast.Exp) ([]value.Value, error) {
	switch ex := exp.(type) {
	case ast.SimpleExp:
		evaluated, err := e.evaluateSimpleExpression(ex)
		if err != nil {
			return nil, err
		}
		return values(evaluated), nil
	case ast.ComplexExp:
		return e.evaluateComplexExpression(ex)
	default:
		return nil, fmt.Errorf("%T unsupported", exp)
	}
}

func (e *Engine) evaluateComplexExpression(exp ast.ComplexExp) ([]value.Value, error) {
	switch ex := exp.(type) {
	case ast.PrefixExp:
		return e.evaluatePrefixExpression(ex)
	default:
		return nil, fmt.Errorf("%T unsupported", exp)
	}
}

func (e *Engine) evaluatePrefixExpression(exp ast.PrefixExp) ([]value.Value, error) {
	var current value.Value
	var currentName string

	if exp.Exp != nil {
		currentName = "(<exp>)"

		results, err := e.evaluateExpression(exp.Exp)
		if err != nil {
			return nil, fmt.Errorf("expression: %w", err)
		}
		if len(results) == 0 {
			current = nil
		} else {
			// if an expression is in parenthesis, which is the case here,
			// the values it evaluates to are cut down to one, meaning that all
			// elements except the first one are discarded
			current = results[0]
		}
	} else {
		name := exp.Name.Value()
		current, _ = e.variable(name) // we don't care whether or not the variable exists, and in case of fragments, the loop below will handle nil values
		currentName = name
	}

	if len(exp.Fragments) == 0 {
		return values(current), nil
	}

	var results []value.Value

	for i, fragment := range exp.Fragments {
		if current == nil {
			return nil, fmt.Errorf("cannot index nil value of variable '%s'", currentName)
		}

		results = nil

		if fragment.Exp != nil {
			return nil, fmt.Errorf("explicit indexing (a.[x]) not supported yet")
		}

		if fragment.Name != nil {
			table, ok := current.(*value.Table)
			if !ok {
				return nil, fmt.Errorf("cannot index variable of type %s", current.Type())
			}
			current, ok = table.Get(fragment.Name.Value())
			if !ok {
				return nil, fmt.Errorf("variable '%s' has no field '%s'", currentName, fragment.Name.Value())
			}
			currentName = fragment.Name.Value()
		}

		if fragment.Args != nil {
			// this fragment is a function call

			fn, ok := current.(*value.Function)
			if !ok {
				return nil, fmt.Errorf("cannot call non-function variable '%s'", currentName)
			}

			args, err := e.evaluateArgs(*fragment.Args)
			if err != nil {
				return nil, fmt.Errorf("args: %w", err)
			}

			res, err := e.call(fn, args...)
			if err != nil {
				return nil, fmt.Errorf("call '%s': %w", currentName, err)
			}
			results = res
			if len(res) == 0 && i < len(exp.Fragments)-1 {
				// One fragment function call can only return nil, if it's the last fragment. Otherwise,
				// subsequent calls would attempt to call something on nil.
				return nil, fmt.Errorf("calling '%s' returned nil, but it is not the last call in the chain", currentName)
			}
			if len(res) > 0 {
				current = res[0]
				currentName += "(...)"
			}
		}
	}
	return results, nil
}

func (e *Engine) evaluateVar(v ast.Var) ([]value.Value, error) {
	if v.Name == nil {
		return nil, fmt.Errorf("only names are supported as variables")
	}

	value, ok := e.variable(v.Name.Value())
	if !ok {
		return nil, nil
	}
	return values(value), nil
}

func (e *Engine) evaluateSimpleExpression(exp ast.SimpleExp) (value.Value, error) {
	switch {
	case exp.String != nil:
		return value.NewString(stringTokenToString(exp.String)), nil
	}
	return nil, fmt.Errorf("%T unsupported", exp)
}

func stringTokenToString(token token.Token) string {
	val := token.Value()
	return val[1 : len(val)-1]
}
