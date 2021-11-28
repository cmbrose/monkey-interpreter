package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"testing"
)

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input              string
		expectedIdentifier string
		expectedValue      interface{}
	}{
		{"let x = 5;", "x", 5},
		{"let y = true;", "y", true},
		{"let foobar = y;", "foobar", "y"},
	}

	for _, tt := range tests {
		program := parseAndCheckErrors(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testLetStatement(t, stmt, tt.expectedIdentifier, tt.expectedValue) {
			return
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input         string
		expectedValue interface{}
	}{
		{"return 5;", 5},
		{"return 993322;", 993322},
		{"return true;", true},
		{"return y;", "y"},
	}

	for _, tt := range tests {
		program := parseAndCheckErrors(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statements. got=%d",
				len(program.Statements))
		}

		stmt := program.Statements[0]
		if !testReturnStatement(t, stmt) {
			return
		}

		val := stmt.(*ast.ReturnStatement).ReturnValue
		if !testLiteralExpression(t, val, tt.expectedValue) {
			return
		}
	}
}

func TestIdentifier(t *testing.T) {
	input := "foobar;"

	program := parseAndCheckErrors(input, t)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have the expected number of statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement, got=%T",
			program.Statements[0])
	}

	if !testIdentifier(t, stmt.Expression, "foobar") {
		return
	}
}

func TestIntegerLiteralExpression(t *testing.T) {
	input := "5;"

	program := parseAndCheckErrors(input, t)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have the expected number of statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement, got=%T",
			program.Statements[0])
	}

	intLit, ok := stmt.Expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp is not *ast.IntegerLiteral. got=%T", stmt.Expression)
	}
	if intLit.Value != 5 {
		t.Errorf("ident.Value not %d. got=%d", 5, intLit.Value)
	}
	if intLit.TokenLiteral() != "5" {
		t.Errorf("ident.TokenLiteral not %s. got=%s", "5", intLit.TokenLiteral())
	}
}

func TestStringLiteralExpression(t *testing.T) {
	input := `"hello world";`
	program := parseAndCheckErrors(input, t)

	if len(program.Statements) != 1 {
		t.Fatalf("program does not have the expected number of statements. got=%d",
			len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement, got=%T",
			program.Statements[0])
	}

	literal, ok := stmt.Expression.(*ast.StringLiteral)
	if !ok {
		t.Fatalf("exp not *ast.StringLiteral. got=%T", stmt.Expression)
	}

	if literal.Value != "hello world" {
		t.Errorf("literal.Value not %q. got=%q", "hello world", literal.Value)
	}
}

func TestParsingPrefixExpressions(t *testing.T) {
	prefixTests := []struct {
		input    string
		operator string
		value    interface{}
	}{
		{"!5", "!", 5},
		{"-15", "-", 15},
		{"!true;", "!", true},
		{"!false;", "!", false},
	}

	for _, tt := range prefixTests {
		program := parseAndCheckErrors(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program does not have the expected number of statements. got=%d",
				len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not *ast.ExpressionStatement, got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.PrefixExpression)
		if !ok {
			t.Fatalf("exp is not *ast.PrefixExpression. got=%T", stmt.Expression)
		}
		if exp.Operator != tt.operator {
			t.Fatalf("exp.Operator is not '%s'. got=%s",
				tt.operator, exp.Operator)
		}
		if !testLiteralExpression(t, exp.Right, tt.value) {
			return
		}
	}
}

