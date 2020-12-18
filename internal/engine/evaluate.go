package engine

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/tsatke/lua/internal/ast"
	"github.com/tsatke/lua/internal/engine/value"
	"github.com/tsatke/lua/internal/token"
)

func (e *Engine) evaluateChunk(chunk ast.Chunk) (vs []value.Value, err error) {
	luaFn, err := e.createCallable(ast.ParList{}, chunk.Block)
	if err != nil {
		return nil, fmt.Errorf("create callable: %w", err)
	}

	fn := value.NewFunction("<anonymous>", luaFn)
	results, err := e.call(fn)

	var luaErr Error
	if errors.As(err, &luaErr) {
		return nil, luaErr
	}
	return results, nil
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
	case ast.LastStatement:
		return e.evaluateLastStatement(s)
	}
	return nil, fmt.Errorf("%T unsupported", stmt)
}

func (e *Engine) evaluateLastStatement(stmt ast.LastStatement) ([]value.Value, error) {
	if stmt.Break {
		return nil, fmt.Errorf("break not supported yet")
	}

	results, err := e.evaluateExpList(stmt.ExpList)
	if err != nil {
		return nil, fmt.Errorf("explist: %w", err)
	}
	panic(Return{
		Values: results,
	})
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
	if e.valueIsLogicallyTrue(ifCond) {
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

func (e *Engine) valueIsLogicallyTrue(val value.Value) bool {
	return !(val == nil || val == value.False || val == value.Nil)
}

func (e *Engine) evaluateDoBlock(block ast.DoBlock) ([]value.Value, error) {
	return e.evaluateBlock(block.Do)
}

func (e *Engine) evaluateFunction(decl ast.Function) ([]value.Value, error) {
	fnName := "<anonymous>"
	isAnonymous := decl.FuncName == nil
	if !isAnonymous {
		if decl.FuncName.Name2 != nil {
			return nil, fmt.Errorf("function with ':' not supported")
		}
		if len(decl.FuncName.Name1) != 1 {
			return nil, fmt.Errorf("only plain functions supported")
		}

		fnName = decl.FuncName.Name1[0].Value()
	}

	luaFn, err := e.createCallable(decl.FuncBody.ParList, decl.FuncBody.Block)
	if err != nil {
		return nil, fmt.Errorf("create callable: %w", err)
	}

	functionValue := value.NewFunction(fnName, luaFn)
	if isAnonymous {
		return values(functionValue), nil
	}
	e.assign(e._G, fnName, functionValue) // change the function Name when we support more than just one Name fragment
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

func (e *Engine) evaluateFunctionCall(call ast.FunctionCall) (vs []value.Value, err error) {
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
		return []value.Value{value.NewString(args.String.Value())}, nil
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
			return nil, fmt.Errorf("expression: %w", err)
		}
		vals = append(vals, results...)
	}
	return vals, nil
}

func (e *Engine) evaluateAssignment(assignment ast.Assignment) error {
	if len(assignment.ExpList) == 1 {
		values, err := e.evaluateExpression(assignment.ExpList[0])
		if err != nil {
			return fmt.Errorf("explist: %w", err)
		}

		for i := 0; i < len(values); i++ {
			if err := e.evaluateAssign(assignment.VarList[i], values[i]); err != nil {
				return fmt.Errorf("assign: %w", err)
			}
		}
	}

	varAmount, expAmount := len(assignment.VarList), len(assignment.ExpList)
	amount := varAmount
	if expAmount < varAmount {
		amount = expAmount
	}

	expressions, err := e.evaluateExpList(assignment.ExpList)
	if err != nil {
		return fmt.Errorf("explist: %w", err)
	}

	for i := 0; i < amount; i++ {
		if err := e.evaluateAssign(assignment.VarList[i], expressions[i]); err != nil {
			return fmt.Errorf("assign: %w", err)
		}
	}

	return nil
}

func (e *Engine) evaluateAssign(v ast.Var, val value.Value) error {
	if v.Name == nil {
		return fmt.Errorf("can only assign to simple variable")
	}

	name := v.Name.Value()
	scope := e._G
	// if the variable is already declared in the current scope (either
	// we are currently in the global scope, or the variable has been declared
	// with 'local'), assign in the current scope
	if _, ok := e.currentScope().Fields[name]; ok {
		scope = e.currentScope()
	}
	e.assign(scope, name, val)
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
	case ast.UnopExp:
		return e.evaluateUnopExpression(ex)
	case ast.BinopExp:
		return e.evaluateBinopExpression(ex)
	case ast.Function:
		return e.evaluateFunction(ex)
	default:
		return nil, fmt.Errorf("%T unsupported", exp)
	}
}

