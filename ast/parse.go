package ast

import (
	"bytes"
	"fmt"
	"strconv"
)

type parser struct {
	buf []token
	pos int
}

func newParser(tokens []token) *parser {
	return &parser{buf: tokens}
}

func (p *parser) parse() *Module {
	return p.parseModule()
}

// parseModule parses a module:
// 	( module <name>? <typedef>* <func>* <import>* <export>* <table>? <memory>? <global>* <elem>* <data>* <start>? )
func (p *parser) parseModule() *Module {
	m := new(Module)
	p.expect(LPAREN)
	p.expect(MODULE)
	p.maybeName(&m.Name)
	for {
		switch {
		case p.match(LPAREN, TYPE):
			m.Types = append(m.Types, p.parseTypeDef())
		case p.match(LPAREN, FUNC):
			m.Funcs = append(m.Funcs, p.parseFunc())
		case p.peek().typ == RPAREN:
			return m
		default:
			panic(fmt.Sprintf("malformed module: %s", p.peek()))
		}
	}
}

// parseTypeDef parses a typedef:
// 	( type <name>? ( func <funcsig> ) )
//
// '(' 'type' has been read.
func (p *parser) parseTypeDef() *TypeDef {
	def := new(TypeDef)
	p.maybeName(&def.Name)
	p.expect(LPAREN)
	p.expect(FUNC)
	def.Func = p.parseFuncSig()
	return def
}

// parseFunc parses a func (excluding sugar):
// 	( func <name>? <func_sig> <local>* <instr>* )
// 	( func <name>? ( export <string> ) <func_sig> <local>* <instr>* ) ;; = (export <string> (func <N>) (func <name>? <func_sig> <local>* <instr>*)
// 	( func <name>? ( import <string> <string> ) <func_sig>)           ;; = (import <name>? <string> <string> (func <func_sig>))
//
// '(' 'func' has been read.
func (p *parser) parseFunc() *Func {
	fn := new(Func)
	p.maybeName(&fn.Name)
	switch {
	case p.match(LPAREN, EXPORT):
		name, _ := strconv.Unquote(string(p.expect(STRING).text))
		fn.Export = &EmbeddedExport{Name: name}
		p.expect(RPAREN)
	case p.match(LPAREN, IMPORT):
		module, _ := strconv.Unquote(string(p.expect(STRING).text))
		name, _ := strconv.Unquote(string(p.expect(STRING).text))
		fn.Import = &EmbeddedImport{Module: module, Name: name}
		p.expect(RPAREN)
	}
	fn.Signature = p.parseFuncSig()
	if fn.Import != nil {
		return fn
	}
	fn.Locals = p.parseLocalList()
	p.expect(RPAREN)
	return fn
}

// parseInstruction parses an instr.
func (p *parser) parseInstruction() *Instruction {
	return nil
}

// parseLocalList parses a list of locals.
// 	local: ( local <type>* ) | ( local <name> <type> )
func (p *parser) parseLocalList() []*Local {
	var locals []*Local
	for p.match(LPAREN, LOCAL) {
		if name, hasName := p.accept(NAME); hasName {
			return []*Local{{
				Name: extractName(name),
				Type: p.exceptIsType().typ,
			}}
		}
		var locals []*Local
		for {
			t, isTyp := p.acceptIsType()
			if !isTyp {
				break
			}
			locals = append(locals, &Local{Type: t.typ})
		}
	}
	return locals
}

// parseFuncSig parses a func_sig:
// 	( type <var> ) | <param>* <result>*
// 	param: ( param <type>* ) | ( param <name> <type> )
// 	result: ( result <type> )
func (p *parser) parseFuncSig() *FuncSig {
	switch {
	case p.match(LPAREN, TYPE):
		v := p.parseVariable()
		return &FuncSig{Type: &FuncSigType{Var: v}}
	case p.match(LPAREN, PARAM), p.match(LPAREN, RESULT):
		sig := new(FuncSig)
		p.unreadN(2)
		sig.Params = p.parseParamList()
		sig.Results = p.parseResultList()
		return sig
	default:
		return new(FuncSig)
	}
}

