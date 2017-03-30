package ast

import (
	"bytes"
	"fmt"
	"testing"
)

var lexertests = []struct {
	in   string
	want []token
}{
	{"  (\n   module \n)    ", []token{
		tok(LPAREN, "("),
		tok(MODULE, "module"),
		tok(RPAREN, ")"),
	}},

	// strings
	{`  ""  `, []token{tSTRING(`""`)}},
	{` "a b c "`, []token{tSTRING(`"a b c "`)}},
	{`   "\""`, []token{tSTRING(`"\""`)}},
	{`  "\\"`, []token{tSTRING(`"\\"`)}},
	{` "\\\""`, []token{tSTRING(`"\\\""`)}},
	{`    "\\\\"`, []token{tSTRING(`"\\\\"`)}},
	{` "`, []token{tERROR("unclosed string literal")}},
	{" \"\n", []token{tERROR("unclosed string literal")}},
	{`"foo" "bar"`, []token{tSTRING(`"foo"`), tSTRING(`"bar"`)}},

	// names
	{"$foo", []token{tNAME("$foo")}},

	{`$foo "bar"`, []token{tNAME("$foo"), tSTRING(`"bar"`)}},

	// numbers
	{"0123 123 -123 +123", []token{
		tNUMBER("0123"),
		tNUMBER("123"),
		tNUMBER("-123"),
		tNUMBER("+123"),
	}},
	{"0xaBc -0XaBc +0xaBc", []token{
		tNUMBER("0xaBc"),
		tNUMBER("-0XaBc"),
		tNUMBER("+0xaBc"),
	}},
	{"0. 0.123 -0.123 +0.123", []token{
		tNUMBER("0."),
		tNUMBER("0.123"),
		tNUMBER("-0.123"),
		tNUMBER("+0.123"),
	}},
	{"1.23e10 -1.23E-10 +1.23e+10 +1e+10 +1.e+10", []token{
		tNUMBER("1.23e10"),
		tNUMBER("-1.23E-10"),
		tNUMBER("+1.23e+10"),
		tNUMBER("+1e+10"),
		tNUMBER("+1.e+10"),
	}},
	{"0xabc.def 0xabc.defE2 0xabc.defe2", []token{
		tNUMBER("0xabc.def"),
		tNUMBER("0xabc.defE2"),
		tNUMBER("0xabc.defe2"),
	}},
	// FIXME
	//{"0xabc.defe-2 0xabc.defp+2", []token{
	//        tNUMBER("0xabc.defe-2"),
	//        tNUMBER("0xabc.defp+2"),
	//}},
	//{"inf -inf +inf infinity -infinity +infinity", []token{
	//        tNUMBER("inf"),
	//        tNUMBER("-inf"),
	//        tNUMBER("+inf"),
	//        tNUMBER("infinity"),
	//        tNUMBER("-infinity"),
	//        tNUMBER("+infinity"),
	//}},
	//{"nan nan:0xaBc", []token{tNUMBER("nan"), tNUMBER("nan:0xaBc")}},

	// atoms
	{"i32 anyfunc add rotl call_indirect", []token{
		tok(I32, "i32"),
		tok(ANYFUNC, "anyfunc"),
		tok(ADD, "add"),
		tok(ROTL, "rotl"),
		tok(CALL_INDIRECT, "call_indirect"),
	}},
	{"offset=0x03 align=8 trunc_s i64.extend_s/i32", []token{
		tok(OFFSET, "offset"),
		tok(EQUAL, "="),
		tNUMBER("0x03"),

		tok(ALIGN, "align"),
		tok(EQUAL, "="),
		tNUMBER("8"),

		tok(TRUNC, "trunc"),
		tok(UNDERSCORE, "_"),
		tok(S, "s"),

		tok(I64, "i64"),
		tok(DOT, "."),
		tok(EXTEND, "extend"),
		tok(UNDERSCORE, "_"),
		tok(S, "s"),
		tok(SLASH, "/"),
		tok(I32, "i32"),
	}},
}

func TestLexer(t *testing.T) {
	for _, tt := range lexertests {
		l := newLexer(bytes.NewReader([]byte(tt.in)))
		got, err := l.lex()
		if err != nil {
			t.Fatalf("%s: %v", tt.in, err)
		}
		if !equal(got, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.in, got, tt.want)
		}
	}
}

func equal(a, b []token) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].typ != b[i].typ || !bytes.Equal(a[i].text, b[i].text) {
			return false
		}
	}
	return true
}

func tok(typ tokenType, text string) token { return token{typ: typ, text: []byte(text)} }

func tNAME(s string) token   { return token{typ: NAME, text: []byte(s)} }
func tSTRING(s string) token { return token{typ: STRING, text: []byte(s)} }
func tNUMBER(s string) token { return token{typ: NUMBER, text: []byte(s)} }

func tERROR(format string, args ...interface{}) token {
	return token{typ: ERROR, text: []byte(fmt.Sprintf(format, args...))}
}
