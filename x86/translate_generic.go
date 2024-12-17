package x86

import (
	"encoding/binary"
	"errors"
)

func pIf(pred func(ops []Operand) bool, then translateFunc, otherwise translateFunc) translateFunc {
	return func(ops []Operand) ([]byte, error) {
		if pred(ops) {
			return then(ops)
		}
		return otherwise(ops)
	}
}

func gRR(base []byte, mustSameSize bool) func([]Operand) ([]byte, error) {
	return func(ops []Operand) ([]byte, error) {
		return genericRegReg(base, mustSameSize, ops[0].(Register), ops[1].(Register))
	}
}

func gRI(base []byte, class byte, immFmt immFmt) func([]Operand) ([]byte, error) {
	return func(ops []Operand) ([]byte, error) {
		return genericRegImm(base, class, immFmt, ops[0].(Register), ops[1].(Immediate))
	}
}

func cRI(base []byte, immFmt immFmt) func([]Operand) ([]byte, error) {
	return func(ops []Operand) ([]byte, error) {
		return compressedRegImm(base, immFmt, ops[0].(Register), ops[1].(Immediate))
	}
}

func genericRegImm(
	base []byte,
	class byte,
	immFmt immFmt,
	reg Register,
	imm Immediate,
) ([]byte, error) {
	immBytes, err := translateImmByFmt(imm.Value(), reg, immFmt)
	if err != nil {
		return nil, err
	}

	return append(genericReg(base, reg, class), immBytes...), nil
}

func genericRegReg(
	base []byte,
	mustSameSize bool,
	reg1 Register,
	reg2 Register,
) ([]byte, error) {
	if mustSameSize && reg1.Size() != reg2.Size() {
		return nil, errors.New("given registers must be the same size")
	}

	opcode := genericRegNoPrefix(base, reg1, reg2.EncodeByte(), 0b11)
	if (reg1.IsRex() || reg2.IsRex()) && (reg1.IsRexExcluded() || reg2.IsRexExcluded()) {
		return nil, errors.New("given register cannot be encoded with a REX prefix")
	}

	return append(prefixRR(reg1, reg2), opcode...), nil
}

func genericReg(base []byte, reg Register, class byte) []byte {
	prefix := prefixR(reg)
	return append(prefix, genericRegNoPrefix(base, reg, class, 0b11)...)
}

func genericRegNoPrefix(base []byte, reg Register, class byte, mod byte) []byte {
	opcode := base
	if reg.Size() != 8 && class&opFmtClassNotChange == 0 {
		opcode = base
		if class == opFmtClassCompactReg {
			opcode[len(opcode)-1] += 8
		} else {
			opcode[len(opcode)-1]++
		}
	}

	if class == opFmtClassCompactReg {
		opcode[len(opcode)-1] += reg.EncodeByte()
		return opcode
	} else {
		return append(opcode, encodeModRM(mod, class&0b111, reg.EncodeByte()))
	}
}

func compressedRegImm(
	base []byte,
	immFmt immFmt,
	reg Register,
	imm Immediate,
) ([]byte, error) {
	immBytes, err := translateImmByFmt(imm.Value(), reg, immFmt)
	if err != nil {
		return nil, err
	}

	opcode := base
	if reg.Size() != 8 {
		opcode[len(opcode)-1]++
	}

	return append(append(prefixR(reg), opcode...), immBytes...), nil
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

func translateImmByFmt(imm uint, reg Register, immFmt immFmt) ([]byte, error) {
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
	return (mod << 6) | (reg << 3) | mem
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
