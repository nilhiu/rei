package x86

import (
	"encoding/binary"
	"errors"
)

// This might not be the best decision, but for now it's okay.
type opcodeBase struct {
	for8Bit []byte
	forRest []byte
}

func (o opcodeBase) getBySize(sz uint) []byte {
	switch sz {
	case 8:
		return o.for8Bit
	default:
		return o.forRest
	}
}

type translateFunc func([]Operand) ([]byte, error)

type opcodeFormat struct {
	operands   [][]OpType
	translates []translateFunc
}

var instrToFormat = map[Mnemonic]opcodeFormat{
	Mov: {
		[][]OpType{
			{OpRegister, OpRegister},
			{OpRegister, OpImmediate},
		},
		[]translateFunc{
			gRR(opcodeBase{[]byte{0x88}, []byte{0x89}}, true),
			gRI(opcodeBase{[]byte{0xB0}, []byte{0xB8}}, ^uint(0)),
		},
	},
}

func gRR(base opcodeBase, mustSameSize bool) func([]Operand) ([]byte, error) {
	return func(ops []Operand) ([]byte, error) {
		return genericRegReg(base, mustSameSize, ops[0].(Register), ops[1].(Register))
	}
}

// if doesn't have class give value of `^byte(0)`
func gRI(base opcodeBase, class uint) func([]Operand) ([]byte, error) {
	return func(ops []Operand) ([]byte, error) {
		return genericRegImm(base, class, ops[0].(Register), ops[1].(Immediate))
	}
}

func genericRegImm(
	base opcodeBase,
	class uint,
	reg Register,
	imm Immediate,
) ([]byte, error) {
	immBytes, err := translateImmToRegNative(imm.Value(), reg)
	if err != nil {
		return nil, err
	}

	return append(genericReg(base, reg, class), immBytes...), nil
}

func genericRegReg(
	base opcodeBase,
	mustSameSize bool,
	reg1 Register,
	reg2 Register,
) ([]byte, error) {
	if mustSameSize && reg1.Size() != reg2.Size() {
		return nil, errors.New("given registers must be the same size")
	}

	opcode := genericRegNoPrefix(base, reg1, uint(reg2))
	if (reg1.IsRex() || reg2.IsRex()) && (reg1.IsRexExcluded() || reg2.IsRexExcluded()) {
		return nil, errors.New("given register cannot be encoded with a REX prefix")
	}

	return append(prefixRR(reg1, reg2), opcode...), nil
}

func genericReg(base opcodeBase, reg Register, class uint) []byte {
	prefix := prefixR(reg)
	return append(prefix, genericRegNoPrefix(base, reg, class)...)
}

func genericRegNoPrefix(base opcodeBase, reg Register, class uint) []byte {
	opcode := base.getBySize(reg.Size())

	if class != ^uint(0) {
		return append(opcode, encodeModRM(0b11, Register(class), reg))
	} else {
		opcode[len(opcode)-1] += reg.EncodeByte()
		return opcode
	}
}

func prefixRR(reg1 Register, reg2 Register) []byte {
	prefix := []byte{}
	if reg1.Size() == 16 || reg2.Size() == 16 {
		prefix = []byte{0x66}
	}
	if reg1.IsRex() || reg2.IsRex() {
		prefix = append(prefix, encodeRexRR(reg1, reg2))
	}
	return prefix
}

func prefixR(reg Register) []byte {
	prefix := []byte{}
	if reg.Size() == 16 {
		prefix = []byte{0x66}
	}
	if reg.IsRex() {
		prefix = append(prefix, encodeRexR(reg))
	}
	return prefix
}

func translateImmToRegNative(imm uint, reg Register) ([]byte, error) {
	if imm > uint(1)<<reg.Size() {
		return nil, errors.New("immediate too big")
	}

	switch reg.Size() {
	case 8:
		return []byte{byte(imm)}, nil
	case 16:
		return binary.LittleEndian.AppendUint16([]byte{}, uint16(imm)), nil
	case 32:
		return binary.LittleEndian.AppendUint32([]byte{}, uint32(imm)), nil
	case 64:
		return binary.LittleEndian.AppendUint64([]byte{}, uint64(imm)), nil
	}

	return nil, errors.New("unreachable")
}

func encodeModRM(mod byte, reg Register, mem Register) byte {
	switch mod {
	case 0b11:
		return (mod << 6) | (reg.EncodeByte() << 3) | mem.EncodeByte()
	default:
		panic("any 'mod' value other than 0b11 is unsupported for now")
	}
}

func encodeRexR(reg Register) byte {
	return encodeRexRR(reg, Register(0))
}

func encodeRexRR(reg1 Register, reg2 Register) byte {
	var rex byte = 0x40
	if reg1.IsRexB() {
		rex |= 0x01
	}
	if reg2.IsRexB() {
		rex |= 0x04
	}
	if reg1.Size() == 64 {
		rex |= 0x08
	}
	return rex
}