func TestParsingInfixExpressions(t *testing.T) {
	infixTests := []struct {
		input      string
		leftValue  interface{}
		operator   string
		rightValue interface{}
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 > 5", 5, ">", 5},
		{"5 < 5", 5, "<", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
		{"true == true", true, "==", true},
		{"true != false", true, "!=", false},
		{"false == false", false, "==", false},
		{"x = 1", "x", "=", 1},
	}

	for _, tt := range infixTests {
		program := parseAndCheckErrors(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		if !testInfixExpression(t, stmt.Expression, tt.leftValue, tt.operator, tt.rightValue) {
			return
		}
	}
}

func TestOperatorPrecedenceParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			"-a * b",
			"((-a) * b);",
		},
		{
			"!-a",
			"(!(-a));",
		},
		{
			"a + b + c",
			"((a + b) + c);",
		},
		{
			"a + b - c",
			"((a + b) - c);",
		},
		{
			"a * b * c",
			"((a * b) * c);",
		},
		{
			"a * b / c",
			"((a * b) / c);",
		},
		{
			"a + b / c",
			"(a + (b / c));",
		},
		{
			"a + b * c + d / e - f",
			"(((a + (b * c)) + (d / e)) - f);",
		},
		{
			"3 + 4; -5 * 5",
			"(3 + 4);((-5) * 5);",
		},
		{
			"5 > 4 == 3 < 4",
			"((5 > 4) == (3 < 4));",
		},
		{
			"5 < 4 != 3 > 4",
			"((5 < 4) != (3 > 4));",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)));",
		},
		{
			"3 + 4 * 5 == 3 * 1 + 4 * 5",
			"((3 + (4 * 5)) == ((3 * 1) + (4 * 5)));",
		},
		{
			"true",
			"true;",
		},
		{
			"false",
			"false;",
		},
		{
			"3 > 5 == false",
			"((3 > 5) == false);",
		},
		{
			"3 < 5 == true",
			"((3 < 5) == true);",
		},
		{
			"1 + (2 + 3) + 4",
			"((1 + (2 + 3)) + 4);",
		},
		{
			"(5 + 5) * 2",
			"((5 + 5) * 2);",
		},
		{
			"2 / (5 + 5)",
			"(2 / (5 + 5));",
		},
		{
			"-(5 + 5)",
			"(-(5 + 5));",
		},
		{
			"!(true == true)",
			"(!(true == true));",
		},
		{
			"a + add(b * c) + d",
			"((a + add((b * c))) + d);",
		},
		{
			"add(a, b, 1, 2 * 3, 4 + 5, add(6, 7 * 8))",
			"add(a, b, 1, (2 * 3), (4 + 5), add(6, (7 * 8)));",
		},
		{
			"add(a + b + c * d / f + g)",
			"add((((a + b) + ((c * d) / f)) + g));",
		},
		{
			"x = 1",
			"(x = 1);",
		},
		{
			"x = 1 + 1",
			"(x = (1 + 1));",
		},
		{
			"x = (1 + 1)",
			"(x = (1 + 1));",
		},
		{
			"(x = 1) + 1",
			"((x = 1) + 1);",
		},
	}

	for _, tt := range tests {
		program := parseAndCheckErrors(tt.input, t)
		actual := program.String()
		if actual != tt.expected {
			t.Errorf("expected=%q, got=%q", tt.expected, actual)
		}
	}
}

