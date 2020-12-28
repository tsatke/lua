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

	fn := value.NewFunction(chunk.Name, luaFn)
	results, err := e.call(fn)

	var luaErr Error
	if errors.As(err, &luaErr) {
		return nil, luaErr
	}
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (e *Engine) evaluateBlock(block ast.Block) ([]value.Value, error) {
	e.enterNewScope()
	defer e.leaveScope()

	for _, stmt := range block.StatementsWithoutLast() {
		if _, err := e.evaluateStatement(stmt); err != nil {
			return nil, fmt.Errorf("stmt: %w", err)
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
	case ast.LocalFunction:
		return e.evaluateLocalFunction(s)
	case ast.IfBlock:
		return e.evaluateIfBlock(s)
	case ast.DoBlock:
		return e.evaluateDoBlock(s)
	case ast.LastStatement:
		return e.evaluateLastStatement(s)
	case ast.RepeatBlock:
		return e.evaluateRepeatBlock(s)
	case ast.WhileBlock:
		return e.evaluateWhileBlock(s)
	case ast.ForBlock:
		return e.evaluateForBlock(s)
	case ast.ForInBlock:
		return e.evaluateForInBlock(s)
	}
	return nil, fmt.Errorf("%T unsupported", stmt)
}

func (e *Engine) evaluateForInBlock(block ast.ForInBlock) ([]value.Value, error) {
	exps, err := e.evaluateExpList(block.In)
	if err != nil {
		return nil, fmt.Errorf("explist: %w", err)
	}

	if len(exps) != 3 {
		return nil, fmt.Errorf("explist didn't evaluate to 3 expressions, iter, state and initial, only got %d", len(exps))
	}

	iter := exps[0].(*value.Function)
	state := exps[1]
	init := exps[2]

	e.enterNewScope()
	defer e.leaveScope()
	defer recoverBreak()

	forScope := e.currentScope()

	for {
		vars, err := e.call(iter, state, init)
		if err != nil {
			return nil, fmt.Errorf("call iter: %w", err)
		}
		if len(vars) == 0 || vars[0] == nil || vars[0] == value.Nil {
			break
		}
		init = vars[0]

		for i, name := range block.NameList {
			if len(vars) > i {
				e.assign(forScope, name.Value(), vars[i])
			} else {
				e.assign(forScope, name.Value(), value.Nil)
			}
		}

		_, err = e.evaluateBlock(block.Do)
		if err != nil {
			return nil, fmt.Errorf("block: %w", err)
		}
	}

	return nil, nil
}

func (e *Engine) evaluateForBlock(block ast.ForBlock) ([]value.Value, error) {
	var from, to, step float64

	results, err := e.evaluateExpression(block.From)
	if err != nil {
		return nil, fmt.Errorf("exp (from): %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("expression didn't evaluate to any value")
	}

	fromNums, err := e.tonumber(results[0])
	if err != nil {
		return nil, fmt.Errorf("need a number as from-argument in for loop")
	}
	from = fromNums[0].(value.Number).Value()

	results, err = e.evaluateExpression(block.To)
	if err != nil {
		return nil, fmt.Errorf("exp (to): %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("expression didn't evaluate to any value")
	}

	toNums, err := e.tonumber(results[0])
	if err != nil {
		return nil, fmt.Errorf("need a number as to-argument in for loop")
	}
	to = toNums[0].(value.Number).Value()

	if block.Step != nil {
		results, err = e.evaluateExpression(block.Step)
		if err != nil {
			return nil, fmt.Errorf("exp (step): %w", err)
		}
		if len(results) == 0 {
			return nil, fmt.Errorf("expression didn't evaluate to any value")
		}

		stepNums, err := e.tonumber(results[0])
		if err != nil {
			return nil, fmt.Errorf("need a number as step-argument in for loop")
		}
		step = stepNums[0].(value.Number).Value()
	} else {
		step = 1 // default value for step
	}

	e.enterNewScope()
	defer e.leaveScope()
	defer recoverBreak()

	forScope := e.currentScope()

	// begin implementation as stated in the documentation

	from -= step
	for {
		from += step
		if (step >= 0 && from > to) || (step < 0 && from < to) {
			break
		}
		e.assign(forScope, block.Name.Value(), value.NewNumber(from))

		_, err = e.evaluateBlock(block.Do)
		if err != nil {
			return nil, fmt.Errorf("block: %w", err)
		}
	}
	return nil, nil
}

func (e *Engine) evaluateLocalFunction(fn ast.LocalFunction) ([]value.Value, error) {
	fnName := fn.Name.Value()

	luaFn, err := e.createCallable(fn.FuncBody.ParList, fn.FuncBody.Block)
	if err != nil {
		return nil, fmt.Errorf("create callable: %w", err)
	}

	functionValue := value.NewFunction(fnName, luaFn)
	e.assign(e.currentScope(), fnName, functionValue)
	return nil, nil
}

func (e *Engine) evaluateWhileBlock(block ast.WhileBlock) ([]value.Value, error) {
	defer recoverBreak()

	for {
		results, err := e.evaluateExpression(block.While)
		if err != nil {
			return nil, fmt.Errorf("exp: %w", err)
		}
		if len(results) == 0 {
			return nil, fmt.Errorf("expression didn't evaluate to any value")
		}
		result := results[0]
		if !e.valueIsLogicallyTrue(result) {
			break
		}

		_, err = e.evaluateBlock(block.Do)
		if err != nil {
			return nil, fmt.Errorf("block: %w", err)
		}
	}
	return nil, nil
}

func (e *Engine) evaluateRepeatBlock(block ast.RepeatBlock) ([]value.Value, error) {
	defer recoverBreak()

	for {
		_, err := e.evaluateBlock(block.Repeat)
		if err != nil {
			return nil, fmt.Errorf("block: %w", err)
		}

		results, err := e.evaluateExpression(block.Until)
		if err != nil {
			return nil, fmt.Errorf("exp: %w", err)
		}
		if len(results) == 0 {
			return nil, fmt.Errorf("expression didn't evaluate to any value")
		}
		result := results[0]
		if e.valueIsLogicallyTrue(result) {
			break
		}
	}
	return nil, nil
}

func (e *Engine) evaluateLastStatement(stmt ast.LastStatement) ([]value.Value, error) {
	if stmt.Break {
		panic(Break{})
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

		for i := 0; i < len(values) && i < len(assignment.VarList); i++ {
			if err := e.evaluateAssign(assignment.VarList[i], values[i]); err != nil {
				return fmt.Errorf("assign: %w", err)
			}
		}
		return nil
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
	if len(v.Fragments) == 0 {
		name := v.Name.Value()
		scope := e._G

		// find a scope where the variable might already exist
		varName := value.NewString(name)
		for i := 0; i < len(e.scopes); i++ {
			if _, ok := e.scopes[i].Fields[varName]; ok {
				scope = e.scopes[i]
			}
		}

		e.assign(scope, name, val)
		return nil
	}

	targetExp := ast.PrefixExp{
		Name:      v.Name,
		Exp:       v.Exp,
		Fragments: v.Fragments,
	}

	targetExp.Fragments = v.Fragments[:len(v.Fragments)-1]

	lastFragment := v.Fragments[len(v.Fragments)-1]
	if lastFragment.Args != nil {
		return fmt.Errorf("cannot assign to a function call")
	}

	targets, err := e.evaluatePrefixExpression(targetExp)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return fmt.Errorf("cannot assign to nil value")
	}
	target := targets[0]

	table, ok := target.(*value.Table)
	if !ok {
		return fmt.Errorf("cannot index variable of type %s", target.Type())
	}

	if lastFragment.Name != nil {
		e.assign(table, lastFragment.Name.Value(), val)
	} else {
		results, err := e.evaluateExpression(lastFragment.Exp)
		if err != nil {
			return fmt.Errorf("index exp: %w", err)
		}
		if len(results) == 0 {
			return fmt.Errorf("index exp didn't evaluate to any value")
		}
		return e.performCreateIndex(table, results[0], val)
	}
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
	case ast.TableConstructor:
		return e.evaluateTableConstructor(ex)
	default:
		return nil, fmt.Errorf("%T unsupported", exp)
	}
}

func (e *Engine) evaluateTableConstructor(tblCtor ast.TableConstructor) ([]value.Value, error) {
	tbl := value.NewTable()

	anonymousFieldIndex := 1
	for _, field := range tblCtor.Fields {
		var key value.Value
		if field.Anonymous() {
			key = value.NewNumber(float64(anonymousFieldIndex))
			anonymousFieldIndex++
		} else if field.LeftName != nil {
			key = value.NewString(field.LeftName.Value())
		} else {
			vals, err := e.evaluateExpression(field.LeftExp)
			if err != nil {
				return nil, fmt.Errorf("left exp: %w", err)
			}
			if len(vals) == 0 {
				return nil, fmt.Errorf("left exp didn't evaluate to any value")
			}
			key = vals[0]
		}

		vals, err := e.evaluateExpression(field.RightExp)
		if err != nil {
			return nil, fmt.Errorf("right exp: %w", err)
		}
		if len(vals) == 0 {
			return nil, fmt.Errorf("right exp didn't evaluate to any value")
		}
		val := vals[0]
		tbl.Set(key, val)
	}
	return values(tbl), nil
}

func (e *Engine) evaluateUnopExpression(exp ast.UnopExp) ([]value.Value, error) {
	operands, err := e.evaluateExpression(exp.Exp)
	if err != nil {
		return nil, fmt.Errorf("operand: %w", err)
	}
	operand := operands[0]

	var event string
	var metaMethod *value.Function

	switch exp.Unop.Value() {
	case "-":
		if num, ok := operand.(value.Number); ok {
			results, err := e.multiply(value.NewNumber(-1), num)
			if err != nil {
				return nil, fmt.Errorf("arithmetic unary negation: %w", err)
			}
			return results, nil
		}
		event = "__unm"
	case "not":
		if boolVal, ok := operand.(value.Boolean); ok {
			if boolVal {
				return values(value.False), nil
			}
			return values(value.True), nil
		}
		if e.valueIsLogicallyTrue(operand) {
			return values(value.False), nil
		}
		return values(value.True), nil
	case "~":
		if num, ok := operand.(value.Number); ok {
			results, err := e.evaluateBitwiseNot(num)
			if err != nil {
				return nil, err
			}
			return results, nil
		}
		event = "__bnot"
	case "#":
		if operand.Type() == value.TypeString || operand.Type() == value.TypeTable {
			results, err := e.evaluateLen(operand)
			if err != nil {
				return nil, err
			}
			return results, nil
		}
		// metamethod needs to be called inside evaluateLen
	}

	if event != "" {
		metaMethod, err = e.metaMethodFunction(operand, event)
		if err != nil {
			return nil, fmt.Errorf("metaMethodFunction: %w", err)
		}
	}

	if metaMethod == nil {
		return nil, fmt.Errorf("unsupported unary operator '%s' on %s", exp.Unop.Value(), operand.Type())
	}

	results, err := e.call(metaMethod, operand)
	if err != nil {
		return nil, err
	}
	return results, nil
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
	case "^":
		operator = e.evaluatePower
	case "==":
		operator = e.evaluateEqual
	case "~=":
		operator = e.evaluateUnequal
	case "~":
		operator = e.evaluateBitwiseXor
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

func (e *Engine) evaluateLen(val value.Value) ([]value.Value, error) {
	results, err := e.length(val)
	if err != nil {
		return nil, fmt.Errorf("len: %w", err)
	}
	return results, nil
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

func (e *Engine) evaluateBitwiseXor(left, right value.Value) ([]value.Value, error) {
	results, err := e.bitwiseXor(left, right)
	if err != nil {
		return nil, fmt.Errorf("bitwise xor: %w", err)
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

func (e *Engine) evaluatePower(left, right value.Value) ([]value.Value, error) {
	results, err := e.power(left, right)
	if err != nil {
		return nil, fmt.Errorf("power: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateEqual(left, right value.Value) ([]value.Value, error) {
	results, err := e.cmpEqual(left, right)
	if err != nil {
		return nil, fmt.Errorf("compare: %w", err)
	}
	return results, nil
}

func (e *Engine) evaluateUnequal(left, right value.Value) ([]value.Value, error) {
	results, err := e.evaluateEqual(left, right)
	if err != nil {
		return nil, err
	}

	result := results[0]
	if e.valueIsLogicallyTrue(result) {
		return values(value.False), nil
	}
	return values(value.True), nil
}

func (e *Engine) evaluateLess(left, right value.Value) ([]value.Value, error) {
	if !(left.Type() == value.TypeNumber && right.Type() == value.TypeNumber) &&
		!(left.Type() == value.TypeString && right.Type() == value.TypeString) {
		results, ok, err := e.binaryMetaMethodOperation("__lt", left, right)
		if !ok {
			if err != nil {
				return nil, err
			}
		} else {
			if len(results) < 1 {
				return nil, fmt.Errorf("__lt did not evaluate to any value")
			}
			if e.valueIsLogicallyTrue(results[0]) {
				return values(value.True), nil
			}
			return values(value.False), nil
		}
	}

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
	if !(left.Type() == value.TypeNumber && right.Type() == value.TypeNumber) &&
		!(left.Type() == value.TypeString && right.Type() == value.TypeString) {
		// use two events, first __le, then __lt
		results, ok, err := e.binaryMetaMethodOperation("__le", left, right)
		if !ok {
			if err != nil {
				return nil, err
			}
		} else {
			if len(results) < 1 {
				return nil, fmt.Errorf("%s did not evaluate to any value", "__le")
			}
			if e.valueIsLogicallyTrue(results[0]) {
				return values(value.True), nil
			}
			return values(value.False), nil
		}

		results, ok, err = e.binaryMetaMethodOperation("__lt", right, left)
		if !ok {
			if err != nil {
				return nil, err
			}
		} else {
			if len(results) < 1 {
				return nil, fmt.Errorf("%s did not evaluate to any value", "__lt")
			}
			if !e.valueIsLogicallyTrue(results[0]) {
				return values(value.True), nil
			}
			return values(value.False), nil
		}
	}

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
		isLast := i < len(exp.Fragments)-1

		if current == nil {
			if fragment.Args != nil {
				return nil, fmt.Errorf("cannot call nil value of variable '%s'", currentName)
			}
			return nil, fmt.Errorf("cannot index nil value of variable '%s'", currentName)
		}

		results = nil

		if fragment.Exp != nil {
			table, ok := current.(*value.Table)
			if !ok {
				return nil, fmt.Errorf("cannot index variable of type %s", current.Type())
			}
			vals, err := e.evaluateExpression(fragment.Exp)
			if err != nil {
				return nil, fmt.Errorf("index exp: %w", err)
			}
			if len(vals) == 0 {
				return nil, fmt.Errorf("index exp didn't evaluate to any value")
			}
			indexKey := vals[0]

			indexResults, err := e.performIndexOperation(table, indexKey)
			if err != nil {
				return nil, fmt.Errorf("index: %w", err)
			}
			results = indexResults
			current = indexResults[0]
			currentName = currentName + "[<index>]"
		} else {
			if fragment.Name != nil {
				table, ok := current.(*value.Table)
				if !ok {
					return nil, fmt.Errorf("cannot index variable of type %s", current.Type())
				}
				current, ok = table.Get(value.NewString(fragment.Name.Value()))
				if !ok {
					return nil, fmt.Errorf("variable '%s' has no field '%s'", currentName, fragment.Name.Value())
				}
				results = values(current)
				currentName = fragment.Name.Value()
			}

			if fragment.Args != nil {
				// this fragment is a function call

				args, err := e.evaluateArgs(*fragment.Args)
				if err != nil {
					return nil, fmt.Errorf("args: %w", err)
				}

				res, err := e.attemptCall(current, args...)
				if err != nil {
					return nil, fmt.Errorf("call '%s': %w", currentName, err)
				}
				results = res
				if len(res) == 0 && isLast {
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
