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

type immediateFormat struct {
	for8Bit  byte
	for16Bit byte
	for32Bit byte
	for64Bit byte
}

func (i immediateFormat) getBySize(sz uint) byte {
	switch sz {
	case 8:
		return i.for8Bit
	case 16:
		return i.for16Bit
	case 32:
		return i.for32Bit
	case 64:
		return i.for64Bit
	}

	panic("shouldn't be here")
}

var mnemonicToFormat = map[Mnemonic]opcodeFormat{
	Add: {
		[][]OpType{
			{OpRegister, OpImmediate},
			{OpRegister, OpRegister},
		},
		[]translateFunc{
			pIf(
				func(ops []Operand) bool { return ops[0].(Register).Size() != 8 && ops[1].Value() <= 0x7F },
				gRI(opcodeBase{[]byte{0x83}, []byte{0x83}}, 0, immediateFormat{8, 8, 8, 8}),
				pIf(
					func(ops []Operand) bool { return ops[0].(Register).isARegister() },
					cRI(opcodeBase{[]byte{0x04}, []byte{0x05}}, immediateFormat{8, 16, 32, 32}),
					gRI(opcodeBase{[]byte{0x80}, []byte{0x81}}, 0, immediateFormat{8, 16, 32, 32}),
				),
			),
			gRR(opcodeBase{[]byte{0x00}, []byte{0x01}}, true),
		},
	},
	Mov: {
		[][]OpType{
			{OpRegister, OpRegister},
			{OpRegister, OpImmediate},
		},
		[]translateFunc{
			gRR(opcodeBase{[]byte{0x88}, []byte{0x89}}, true),
			gRI(opcodeBase{[]byte{0xB0}, []byte{0xB8}}, ^byte(0), immediateFormat{8, 16, 32, 64}),
		},
	},
}

func pIf(pred func(ops []Operand) bool, then translateFunc, otherwise translateFunc) translateFunc {
	return func(ops []Operand) ([]byte, error) {
		if pred(ops) {
			return then(ops)
		}
		return otherwise(ops)
	}
}

func gRR(base opcodeBase, mustSameSize bool) func([]Operand) ([]byte, error) {
	return func(ops []Operand) ([]byte, error) {
		return genericRegReg(base, mustSameSize, ops[0].(Register), ops[1].(Register))
	}
}

// if doesn't have class give value of `^byte(0)`
func gRI(base opcodeBase, class byte, immFmt immediateFormat) func([]Operand) ([]byte, error) {
	return func(ops []Operand) ([]byte, error) {
		return genericRegImm(base, class, immFmt, ops[0].(Register), ops[1].(Immediate))
	}
}

func cRI(base opcodeBase, immFmt immediateFormat) func([]Operand) ([]byte, error) {
	return func(ops []Operand) ([]byte, error) {
		return compressedRegImm(base, immFmt, ops[0].(Register), ops[1].(Immediate))
	}
}

func genericRegImm(
	base opcodeBase,
	class byte,
	immFmt immediateFormat,
	reg Register,
	imm Immediate,
) ([]byte, error) {
	immBytes, err := translateImmByFormat(imm.Value(), reg, immFmt)
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

	opcode := genericRegNoPrefix(base, reg1, reg2.EncodeByte())
	if (reg1.IsRex() || reg2.IsRex()) && (reg1.IsRexExcluded() || reg2.IsRexExcluded()) {
		return nil, errors.New("given register cannot be encoded with a REX prefix")
	}

	return append(prefixRR(reg1, reg2), opcode...), nil
}

func genericReg(base opcodeBase, reg Register, class byte) []byte {
	prefix := prefixR(reg)
	return append(prefix, genericRegNoPrefix(base, reg, class)...)
}

func genericRegNoPrefix(base opcodeBase, reg Register, class byte) []byte {
	opcode := base.getBySize(reg.Size())

	if class != ^byte(0) {
		return append(opcode, encodeModRM(0b11, class, reg.EncodeByte()))
	} else {
		opcode[len(opcode)-1] += reg.EncodeByte()
		return opcode
	}
}

func compressedRegImm(
	base opcodeBase,
	immFmt immediateFormat,
	reg Register,
	imm Immediate,
) ([]byte, error) {
	immBytes, err := translateImmByFormat(imm.Value(), reg, immFmt)
	if err != nil {
		return nil, err
	}

	return append(append(prefixR(reg), base.getBySize(reg.Size())...), immBytes...), nil
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

func translateImmByFormat(imm uint, reg Register, immFmt immediateFormat) ([]byte, error) {
	sz := immFmt.getBySize(reg.Size())

	if imm > uint(1)<<sz {
		return nil, errors.New("immediate too big")
	}

	switch sz {
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

func encodeModRM(mod byte, reg byte, mem byte) byte {
	switch mod {
	case 0b11:
		return (mod << 6) | (reg << 3) | mem
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
