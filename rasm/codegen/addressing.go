package codegen

type SIBAddressing struct {
	Scale        uint32
	Index        uint32
	Base         uint32
	Displacement int64
}

func (a SIBAddressing) Type() OperandType {
	return OpTypeSIBAddressing
}

func (a SIBAddressing) Value() any {
	return &a
}

// TODO: Improved named addressing, this won't work when offsetting
// (ex. [label+4])
type NamedAddressing struct {
	Ident string
}

func (a NamedAddressing) Type() OperandType {
	return OpTypeNamedAddressing
}

func (a NamedAddressing) Value() any {
	return a.Ident
}
