package engine

import (
	"fmt"

	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/engine/value"
	"github.com/tsatke/lua/internal/token"
)

func (e *Engine) evaluate(block ast.Block) error {
	for _, stmt := range block {
		if err := e.evaluateStatement(stmt); err != nil {
			return fmt.Errorf("statement %T: %w", stmt, err)
		}
	}
	return nil
}

func (e *Engine) evaluateStatement(stmt ast.Statement) error {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error_); ok {
				_, _ = e.stderr.Write([]byte(fmt.Sprintf("error: %s", err.message)))
			}
		}
	}()
	switch s := stmt.(type) {
	case ast.Assignment:
		return e.evaluateAssignment(s)
	case ast.FunctionCall:
		_, err := e.evaluateFunctionCall(s)
		return err
	}
	return fmt.Errorf("%T unsupported", stmt)
}

func (e *Engine) evaluateFunctionCall(call ast.FunctionCall) (value.Value, error) {
	if call.Name != nil {
		return nil, fmt.Errorf("':'-calls unsupported")
	}

	if call.PrefixExp.(ast.PrefixExp).Var.Name == nil {
		return nil, fmt.Errorf("only simple function calls are supported")
	}

	// obtain function
	fnName := call.PrefixExp.(ast.PrefixExp).Var.Name.Value()
	val, ok := e.variable(fnName)
	if !ok {
		return nil, fmt.Errorf("attempt to call a nil value ('%s')", fnName)
	}
	if val.Type() != value.TypeFunction {
		return nil, fmt.Errorf("attempt to call value of type %s ('%s')", val.Type(), fnName)
	}
	fn := val.(*value.Function)

	// evaluate arguments
	args, err := e.evaluateArgs(call.Args)
	if err != nil {
		return nil, fmt.Errorf("args: %w", err)
	}

	result, err := fn.Callable(args...)
	if err != nil {
		return nil, fmt.Errorf("error while calling %s: %w", fn.Name, err)
	}
	return result, nil
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
		val, err := e.evaluateExpression(exp)
		if err != nil {
			return nil, fmt.Errorf("evaluate expression: %w", err)
		}
		vals = append(vals, val)
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
	val, err := e.evaluateExpression(exp)
	if err != nil {
		return fmt.Errorf("expression: %w", err)
	}
	e.assign(e.currentScope, name, val)
	return nil
}

func (e *Engine) evaluateExpression(exp ast.Exp) (value.Value, error) {
	switch ex := exp.(type) {
	case ast.SimpleExp:
		return e.evaluateSimpleExpression(ex)
	case ast.ComplexExp:
		return e.evaluateComplexExpression(ex)
	default:
		return nil, fmt.Errorf("%T unsupported", exp)
	}
}

func (e *Engine) evaluateComplexExpression(exp ast.ComplexExp) (value.Value, error) {
	switch ex := exp.(type) {
	case ast.PrefixExp:
		return e.evaluatePrefixExpression(ex)
	default:
		return nil, fmt.Errorf("%T unsupported", exp)
	}
}

func (e *Engine) evaluatePrefixExpression(exp ast.PrefixExp) (value.Value, error) {
	if exp.FunctionCall != nil {
		res, err := e.evaluateFunctionCall(exp.FunctionCall.(ast.FunctionCall))
		if err != nil {
			return nil, fmt.Errorf("function call: %w", err)
		}
		return res, nil
	}

	if exp.Exp != nil {
		res, err := e.evaluateExpression(exp.Exp)
		if err != nil {
			return nil, fmt.Errorf("expression: %w", err)
		}
		return res, nil
	}

	// only var is left
	res, err := e.evaluateVar(exp.Var)
	if err != nil {
		return nil, fmt.Errorf("var: %w", err)
	}
	return res, nil
}

func (e *Engine) evaluateVar(v ast.Var) (value.Value, error) {
	if v.Name == nil {
		return nil, fmt.Errorf("only names are supported as variables")
	}

	value, ok := e.variable(v.Name.Value())
	if !ok {
		return nil, nil
	}
	return value, nil
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