func (e *Engine) evaluateUnopExpression(exp ast.UnopExp) ([]value.Value, error) {
	operands, err := e.evaluateExpression(exp.Exp)
	if err != nil {
		return nil, fmt.Errorf("operand: %w", err)
	}
	operand := operands[0]

	switch exp.Unop.Value() {
	case "-":
		if num, ok := operand.(value.Number); ok {
			results, err := e.multiply(value.NewNumber(-1), num)
			if err != nil {
				return nil, fmt.Errorf("arithmetic unary negation: %w", err)
			}
			return results, nil
		}
	case "not":
		if boolVal, ok := operand.(value.Boolean); ok {
			if boolVal {
				return values(value.False), nil
			}
			return values(value.True), nil
		}
	case "~":
		return e.evaluateBitwiseNot(operand)
	}
	return nil, fmt.Errorf("unsupported unary operator '%s' on %s", exp.Unop.Value(), operand.Type())
}

func (e *Engine) evaluateBinopExpression(exp ast.BinopExp) ([]value.Value, error) {
	switch exp.Binop.Value() {
	case "or":
		return e.evaluateBinopLazy(exp)
	}
	return e.evaluateBinopEager(exp)
}

func (e *Engine) evaluateBinopLazy(exp ast.BinopExp) ([]value.Value, error) {
	switch exp.Binop.Value() {
	case "or":
		lefts, err := e.evaluateExpression(exp.Left)
		if err != nil {
			return nil, fmt.Errorf("left exp: %w", err)
		}
		if e.valueIsLogicallyTrue(lefts[0]) {
			return values(lefts[0]), nil
		}

		rights, err := e.evaluateExpression(exp.Right)
		if err != nil {
			return nil, fmt.Errorf("right exp: %w", err)
		}
		return values(rights[0]), nil
	}
	return nil, fmt.Errorf("unsupported binary operator %s", exp.Binop.Value())
}

func (e *Engine) evaluateBinopEager(exp ast.BinopExp) ([]value.Value, error) {
	lefts, err := e.evaluateExpression(exp.Left)
	if err != nil {
		return nil, fmt.Errorf("left exp: %w", err)
	}
	rights, err := e.evaluateExpression(exp.Right)
	if err != nil {
		return nil, fmt.Errorf("right exp: %w", err)
	}
	left, right := lefts[0], rights[0]

	var operator func(left, right value.Value) ([]value.Value, error)

	switch exp.Binop.Value() {
	case "+":
		operator = e.evaluateAddition
	case "-":
		operator = e.evaluateSubtraction
	case "*":
		operator = e.evaluateMultiplication
	case "/":
		operator = e.evaluateDivision
	case "//":
		operator = e.evaluateFloorDivision
	case "==":
		operator = e.evaluateEqual
	case "~=":
		operator = e.evaluateUnequal
	case "<":
		operator = e.evaluateLess
	case "<=":
		operator = e.evaluateLessOrEqual
	case ">":
		operator = e.evaluateGreater
	case ">=":
		operator = e.evaluateGreaterOrEqual
	case "and":
		operator = e.evaluateAnd
	case "|":
		operator = e.evaluateBitwiseOr
	case "&":
		operator = e.evaluateBitwiseAnd
	case "<<":
		operator = e.evaluateBitwiseLeftShift
	case ">>":
		operator = e.evaluateBitwiseRightShift
	case "..":
		operator = e.evaluateConcatenation
	case "%":
		operator = e.evaluateModulo
	}
	if operator == nil {
		return nil, fmt.Errorf("unsupported binary operator %s", exp.Binop.Value())
	}
	return operator(left, right)
}

