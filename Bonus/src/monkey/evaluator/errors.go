package evaluator

import (
	"fmt"
	"monkey/object"
)

func newError(message string, args ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(message, args...)}
}

func wrongNumberOfArgumentsError(expected, actual int) *object.Error {
	return newError("wrong number of arguments: expected=%d, got=%d", expected, actual)
}

func unsupportedArgumentType(name string, arg object.Object) *object.Error {
	return newError("argument to `%s` not supported: %s", name, arg.Type())
}
