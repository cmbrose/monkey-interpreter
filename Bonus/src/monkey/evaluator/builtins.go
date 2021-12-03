package evaluator

import "monkey/object"

var builtins = map[string]*object.Builtin{
	"len":   {Fn: lenBuiltin},
	"first": {Fn: firstBuiltin},
	"last":  {Fn: lastBuiltin},
	"rest":  {Fn: restBuiltin},
	"push":  {Fn: pushBuiltin},
	"pop":   {Fn: popBuiltin},
}

func lenBuiltin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return wrongNumberOfArgumentsError(1, len(args))
	}

	switch arg := args[0].(type) {
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}

	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}

	default:
		return unsupportedArgumentType("len", args[0])
	}
}

func firstBuiltin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return wrongNumberOfArgumentsError(1, len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		if len(arg.Elements) > 0 {
			return arg.Elements[0]
		}

		return newError("array has no elements")

	default:
		return unsupportedArgumentType("first", args[0])
	}
}

func lastBuiltin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return wrongNumberOfArgumentsError(1, len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		if len(arg.Elements) > 0 {
			return arg.Elements[len(arg.Elements)-1]
		}

		return newError("array has no elements")

	default:
		return unsupportedArgumentType("last", args[0])
	}
}

func restBuiltin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return wrongNumberOfArgumentsError(1, len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		length := len(arg.Elements)
		if length > 0 {
			newElements := make([]object.Object, length-1)
			copy(newElements, arg.Elements[1:length])
			return &object.Array{Elements: newElements}
		}

		return newError("array has no elements")

	default:
		return unsupportedArgumentType("rest", args[0])
	}
}

func pushBuiltin(args ...object.Object) object.Object {
	if len(args) != 2 {
		return wrongNumberOfArgumentsError(2, len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		length := len(arg.Elements)

		newElements := make([]object.Object, length+1)
		copy(newElements, arg.Elements)
		newElements[length] = args[1]
		return &object.Array{Elements: newElements}

	default:
		return unsupportedArgumentType("push", args[0])
	}
}

func popBuiltin(args ...object.Object) object.Object {
	if len(args) != 1 {
		return wrongNumberOfArgumentsError(1, len(args))
	}

	switch arg := args[0].(type) {
	case *object.Array:
		length := len(arg.Elements)
		if length > 0 {
			newElements := make([]object.Object, length-1)
			copy(newElements, arg.Elements[0:length-1])
			return &object.Array{Elements: newElements}
		}

		return newError("array has no elements")

	default:
		return unsupportedArgumentType("pop", args[0])
	}
}