func (e *Engine) evaluateBitwiseNot(val value.Value) ([]value.Value, error) {
	results, err := e.bitwiseNot(val)
	if err != nil {
		return nil, fmt.Errorf("bitwise not: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateConcatenation(left, right value.Value) ([]value.Value, error) {
	results, err := e.concatenation(left, right)
	if err != nil {
		return nil, fmt.Errorf("concat: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateModulo(left, right value.Value) ([]value.Value, error) {
	results, err := e.modulo(left, right)
	if err != nil {
		return nil, fmt.Errorf("modulo: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateBitwiseOr(left, right value.Value) ([]value.Value, error) {
	results, err := e.bitwiseOr(left, right)
	if err != nil {
		return nil, fmt.Errorf("bitwise or: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateBitwiseAnd(left, right value.Value) ([]value.Value, error) {
	results, err := e.bitwiseAnd(left, right)
	if err != nil {
		return nil, fmt.Errorf("bitwise and: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateBitwiseLeftShift(left, right value.Value) ([]value.Value, error) {
	results, err := e.bitwiseLeftShift(left, right)
	if err != nil {
		return nil, fmt.Errorf("left shift: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateBitwiseRightShift(left, right value.Value) ([]value.Value, error) {
	results, err := e.bitwiseRightShift(left, right)
	if err != nil {
		return nil, fmt.Errorf("right shift: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateAnd(left, right value.Value) ([]value.Value, error) {
	results, err := e.and(left, right)
	if err != nil {
		return nil, fmt.Errorf("and: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateAddition(left, right value.Value) ([]value.Value, error) {
	results, err := e.add(left, right)
	if err != nil {
		return nil, fmt.Errorf("add: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateSubtraction(left, right value.Value) ([]value.Value, error) {
	results, err := e.subtract(left, right)
	if err != nil {
		return nil, fmt.Errorf("subtract: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateMultiplication(left, right value.Value) ([]value.Value, error) {
	results, err := e.multiply(left, right)
	if err != nil {
		return nil, fmt.Errorf("multiply: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateDivision(left, right value.Value) ([]value.Value, error) {
	results, err := e.divide(left, right)
	if err != nil {
		return nil, fmt.Errorf("divide: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateFloorDivision(left, right value.Value) ([]value.Value, error) {
	results, err := e.floorDivide(left, right)
	if err != nil {
		return nil, fmt.Errorf("floor divide: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateEqual(left, right value.Value) ([]value.Value, error) {
	eq, err := e.equal(left, right)
	if err != nil {
		return nil, fmt.Errorf("compare: %w", err)
	}
	if eq {
		return values(value.True), nil
	}
	return values(value.False), nil
}

func (e *Engine) evaluateUnequal(left, right value.Value) ([]value.Value, error) {
	eq, err := e.equal(left, right)
	if err != nil {
		return nil, fmt.Errorf("compare: %w", err)
	}
	if !eq {
		return values(value.True), nil
	}
	return values(value.False), nil
}

func (e *Engine) evaluateLess(left, right value.Value) ([]value.Value, error) {
	less, err := e.less(left, right)
	if err != nil {
		return nil, fmt.Errorf("compare: %w", err)
	}
	if less {
		return values(value.True), nil
	}
	return values(value.False), nil
}

func (e *Engine) evaluateLessOrEqual(left, right value.Value) ([]value.Value, error) {
	lessEq, err := e.lessEqual(left, right)
	if err != nil {
		return nil, fmt.Errorf("compare: %w", err)
	}
	if lessEq {
		return values(value.True), nil
	}
	return values(value.False), nil
}

func (e *Engine) evaluateGreater(left, right value.Value) ([]value.Value, error) {
	greater, err := e.less(right, left)
	if err != nil {
		return nil, fmt.Errorf("compare: %w", err)
	}
	if greater {
		return values(value.True), nil
	}
	return values(value.False), nil
}

func (e *Engine) evaluateGreaterOrEqual(left, right value.Value) ([]value.Value, error) {
	greaterEq, err := e.lessEqual(right, left)
	if err != nil {
		return nil, fmt.Errorf("compare: %w", err)
	}
	if greaterEq {
		return values(value.True), nil
	}
	return values(value.False), nil
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

func (e *Engine) evaluateSimpleExpression(exp ast.SimpleExp) (value.Value, error) {
	switch {
	case exp.String != nil:
		return value.NewString(exp.String.Value()), nil
	case exp.Number != nil:
		val, err := strconv.ParseFloat(exp.Number.Value(), 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse value '%s' as number", exp.Number.Value())
		}
		return value.NewNumber(val), nil
	case exp.True != nil:
		return value.True, nil
	case exp.False != nil:
		return value.False, nil
	case exp.Nil != nil:
		return value.Nil, nil
	}
	return nil, fmt.Errorf("%T not supported as simple exp", exp)
}