// parseParamList parses a list of params.
func (p *parser) parseParamList() []*Param {
	var params []*Param
	for p.match(LPAREN, PARAM) {
		params = append(params, p.parseParam())
	}
	return params
}

// parseParam parses a param.
// 	( param <type>* ) | ( param <name> <type> )
func (p *parser) parseParam() *Param {
	p.expect(LPAREN)
	p.expect(PARAM)
	if name, hasName := p.accept(NAME); hasName {
		return &Param{
			Name:  extractName(name),
			Types: []tokenType{p.exceptIsType().typ},
		}
	}
	param := new(Param)
	for {
		t, isTyp := p.acceptIsType()
		if !isTyp {
			break
		}
		param.Types = append(param.Types, t.typ)
	}
	p.expect(RPAREN)
	return param
}

// parseResultList parses a list of results.
// 	result: ( result <type> )
func (p *parser) parseResultList() []tokenType {
	var res []tokenType
	for p.match(LPAREN, RESULT) {
		res = append(res, p.exceptIsType().typ)
	}
	return res
}

func (p *parser) parseVariable() *Variable {
	v := p.expect(NAME, NUMBER)
	if v.typ == NAME {
		return &Variable{Name: extractName(v)}
	}
	return &Variable{Index: extractInteger(v)}
}

func (p *parser) maybeName(field *string) {
	if tok, isName := p.accept(NAME); isName {
		*field = extractName(tok)
	}
}

func extractName(tok token) string {
	if tok.typ != NAME {
		panic(tok)
	}
	return string(bytes.TrimPrefix(tok.text, []byte("$")))
}

func extractInteger(tok token) int {
	if tok.typ != NUMBER {
		panic(tok)
	}
	n, err := strconv.Atoi(string(tok.text))
	if err != nil {
		panic(err)
	}
	return n
}

// read returns the next token.
// On EOF, it returns the zero value.
func (p *parser) read() (t token) {
	if p.pos == len(p.buf) {
		p.pos++
		return token{}
	}
	t = p.buf[p.pos]
	p.pos++
	return
}

// peek returns the next token without advancing the reader.
// On EOF, it returns the zero value.
func (p *parser) peek() token {
	if p.pos == len(p.buf) {
		return token{}
	}
	return p.buf[p.pos+1]
}

func (p *parser) unread() {
	if p.pos == 0 {
		panic("unread at position 0")
	}
	p.pos--
}

func (p *parser) unreadN(n int) {
	if p.pos-n < 0 {
		panic("unread at position 0")
	}
	p.pos -= n
}

// accept consumes the next token if it is in the valid set.
// TODO: disallow empty argument
func (p *parser) accept(v tokenType, alid ...tokenType) (t token, isValid bool) {
	valid := append([]tokenType{v}, alid...)
	tok := p.read()
	for _, typ := range valid {
		if tok.typ == typ {
			return tok, true
		}
	}
	p.unread()
	return token{}, false
}

func (p *parser) acceptIsType() (token, bool) { return p.accept(F32, F64, I32, I64) }

func (p *parser) expect(v tokenType, alid ...tokenType) token {
	valid := append([]tokenType{v}, alid...)
	tok := p.read()
	for _, typ := range valid {
		if tok.typ == typ {
			return tok
		}
	}
	panic(fmt.Sprintf("expected one of %s, found %s", valid, tok))
}

func (p *parser) exceptIsType() token { return p.expect(F32, F64, I32, I64) }

func (p *parser) match(h tokenType, t ...tokenType) bool {
	tokens := append([]tokenType{h}, t...)
	for i, t := range tokens {
		if p.read().typ != t {
			for j := 0; j <= i; j++ {
				p.unread()
			}
			return false
		}
	}
	return true
}
