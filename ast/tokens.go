package ast

import "fmt"

// token or text string returned from the lexer.
type token struct {
	typ  tokenType
	text []byte
}

func (t token) String() string {
	return fmt.Sprintf("%s(%s)", t.typ, string(t.text))
}

// isZero reports whether t == token{}.
// (Structs that contain a slice are not comparable.)
func (t token) isZero() bool {
	return t.typ == 0 && t.text == nil
}

func (t token) isVar() bool {
	return t.typ == NUMBER || t.typ == NAME
}

//go:generate stringer -type=tokenType
type tokenType int

const (
	ERROR tokenType = iota

	DOT
	EQUAL
	LPAREN
	RPAREN
	SLASH
	UNDERSCORE

	NAME
	NUMBER // value
	STRING

	beginType
	F32
	F64
	I32
	I64
	endType

	beginElemType
	ANYFUNC
	endElemType

	beginUnOp
	CLZ
	CTZ
	EQZ
	POPCNT
	endUnOp

	beginBinOp
	ADD
	AND
	DIV
	MUL
	OR
	REM
	ROTL
	ROTR
	SHL
	SHR
	SUB
	XOR
	endBinOp

	beginRelOp
	EQ
	GE
	GT
	LE
	LT
	NE
	endRelOp

	beginSign
	S
	U
	endSign

	beginCvtOp
	CONVERT
	DEMOTE
	EXTEND
	PROMOTE
	REINTERPRET
	TRUNC
	endCvtOp

	ALIGN
	OFFSET

	beginInstr
	BLOCK
	IF
	LOOP
	endInstr

	ELSE
	END
	THEN

	MUT

	beginOp
	BR_IF
	BR_TABLE
	CALL
	CALL_INDIRECT
	CONST
	CURRENT_MEMORY
	DROP
	GET_GLOBAL
	GET_LOCAL
	GROW_MEMORY
	LOAD
	NOP
	RETURN
	SELECT
	SET_GLOBAL
	SET_LOCAL
	STORE
	TEE_LOCAL
	UNREACHABLE
	endOp

	DATA
	ELEM
	EXPORT
	FUNC
	GLOBAL
	IMPORT
	LOCAL
	MEMORY
	MODULE
	PARAM
	RESULT
	START
	TABLE
	TYPE
)

var atom = map[string]tokenType{
	"i32": I32,
	"i64": I64,
	"f32": F32,
	"f64": F64,

	"anyfunc": ANYFUNC,

	"clz":    CLZ,
	"ctz":    CTZ,
	"eqz":    EQZ,
	"popcnt": POPCNT,

	"add":  ADD,
	"and":  AND,
	"div":  DIV,
	"mul":  MUL,
	"or":   OR,
	"rem":  REM,
	"rotl": ROTL,
	"rotr": ROTR,
	"shl":  SHL,
	"shr":  SHR,
	"sub":  SUB,
	"xor":  XOR,

	"eq": EQ,
	"ge": GE,
	"gt": GT,
	"le": LE,
	"lt": LT,
	"ne": NE,

	"convert":     CONVERT,
	"demote":      DEMOTE,
	"extend":      EXTEND,
	"promote":     PROMOTE,
	"reinterpret": REINTERPRET,
	"trunc":       TRUNC,

	"align":  ALIGN,
	"mut":    MUT,
	"offset": OFFSET,

	"block": BLOCK,
	"else":  ELSE,
	"end":   END,
	"if":    IF,
	"loop":  LOOP,
	"then":  THEN,

	"br_if":          BR_IF,
	"br_table":       BR_TABLE,
	"call":           CALL,
	"call_indirect":  CALL_INDIRECT,
	"const":          CONST,
	"current_memory": CURRENT_MEMORY,
	"drop":           DROP,
	"get_global":     GET_GLOBAL,
	"get_local":      GET_LOCAL,
	"grow_memory":    GROW_MEMORY,
	"load":           LOAD,
	"nop":            NOP,
	"return":         RETURN,
	"select":         SELECT,
	"set_global":     SET_GLOBAL,
	"set_local":      SET_LOCAL,
	"store":          STORE,
	"tee_local":      TEE_LOCAL,
	"unreachable":    UNREACHABLE,

	"data":   DATA,
	"elem":   ELEM,
	"export": EXPORT,
	"func":   FUNC,
	"global": GLOBAL,
	"import": IMPORT,
	"local":  LOCAL,
	"memory": MEMORY,
	"module": MODULE,
	"param":  PARAM,
	"result": RESULT,
	"start":  START,
	"table":  TABLE,
	"type":   TYPE,
}
