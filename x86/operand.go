package x86

type OpType uint

const (
	OpImmediate OpType = iota
	OpRegister
)

type Operand interface {
	Type() OpType
	Value() uint
}

func (r Register) Type() OpType {
	return OpRegister
}

func (r Register) Value() uint {
	return uint(r.EncodeByte())
}

type Immediate uint

func (imm Immediate) Type() OpType {
	return OpImmediate
}

func (imm Immediate) Value() uint {
	return uint(imm)
}
