package ast

import (
	"strings"
	"testing"
)

func TestParser(t *testing.T) {
	const input = `(module $testmodule
		(type (func (type 0)))
	)
	`
	l := newLexer(strings.NewReader(input))
	tokens, err := l.lex()
	if err != nil {
		t.Fatal("lexer:", err)
	}
	p := newParser(tokens)
	p.parse()
}