func TestBooleanExpression(t *testing.T) {
	tests := []struct {
		input           string
		expectedBoolean bool
	}{
		{"true;", true},
		{"false;", false},
	}

	for _, tt := range tests {
		program := parseAndCheckErrors(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program has not enough statements. got=%d",
				len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		boolean, ok := stmt.Expression.(*ast.Boolean)
		if !ok {
			t.Fatalf("exp not *ast.Boolean. got=%T", stmt.Expression)
		}
		if boolean.Value != tt.expectedBoolean {
			t.Errorf("boolean.Value not %t. got=%t", tt.expectedBoolean,
				boolean.Value)
		}
	}
}

func TestIfExpression(t *testing.T) {
	type expectedClause struct {
		condition             string
		consequenceIdentifier string
	}

	tests := []struct {
		input                         string
		expectedClauses               []expectedClause
		expectedAlternativeIdentifier interface{}
	}{
		{
			"if (x < y) { x }",
			[]expectedClause{
				{"(x < y)", "x"},
			},
			nil,
		},
		{
			"if (x < y) { x } else if (y < z) { y }",
			[]expectedClause{
				{"(x < y)", "x"},
				{"(y < z)", "y"},
			},
			nil,
		},

		{
			"if (x < y) { x } else if (y < z) { y } else if (z < w) { z }",
			[]expectedClause{
				{"(x < y)", "x"},
				{"(y < z)", "y"},
				{"(z < w)", "z"},
			},
			nil,
		},
		{
			"if (x < y) { x } else { y }",
			[]expectedClause{
				{"(x < y)", "x"},
			},
			"y",
		},
		{
			"if (x < y) { x } else if (y < z) { y } else { z }",
			[]expectedClause{
				{"(x < y)", "x"},
				{"(y < z)", "y"},
			},
			"z",
		},
		{
			"if (x < y) { x } else if (y < z) { y } else if (z < w) { z } else { w }",
			[]expectedClause{
				{"(x < y)", "x"},
				{"(y < z)", "y"},
				{"(z < w)", "z"},
			},
			"w",
		},
		{
			"if (x < y) x else if (y < z) y else if (z < w) z else w",
			[]expectedClause{
				{"(x < y)", "x"},
				{"(y < z)", "y"},
				{"(z < w)", "z"},
			},
			"w",
		},
	}

	for _, tt := range tests {
		program := parseAndCheckErrors(tt.input, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Body does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
				program.Statements[0])
		}

		exp, ok := stmt.Expression.(*ast.IfExpression)
		if !ok {
			t.Fatalf("stmt.Expression is not ast.IfExpression. got=%T",
				stmt.Expression)
		}

		if len(exp.Clauses) != len(tt.expectedClauses) {
			t.Fatalf("Incorrect number of clauses. Expected %d. got=%d",
				len(tt.expectedClauses), len(exp.Clauses))
		}

		for idx, expectedClause := range tt.expectedClauses {
			actualClause := exp.Clauses[idx]

			if actualClause.Condition.String() != expectedClause.condition {
				t.Fatalf("Incorrect condition for clause %d. Expected %s. got=%s",
					idx+1, expectedClause.condition, actualClause.Condition.String())
			}

			if len(actualClause.Consequence.Statements) != 1 {
				t.Errorf("consequence is not 1 statements. got=%d\n",
					len(actualClause.Consequence.Statements))
			}

			consequence, ok := actualClause.Consequence.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
					actualClause.Consequence.Statements[0])
			}

			if !testIdentifier(t, consequence.Expression, expectedClause.consequenceIdentifier) {
				return
			}
		}

		if (tt.expectedAlternativeIdentifier != nil) && (exp.Alternative == nil) {
			t.Errorf("exp.Alternative.Statements was nil but expected %+v", tt.expectedAlternativeIdentifier)
		} else if (tt.expectedAlternativeIdentifier == nil) && (exp.Alternative != nil) {
			t.Errorf("exp.Alternative.Statements was not nil. got=%+v", exp.Alternative)
		} else if tt.expectedAlternativeIdentifier != nil {
			if len(exp.Alternative.Statements) != 1 {
				t.Errorf("alternative is not 1 statements. got=%d\n",
					len(exp.Alternative.Statements))
			}

			alternative, ok := exp.Alternative.Statements[0].(*ast.ExpressionStatement)
			if !ok {
				t.Fatalf("Statements[0] is not ast.ExpressionStatement. got=%T",
					exp.Alternative.Statements[0])
			}

			if !testIdentifier(t, alternative.Expression, tt.expectedAlternativeIdentifier.(string)) {
				return
			}
		}
	}
}

func TestFunctionLiteralParsing(t *testing.T) {
	input := `fn(x, y) { x + y; }`

	program := parseAndCheckErrors(input, t)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	function, ok := stmt.Expression.(*ast.FunctionLiteral)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.FunctionLiteral. got=%T",
			stmt.Expression)
	}

	if len(function.Parameters) != 2 {
		t.Fatalf("function literal parameters wrong. want 2, got=%d\n",
			len(function.Parameters))
	}

	testLiteralExpression(t, function.Parameters[0], "x")
	testLiteralExpression(t, function.Parameters[1], "y")

	if len(function.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements has not 1 statements. got=%d\n",
			len(function.Body.Statements))
	}

	bodyStmt, ok := function.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not ast.ExpressionStatement. got=%T",
			function.Body.Statements[0])
	}

	testInfixExpression(t, bodyStmt.Expression, "x", "+", "y")
}

func TestFunctionParameterParsing(t *testing.T) {
	tests := []struct {
		input          string
		expectedParams []string
	}{
		{input: "fn() {};", expectedParams: []string{}},
		{input: "fn(x) {};", expectedParams: []string{"x"}},
		{input: "fn(x, y, z) {};", expectedParams: []string{"x", "y", "z"}},
	}
	for _, tt := range tests {
		program := parseAndCheckErrors(tt.input, t)

		stmt := program.Statements[0].(*ast.ExpressionStatement)
		function := stmt.Expression.(*ast.FunctionLiteral)

		if len(function.Parameters) != len(tt.expectedParams) {
			t.Errorf("length parameters wrong. want %d, got=%d\n",
				len(tt.expectedParams), len(function.Parameters))
		}

		for i, ident := range tt.expectedParams {
			testLiteralExpression(t, function.Parameters[i], ident)
		}
	}
}

