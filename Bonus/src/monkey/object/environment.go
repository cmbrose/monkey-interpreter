package object

import "fmt"

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

type Environment struct {
	store map[string]Object
	outer *Environment
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]

	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}

	return obj, ok
}

// Creates a new variable, shadowing a variable in the outer scope if applicable.
func (e *Environment) Add(name string, val Object) Object {
	if _, ok := e.store[name]; ok {
		return &Error{fmt.Sprintf("Variable '%s' already exists", name)}
	}

	return e.AddOrSet(name, val)
}

// Updates an existing variable.
// Updates the value in an outer scope if the variable isn't found in the immediate scope.
func (e *Environment) Set(name string, val Object) Object {
	if _, ok := e.store[name]; !ok {
		if e.outer != nil {
			return e.outer.Set(name, val)
		}

		return &Error{fmt.Sprintf("Variable '%s' does not exist", name)}
	}

	return e.AddOrSet(name, val)
}

// Creates or updates a variable in the immediate scope, shadowing a variable in the outer scope if applicable.
func (e *Environment) AddOrSet(name string, val Object) Object {
	e.store[name] = val
	return val
}
