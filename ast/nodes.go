package ast

type Module struct {
	Name  string
	Types []*TypeDef
	Funcs []*Func
}

type TypeDef struct {
	Name string
	Func *FuncSig
}

type Func struct {
	Name      string
	Signature *FuncSig
	Locals    []*Local

	Export *EmbeddedExport
	// or
	Import *EmbeddedImport
}

type Instruction struct {
}

type EmbeddedExport struct {
	Name string
}

type EmbeddedImport struct {
	Module string
	Name   string
}

type Local struct {
	Name string    // may be zero
	Type tokenType // of F32, F64, I32, I64
}

type FuncSig struct {
	Type *FuncSigType
	// or
	Params  []*Param    // may be empty
	Results []tokenType // of F32, F64, I32, I64 (may be empty)
}

type FuncSigType struct {
	Var *Variable
}

type Param struct {
	Name  string      // may be zero if len(Types) != 1
	Types []tokenType // of F32, F64, I32, I64
}

type Variable struct {
	// one of
	Index int
	Name  string
}