func TestCallExpressionParsing(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5);"
	program := parseAndCheckErrors(input, t)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("stmt is not ast.ExpressionStatement. got=%T",
			program.Statements[0])
	}

	exp, ok := stmt.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("stmt.Expression is not ast.CallExpression. got=%T",
			stmt.Expression)
	}

	if !testIdentifier(t, exp.Function, "add") {
		return
	}

	if len(exp.Arguments) != 3 {
		t.Fatalf("wrong length of arguments. got=%d", len(exp.Arguments))
	}

	testLiteralExpression(t, exp.Arguments[0], 1)
	testInfixExpression(t, exp.Arguments[1], 2, "*", 3)
	testInfixExpression(t, exp.Arguments[2], 4, "+", 5)
}

func TestForLoopStatementParsing(t *testing.T) {
	input := `for (let x = 0; x < 10; x = x + 1) { x }`

	program := parseAndCheckErrors(input, t)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Body does not contain %d statements. got=%d\n",
			1, len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ForLoopStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ForLoopStatement. got=%T",
			program.Statements[0])
	}

	if !testLetStatement(t, stmt.InitializeStatement, "x", 0) {
		return
	}

	if !testInfixExpression(t, stmt.ContinueExpression, "x", "<", 10) {
		return
	}

	assign, ok := stmt.StepExpression.(*ast.InfixExpression)
	if !ok {
		t.Fatalf("stmt.StepStatement is not ast.InfixExpression. got=%T",
			stmt.StepExpression)
	}

	if !testIdentifier(t, assign.Left, "x") {
		return
	}

	if assign.Operator != "=" {
		t.Fatalf("stmt.StepStatement operator is not %s. got=%s",
			"=", assign.Operator)
	}

	if !testInfixExpression(t, assign.Right, "x", "+", 1) {
		return
	}

	if len(stmt.Body.Statements) != 1 {
		t.Fatalf("function.Body.Statements has not 1 statements. got=%d\n",
			len(stmt.Body.Statements))
	}

	bodyStmt, ok := stmt.Body.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		t.Fatalf("function body stmt is not ast.ExpressionStatement. got=%T",
			stmt.Body.Statements[0])
	}

	testIdentifier(t, bodyStmt.Expression, "x")
}

func TestForLoopStatementInitializeStatementParsing(t *testing.T) {
	tests := []struct {
		input             string
		expectedStatement ast.Statement
	}{
		{
			`let x = 0`,
			&ast.LetStatement{
				Token: token.Token{Literal: "let"},
				Name:  &ast.Identifier{Value: "x"},
				Value: &ast.IntegerLiteral{Token: token.Token{Literal: "0"}},
			},
		},
		{
			`x = 0`,
			&ast.ExpressionStatement{
				Expression: &ast.InfixExpression{
					Left:     &ast.Identifier{Value: "x"},
					Operator: "=",
					Right:    &ast.IntegerLiteral{Token: token.Token{Literal: "0"}},
				},
			},
		},
		{
			``,
			nil,
		},
	}

	for _, tt := range tests {
		forInput := fmt.Sprintf("for (%s; x < 1; x = x + 1){}", tt.input)
		program := parseAndCheckErrors(forInput, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Body does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ForLoopStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ForLoopStatement. got=%T",
				program.Statements[0])
		}

		if tt.expectedStatement == nil {
			if stmt.InitializeStatement != nil {
				t.Fatalf("stmt.InitializeStatement is not nil, got=%T",
					stmt.InitializeStatement)
			}
		} else if stmt.InitializeStatement == nil {
			t.Fatalf("stmt.InitializeStatement is nil, expected=%T",
				tt.expectedStatement)
		} else {
			if fmt.Sprintf("%T", stmt.InitializeStatement) != fmt.Sprintf("%T", tt.expectedStatement) {
				t.Fatalf("stmt.InitializeStatement is not %T, got=%T",
					tt.expectedStatement, stmt.InitializeStatement)
			}

			actual := stmt.InitializeStatement.String()
			if actual != tt.expectedStatement.String() {
				t.Fatalf("stmt.InitializeStatement is not `%s`, got=`%s`",
					tt.expectedStatement.String(), actual)
			}
		}
	}
}

