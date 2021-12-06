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
		{`"a" == "a"`, true},
		{`"a" == "b"`, false},
		{`"a" != "a"`, false},
		{`"a" != "b"`, true},
		{`"a" < "b"`, true},
		{`"a" < "a"`, false},
		{`"a" > "b"`, false},
		{`"a" > "a"`, false},
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
		{
			// This should terminate when i == 5.
			// The i < 10 check avoids an infinite loop if the return doesn't work.
			`
			let i = 0; 
			for (; i < 10; i = i + 1) { 
				if (i == 5) { return 1; } 
			}
			i;
			`,
			5,
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
			"fn(...x, y) { }",
			"variodic parameter must be the last parameter of a function",
		},
		{
			"fn(x) { let x = 0; }(1)",
			"identifier already exists: x",
		},
		{
			"fn(x) { }()",
			"wrong number of arguments: expected=1, got=0",
		},
		{
			"fn(x, y) { }(1)",
			"wrong number of arguments: expected=2, got=1",
		},
		{
			"fn() { }(1)",
			"wrong number of arguments: expected=0, got=1",
		},
		{
			"fn(x) { }(1, 2)",
			"wrong number of arguments: expected=1, got=2",
		},
		{
			"fn(x, ...y) { }()",
			"wrong number of arguments: expected=2, got=0",
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
		{
			"let x = for (;;) {}",
			"cannot assign empty value to variable",
		},
		{
			`"Hello" - "World"`,
			"unknown operator: STRING - STRING",
		},
		{
			`{"name": "Monkey"}[fn(x) { x }];`,
			"unusable as hash key: FUNCTION",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		testErrorObject(t, evaluated, tt.expectedMessage)
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
		expected interface{}
	}{
		{"let identity = fn(x) { x; }; identity(5);", 5},
		{"let identity = fn(x) { return x; }; identity(5);", 5},
		{"let double = fn(x) { x * 2; }; double(5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5, 5);", 10},
		{"let add = fn(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"fn(x) { x; }(5)", 5},

		{"fn(...x) { x; }()", []interface{}{}},
		{"fn(...x) { x; }(5)", []interface{}{5}},
		{"fn(...x) { x; }(5, 10)", []interface{}{5, 10}},
		{"fn(x, ...y) { y; }(5)", []interface{}{}},
		{"fn(x, ...y) { y; }(5, 10)", []interface{}{10}},
	}

	for _, tt := range tests {
		testObject(t, testEval(tt.input), tt.expected)
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
func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)

	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(input)

	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len([])`, 0},
		{`len([1])`, 1},
		{`len([1, 2])`, 2},
		{`len(1)`, "argument to `len` not supported: INTEGER"},
		{`len("one", "two")`, "wrong number of arguments: expected=1, got=2"},

		{`first("foobar")`, "argument to `first` not supported: STRING"},
		{`first()`, "wrong number of arguments: expected=1, got=0"},
		{`first([], [])`, "wrong number of arguments: expected=1, got=2"},
		{`first([])`, "array has no elements"},
		{`first([1])`, 1},
		{`first([1, 2])`, 1},

		{`last("foobar")`, "argument to `last` not supported: STRING"},
		{`last()`, "wrong number of arguments: expected=1, got=0"},
		{`last([], [])`, "wrong number of arguments: expected=1, got=2"},
		{`last([])`, "array has no elements"},
		{`last([1])`, 1},
		{`last([1, 2])`, 2},

		{`rest("foobar")`, "argument to `rest` not supported: STRING"},
		{`rest()`, "wrong number of arguments: expected=1, got=0"},
		{`rest([], [])`, "wrong number of arguments: expected=1, got=2"},
		{`rest([])`, "array has no elements"},
		{`rest([1])`, []interface{}{}},
		{`rest([1, 2])`, []interface{}{2}},

		{`push("foobar", [])`, "argument to `push` not supported: STRING"},
		{`push()`, "wrong number of arguments: expected=2, got=0"},
		{`push([])`, "wrong number of arguments: expected=2, got=1"},
		{`push([], 1)`, []interface{}{1}},
		{`push([1], 2)`, []interface{}{1, 2}},
		{`push([1], "foobar")`, []interface{}{1, "foobar"}},
		{`push([1], [2])`, []interface{}{1, []interface{}{2}}},

		{`pop("foobar")`, "argument to `pop` not supported: STRING"},
		{`pop()`, "wrong number of arguments: expected=1, got=0"},
		{`pop([], [])`, "wrong number of arguments: expected=1, got=2"},
		{`pop([])`, "array has no elements"},
		{`pop([1])`, []interface{}{}},
		{`pop([1, 2])`, []interface{}{1}},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case []interface{}:
			testArrayObject(t, evaluated, expected)
		case string:
			testErrorObject(t, evaluated, expected)
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"
	evaluated := testEval(input)

	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d",
			len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"let i = 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[2];",
			3,
		},
		{
			"let myArray = [1, 2, 3]; myArray[0] + myArray[1] + myArray[2];",
			6,
		},
		{
			"let myArray = [1, 2, 3]; let i = myArray[0]; myArray[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			"index outside array bounds: 3",
		},
		{
			"[1, 2, 3][-1]",
			"array index must be non-negative: -1",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case string:
			testErrorObject(t, evaluated, expected)
		}
	}
}

func TestHashLiterals(t *testing.T) {
	input := `let two = "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`

	evaluated := testEval(input)

	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}
		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`let key = "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
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

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	env := object.NewEnvironment()

	return Eval(program, env)
}

func testObject(t *testing.T, obj object.Object, expected interface{}) bool {
	switch expected := expected.(type) {
	case int:
		return testIntegerObject(t, obj, int64(expected))
	case string:
		return testStringObject(t, obj, expected)
	case bool:
		return testBooleanObject(t, obj, expected)
	case []interface{}:
		return testArrayObject(t, obj, expected)
	case nil:
		return testNullObject(t, obj)
	default:
		t.Fatalf("Unsupported expected type (this is a bug in the test). got=%T", expected)
		return false
	}
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

func testStringObject(t *testing.T, obj object.Object, expected string) bool {
	str, ok := obj.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", obj, obj)
		return false
	}

	if str.Value != expected {
		t.Fatalf("object has wrong value. Expected=%s, got=%s",
			expected, str.Value)
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

func testArrayObject(t *testing.T, obj object.Object, expected []interface{}) bool {
	arr, ok := obj.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", obj, obj)
		return false
	}

	if len(arr.Elements) != len(expected) {
		t.Fatalf("wrong number of items in array. expected=%d, got=%d", len(expected), len(arr.Elements))
		return false
	}

	for idx, item := range expected {
		actual := arr.Elements[idx]
		if !testObject(t, actual, item) {
			return false
		}
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

func testErrorObject(t *testing.T, obj object.Object, msg string) bool {
	errObj, ok := obj.(*object.Error)
	if !ok {
		t.Errorf("object is not object.ErrorObject. got=%T(%+v)", obj, obj)
		return false
	}

	if errObj.Message != msg {
		t.Errorf("wrong error message. expected=%q, got=%q",
			msg, errObj.Message)
		return false
	}

	return true
}
