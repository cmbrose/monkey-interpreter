package evaluator

import (
	"fmt"
	"monkey/ast"
	"monkey/object"
	"strings"
)

var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

func Eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {
	case *ast.Program:
		return evalProgram(node, env)

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)

	case *ast.IfExpression:
		return evalIfExpression(node, env)

	case *ast.LetStatement:
		value := Eval(node.Value, env)
		if isError(value) {
			return value
		}

		res := env.Add(node.Name.Value, value)
		if isError(res) {
			return res
		}

	case *ast.ReturnStatement:
		value := Eval(node.ReturnValue, env)
		if isError(value) {
			return value
		}

		return &object.ReturnValue{Value: value}

	case *ast.ForLoopStatement:
		value := evalForLoopStatement(node, env)
		if isError(value) {
			return value
		}

	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		return evalInfixExpression(node, env)

	case *ast.CallExpression:
		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		return applyFunction(function, args)

	case *ast.Identifier:
		return evalIdentifier(node, env)

	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.StringLiteral:
		return &object.String{Value: node.Value}

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Body: body, Env: env}
	}

	return nil
}

func evalProgram(prog *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range prog.Statements {
		result = Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

func evalForLoopStatement(stmt *ast.ForLoopStatement, env *object.Environment) object.Object {
	loopEnv := object.NewEnclosedEnvironment(env)

	if stmt.InitializeStatement != nil {
		initializeResult := Eval(stmt.InitializeStatement, loopEnv)
		if isError(initializeResult) {
			return initializeResult
		}
	}

	var continueResult object.Object = TRUE

	if stmt.ContinueExpression != nil {
		continueResult := Eval(stmt.ContinueExpression, loopEnv)
		if isError(continueResult) {
			return continueResult
		}
	}

	for isTruthy(continueResult) {
		result := Eval(stmt.Body, loopEnv)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}

		if stmt.StepExpression != nil {
			stepResult := Eval(stmt.StepExpression, loopEnv)
			if isError(stepResult) {
				return stepResult
			}
		}

		if stmt.ContinueExpression != nil {
			continueResult = Eval(stmt.ContinueExpression, loopEnv)
			if isError(continueResult) {
				return continueResult
			}
		}
	}

	return nil
}

func evalIfExpression(expr *ast.IfExpression, env *object.Environment) object.Object {
	for _, clause := range expr.Clauses {
		result := Eval(clause.Condition, env)

		if isError(result) {
			return result
		}

		if isTruthy(result) {
			blockEnv := object.NewEnclosedEnvironment(env)
			return Eval(clause.Consequence, blockEnv)
		}
	}

	if expr.Alternative != nil {
		blockEnv := object.NewEnclosedEnvironment(env)
		return Eval(expr.Alternative, blockEnv)
	} else {
		return NULL
	}
}

func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

func evalInfixExpression(node *ast.InfixExpression, env *object.Environment) object.Object {
	operator := node.Operator

	// This needs to be tested first or left will resolve a value and not an identifier
	if operator == "=" {
		return evalInfixAssignOperator(node, env)
	}

	left := Eval(node.Left, env)
	if isError(left) {
		return left
	}

	right := Eval(node.Right, env)
	if isError(right) {
		return right
	}

	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)

	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)

	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())

	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)

	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalInfixAssignOperator(node *ast.InfixExpression, env *object.Environment) object.Object {
	if ident, ok := node.Left.(*ast.Identifier); ok {
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixAssignExpression(ident, right, env)
	} else {
		return newError("Left side of assign expression must be a variable")
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftInt, leftOk := left.(*object.Integer)
	rightInt, rightOk := right.(*object.Integer)

	if !leftOk || !rightOk {
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	}

	leftVal := leftInt.Value
	rightVal := rightInt.Value

	switch operator {
	// Arithmetic
	case "+":
		return nativeIntToIntegerObject(leftVal + rightVal)
	case "-":
		return nativeIntToIntegerObject(leftVal - rightVal)
	case "*":
		return nativeIntToIntegerObject(leftVal * rightVal)
	case "/":
		return nativeIntToIntegerObject(leftVal / rightVal)

	// Comparison
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)

	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftStr, leftOk := left.(*object.String)
	rightStr, rightOk := right.(*object.String)

	if !leftOk || !rightOk {
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	}

	leftVal := leftStr.Value
	rightVal := rightStr.Value

	comp := strings.Compare(leftVal, rightVal)

	switch operator {
	// Concatenation
	case "+":
		return &object.String{Value: leftVal + rightVal}

	// Comparison
	case "<":
		return nativeBoolToBooleanObject(comp < 0)
	case ">":
		return nativeBoolToBooleanObject(comp > 0)
	case "==":
		return nativeBoolToBooleanObject(comp == 0)
	case "!=":
		return nativeBoolToBooleanObject(comp != 0)

	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalInfixAssignExpression(left *ast.Identifier, right object.Object, env *object.Environment) object.Object {
	return env.Set(left.Value, right)
}

func evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}

		result = append(result, evaluated)
	}

	return result
}

func applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		if len(fn.Parameters) != len(args) {
			return wrongNumberOfArgumentsError(len(fn.Parameters), len(args))
		}

		extendedEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for paramIdx, param := range fn.Parameters {
		env.Add(param.Value, args[paramIdx])
	}

	return env
}

func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

func evalIdentifier(
	node *ast.Identifier,
	env *object.Environment,
) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}

	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}

	return newError("identifier not found: " + node.Value)
}

func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

func nativeIntToIntegerObject(value int64) *object.Integer {
	return &object.Integer{Value: value}
}

func newError(message string, args ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(message, args...)}
}

func wrongNumberOfArgumentsError(expected, actual int) *object.Error {
	return newError("wrong number of arguments: expected=%d, got=%d", expected, actual)
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR_OBJ
}
