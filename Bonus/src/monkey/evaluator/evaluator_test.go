package evaluator

import (
	"monkey/lexer"
	"monkey/object"
	"monkey/parser"
	"testing"
)

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}
	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 > 2) { 10 } else if (2 > 3) { 20 }", nil},
		{"if (1 > 2) { 10 } else if (2 > 3) { 20 } else { 30 }", 30},
		{"if (1 > 2) { 10 } else if (2 < 3) { 20 } else { 30 }", 20},
		{"if (1 < 2) { 10 } else if (2 < 3) { 20 } else { 30 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{
			`if (10 > 1) {
				if (10 > 1) {
					return 10;
				}
				return 1;
			}`,
			10,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + true;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + true; 5;",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-true",
			"unknown operator: -BOOLEAN",
		},
		{
			"true + false;",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5; true + false; 5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { true + false; }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
			if (10 > 1) {
				if (10 > 1) {
					return true + false;
				}
				return 1;
			}
			`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			"let a = 5; 5 = 4;",
			"Left side of assign expression must be a variable",
		},
		{
			"for (i = 0; i < 10; i = i + 1) {}",
			"identifier not found: i",
		},
		{
			"x = 0;",
			"identifier not found: x",
		},
		{
			"let x = 0; let x = 0;",
			"identifier already exists: x",
		},
		{
			"fn(x) { let x = 0; }(1)",
			"identifier already exists: x",
		},
		{
			"fn(x) { }()",
			"function called with incorrect number of arguments: expected=1, got=0.",
		},
		{
			"fn(x, y) { }(1)",
			"function called with incorrect number of arguments: expected=2, got=1.",
		},
		{
			"fn() { }(1)",
			"function called with incorrect number of arguments: expected=0, got=1.",
		},
		{
			"fn(x) { }(1, 2)",
			"function called with incorrect number of arguments: expected=1, got=2.",
		},
		{
			"let x = fn() { }()",
			"cannot assign empty value to variable",
		},
		{
			"let x = if (true) { }",
			"cannot assign empty value to variable",
		},
		{
			"let x = 5; x = if (true) { }",
			"cannot assign empty value to variable",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}

		if errObj.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q",
				tt.expectedMessage, errObj.Message)
		}
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a;", 5},
		{"let a = 5 * 5; a;", 25},
		{"let a = 5; let b = a; b;", 5},
		{"let a = 5; let b = a; let c = a + b + 5; c;", 15},
		{"let a = 5; fn() { let a = 4; }(); a;", 5},
		{"let a = 5; if (true) { let a = 4; } a;", 5},
		{
			`
			let a = 5;
			let b = fn() {
				let a = 4;
				fn() {
					a = 5;
				}();
				a;
			}();
			a + b;
			`,
			10,
		},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestAssignExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5; a = 4; a", 4},
		{"let a = 5; a = 4;", 4},
		{"let a = 5; fn() { a = 4; }(); a", 4},
		{"let a = 5; if (true) { a = 4; } a", 4},
		{"let a = 5; if (false) { a = 4; } a", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) { x + 2; };"
	evaluated := testEval(input)

	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "{ (x + 2); }"
	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, fn.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestClosures(t *testing.T) {
	input := `
	let newAdder = fn(x) {
		fn(y) { x + y };
	};
	let addTwo = newAdder(2);
	addTwo(2);`

	testIntegerObject(t, testEval(input), 4)
}

func TestForLoopStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`
			for (let i = 0; i < 10; i = i + 1) {
				i
			}
			`,
			nil,
		},
		{
			`
			let x = 0; 
			for (let i = 0; i < 10; i = i + 1) { 
				x = x + 1; 
			} 
			x;
			`,
			10,
		},
		{
			`
			let i = 0; 
			for (; i < 10; i = i + 1) { } 
			i;
			`,
			10,
		},
		{
			`
			let x = 0; 
			let i = 0; 
			for (i = 5; i < 10; i = i + 1) { 
				x = x + 1; 
			} 
			x;
			`,
			5,
		},
		{
			`
			let i = 0; 
			for (i = 0; i < 10;) { 
				i = i + 1 
			} 
			i;
			`,
			10,
		},
		{
			`
			let x = 0; 
			let i = 5; 
			for (let i = 0; i < 10; i = i + 1) { 
				x = x + 1; 
			} 
			x;
			`,
			10,
		},
		{
			`
			let x = 0;
			for (let i = 0; i < 10; i = i + 1) { 
				for (let i = 0; i < 10; i = i + 1) { 
					x = x + 1; 
				} 
			} 
			x;
			`,
			100,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else if evaluated != nil {
			t.Fatalf("Expected nil but was %t (%+v)", evaluated, evaluated)
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	env := object.NewEnvironment()

	return Eval(program, env)
}

func testIntegerObject(t *testing.T, obj object.Object, expected int64) bool {
	integer, ok := obj.(*object.Integer)
	if !ok {
		t.Fatalf("object is not Integer. got=%T (%+v)", obj, obj)
		return false
	}

	if integer.Value != expected {
		t.Fatalf("object has wrong value. Expected=%d, got=%d",
			expected, integer.Value)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	boolean, ok := obj.(*object.Boolean)
	if !ok {
		t.Fatalf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}

	if boolean.Value != expected {
		t.Fatalf("object has wrong value. Expected=%t, got=%t",
			expected, boolean.Value)
		return false
	}

	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}
	return true
}
