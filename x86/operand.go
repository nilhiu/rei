package x86

// OpType represents the operand type.
type OpType uint

const (
	OpImmediate OpType = iota // the operand in an immediate/constant
	OpRegister                // the operand is a register
	OpAddress                 // the operand is an address
)

// A Operand is a interface, which operands have to implement.
type Operand interface {
	Type() OpType // returns the type of the operand
	Value() uint  // returns the value of the operand
}

func (r Register) Type() OpType {
	return OpRegister
}

func (r Register) Value() uint {
	return uint(r.EncodeByte())
}

// A Immediate represents an immediate, or a constant, value.
type Immediate uint

func (imm Immediate) Type() OpType {
	return OpImmediate
}

func (imm Immediate) Value() uint {
	return uint(imm)
}

func (a Address) Type() OpType {
	return OpAddress
}

func (a Address) Value() uint {
	return uint(a.EncodeSib())
}
