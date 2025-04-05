package x86

import "github.com/nilhiu/rei/rasm/codegen"

// A Immediate represents an immediate, or a constant, value.
type Immediate uint64

func (imm Immediate) Type() codegen.OperandType {
	return codegen.OpTypeImmediate
}

func (imm Immediate) Value() any {
	return uint64(imm)
}