func TestForLoopStatementContinueExpressionParsing(t *testing.T) {
	tests := []struct {
		input              string
		expectedExpression ast.Expression
	}{
		{
			`x < 10`,
			&ast.InfixExpression{
				Left:     &ast.Identifier{Value: "x"},
				Operator: "<",
				Right:    &ast.IntegerLiteral{Token: token.Token{Literal: "10"}},
			},
		},
		{
			``,
			nil,
		},
	}

	for _, tt := range tests {
		forInput := fmt.Sprintf("for (let x = 0; %s; x = x + 1){}", tt.input)
		program := parseAndCheckErrors(forInput, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Body does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ForLoopStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ForLoopStatement. got=%T",
				program.Statements[0])
		}

		if tt.expectedExpression == nil {
			if stmt.ContinueExpression != nil {
				t.Fatalf("stmt.ContinueExpression is not nil, got=%T",
					stmt.ContinueExpression)
			}
		} else if stmt.ContinueExpression == nil {
			t.Fatalf("stmt.ContinueExpression is nil, expected=%T",
				tt.expectedExpression)
		} else {
			if fmt.Sprintf("%T", stmt.ContinueExpression) != fmt.Sprintf("%T", tt.expectedExpression) {
				t.Fatalf("stmt.ContinueExpression is not %T, got=%T",
					tt.expectedExpression, stmt.ContinueExpression)
			}

			actual := stmt.ContinueExpression.String()
			if actual != tt.expectedExpression.String() {
				t.Fatalf("stmt.ContinueExpression is not `%s`, got=`%s`",
					tt.expectedExpression.String(), actual)
			}
		}
	}
}

func TestForLoopStatementStepExpressionParsing(t *testing.T) {
	tests := []struct {
		input              string
		expectedExpression ast.Expression
	}{
		{
			`x = x + 1`,
			&ast.InfixExpression{
				Left:     &ast.Identifier{Value: "x"},
				Operator: "=",
				Right: &ast.InfixExpression{
					Left:     &ast.Identifier{Value: "x"},
					Operator: "+",
					Right:    &ast.IntegerLiteral{Token: token.Token{Literal: "1"}},
				},
			},
		},
		{
			``,
			nil,
		},
	}

	for _, tt := range tests {
		forInput := fmt.Sprintf("for (let x = 0; x < 10; %s){}", tt.input)
		program := parseAndCheckErrors(forInput, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Body does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ForLoopStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ForLoopStatement. got=%T",
				program.Statements[0])
		}

		if tt.expectedExpression == nil {
			if stmt.StepExpression != nil {
				t.Fatalf("stmt.StepExpression is not nil, got=%T",
					stmt.StepExpression)
			}
		} else if stmt.StepExpression == nil {
			t.Fatalf("stmt.StepExpression is nil, expected=%T",
				tt.expectedExpression)
		} else {
			if fmt.Sprintf("%T", stmt.StepExpression) != fmt.Sprintf("%T", tt.expectedExpression) {
				t.Fatalf("stmt.StepExpression is not %T, got=%T",
					tt.expectedExpression, stmt.StepExpression)
			}

			actual := stmt.StepExpression.String()
			if actual != tt.expectedExpression.String() {
				t.Fatalf("stmt.StepExpression is not `%s`, got=`%s`",
					tt.expectedExpression.String(), actual)
			}
		}
	}
}

func TestForLoopStatementBodyParsing(t *testing.T) {
	tests := []struct {
		input             string
		expectedStatement ast.Expression
	}{
		{
			`{ x = x + 1 }`,
			&ast.InfixExpression{
				Left:     &ast.Identifier{Value: "x"},
				Operator: "=",
				Right: &ast.InfixExpression{
					Left:     &ast.Identifier{Value: "x"},
					Operator: "+",
					Right:    &ast.IntegerLiteral{Token: token.Token{Literal: "1"}},
				},
			},
		},
		{
			`x`,
			&ast.Identifier{Value: "x"},
		},
	}

	for _, tt := range tests {
		forInput := fmt.Sprintf("for (let x = 0; x < 10;) %s", tt.input)
		program := parseAndCheckErrors(forInput, t)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Body does not contain %d statements. got=%d\n",
				1, len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.ForLoopStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.ForLoopStatement. got=%T",
				program.Statements[0])
		}

		if len(stmt.Body.Statements) != 1 {
			t.Fatalf("stmt.Body does not contain %d statements. got=%d\n",
				1, len(stmt.Body.Statements))
		}

		expr, ok := stmt.Body.Statements[0].(*ast.ExpressionStatement)
		if !ok {
			t.Fatalf("stmt.Body.Statements[0] is not ast.ExpressionStatement. got=%T",
				stmt.Body.Statements[0])
		}

		if fmt.Sprintf("%T", expr.Expression) != fmt.Sprintf("%T", tt.expectedStatement) {
			t.Fatalf("stmt.Body.Statements[0] is not %T, got=%T",
				tt.expectedStatement, expr.Expression)
		}

		actual := expr.Expression.String()
		if actual != tt.expectedStatement.String() {
			t.Fatalf("stmt.StepExpression is not `%s`, got=`%s`",
				tt.expectedStatement.String(), actual)
		}
	}
}

