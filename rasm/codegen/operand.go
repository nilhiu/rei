package codegen

import (
	"strconv"

	"github.com/nilhiu/rei/rasm/lexer"
)

// OperandType represents all the types of operands recognized by [CodeGen].
type OperandType uint

const (
	OpTypeImmediate       OperandType = iota // represents an immediate
	OpTypeRegister                           // represents a register
	OpTypeSIBAddressing                      // represents an scale-index-base and displacement addressing mode
	OpTypeNamedAddressing                    // represents an named/label addressing mode
)

// Operant represents the generic structure of all operands. For immediates,
// the `Value` field represents the unsigned representation of the lexed number,
// while for `subID`ed tokens (registers, instructions) it will contain the said subID.
type Operand interface {
	Type() OperandType
	Value() any
}

type operand struct {
	typ   OperandType
	value any
}

func (o *operand) Type() OperandType {
	return o.typ
}

func (o *operand) Value() any {
	return o.value
}

// NewOperand creates a Operand from the given token. If the token is unable
// to be turned into an operand, it will return `nil`.
func NewOperand(tok lexer.Token) Operand {
	switch tok.ID() {
	case lexer.Decimal:
		return newImmediateOperand(tok.Raw())
	case lexer.Register:
		return &operand{
			typ:   OpTypeRegister,
			value: tok.SubID(),
		}
	}

	return nil
}

func newImmediateOperand(str string) Operand {
	if str[0] == '-' {
		val, err := strconv.ParseUint(str[1:], 10, 64)
		if err != nil {
			return nil
		}

		val |= 1 << 63
		return &operand{
			typ:   OpTypeImmediate,
			value: val,
		}
	}

	val, err := strconv.ParseUint(str, 10, 64)
	if err != nil {
		return nil
	}

	return &operand{
		typ:   OpTypeImmediate,
		value: val,
	}
}
