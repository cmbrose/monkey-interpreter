package evaluator

import "monkey/object"

var builtins = map[string]*object.Builtin{
	"len":   {Fn: lenBuiltin},
	"first": {Fn: firstBuiltin},
	"last":  {Fn: lastBuiltin},
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