func testIntegerLiteral(t *testing.T, exp ast.Expression, value int64) bool {
	intLit, ok := exp.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("exp is not *ast.IntegerLiteral. got=%T", exp)
		return false
	}
	if intLit.Value != value {
		t.Errorf("ident.Value not %d. got=%d", value, intLit.Value)
		return false
	}
	if intLit.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d. got=%s", value,
			intLit.TokenLiteral())
		return false
	}
	return true
}

func testIdentifier(t *testing.T, exp ast.Expression, value string) bool {
	ident, ok := exp.(*ast.Identifier)
	if !ok {
		t.Errorf("exp not *ast.Identifier. got=%T", exp)
		return false
	}
	if ident.Value != value {
		t.Errorf("ident.Value not %s. got=%s", value, ident.Value)
		return false
	}
	if ident.TokenLiteral() != value {
		t.Errorf("ident.TokenLiteral not %s. got=%s", value,
			ident.TokenLiteral())
		return false
	}
	return true
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	b, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}
	if b.Value != value {
		t.Errorf("b.Value not %t. got=%t", value, b.Value)
		return false
	}
	if b.TokenLiteral() != fmt.Sprintf("%t", value) {
		t.Errorf("b.TokenLiteral not %t. got=%s", value,
			b.TokenLiteral())
		return false
	}
	return true
}

func testLiteralExpression(
	t *testing.T,
	exp ast.Expression,
	expected interface{},
) bool {
	switch v := expected.(type) {
	case int:
		return testIntegerLiteral(t, exp, int64(v))
	case int64:
		return testIntegerLiteral(t, exp, v)
	case string:
		return testIdentifier(t, exp, v)
	case bool:
		return testBooleanLiteral(t, exp, v)
	}
	t.Errorf("type of exp not handled. got=%T", exp)
	return false
}

func testInfixExpression(
	t *testing.T,
	exp ast.Expression,
	left interface{},
	operator string,
	right interface{},
) bool {
	opExp, ok := exp.(*ast.InfixExpression)
	if !ok {
		t.Errorf("exp is not ast.InfixExpression. got=%T(%s)", exp, exp)
		return false
	}
	if !testLiteralExpression(t, opExp.Left, left) {
		return false
	}
	if opExp.Operator != operator {
		t.Errorf("exp.Operator is not '%s'. got=%q", operator, opExp.Operator)
		return false
	}
	if !testLiteralExpression(t, opExp.Right, right) {
		return false
	}
	return true
}

func testLetStatement(t *testing.T, s ast.Statement, name string, value interface{}) bool {
	if s.TokenLiteral() != "let" {
		t.Fatalf("s.TokenLiteral() was not 'let', got=%s", s.TokenLiteral())
		return false
	}

	letStatement, ok := s.(*ast.LetStatement)
	if !ok {
		t.Fatalf("s not LetStatement, got=%T", s)
		return false
	}

	if letStatement.Name.Value != name {
		t.Fatalf("letStatement.Name was not '%s', got=%s", name, letStatement.Name.Value)
	}

	if letStatement.Name.TokenLiteral() != name {
		t.Errorf("s.Name not '%s', got=%s", name, letStatement.Name)
		return false
	}

	if !testLiteralExpression(t, letStatement.Value, value) {
		return false
	}

	return true
}

func testReturnStatement(t *testing.T, s ast.Statement) bool {
	if s.TokenLiteral() != "return" {
		t.Fatalf("s.TokenLiteral() was not 'return', got=%s", s.TokenLiteral())
		return false
	}

	_, ok := s.(*ast.ReturnStatement)
	if !ok {
		t.Fatalf("s not ReturnStatement, got=%T", s)
		return false
	}

	return true
}

func parseAndCheckErrors(input string, t *testing.T) *ast.Program {
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParseErrors(p, input, t)

	return program
}

func checkParseErrors(p *Parser, input string, t *testing.T) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors parsing %s", len(errors), input)
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}
