package ast

import (
	"monkey/token"
	"testing"
)

func TestString(t *testing.T) {
	p := &Program{
		Statements: []Statement{
			&LetStatement{
				Token: token.Token{Type: token.LET, Literal: "let"},
				Name: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myVar"},
					Value: "myVar",
				},
				Value: &Identifier{
					Token: token.Token{Type: token.IDENT, Literal: "myOtherVar"},
					Value: "myOtherVar",
				},
			},
		},
	}

	if p.String() != "let myVar = myOtherVar;" {
		t.Errorf("program.String() wrong. got=%q", p.String())
	}
}
