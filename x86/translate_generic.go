package x86

import (
	"encoding/binary"
	"errors"

	"github.com/nilhiu/rei/rasm/codegen"
)

func pIf(
	pred func(ops []codegen.Operand) bool,
	then translateFunc,
	otherwise translateFunc,
) translateFunc {
	return func(ops []codegen.Operand) ([]byte, error) {
		if pred(ops) {
			return then(ops)
		}

		return otherwise(ops)
	}
}

func gRR(base []byte, mustSameSize bool) func([]codegen.Operand) ([]byte, error) {
	return func(ops []codegen.Operand) ([]byte, error) {
		return genericRegReg(
			base,
			mustSameSize,
			Register(ops[0].Value().(uint32)),
			Register(ops[1].Value().(uint32)),
		)
	}
}

func gRI(base []byte, class byte, immFmt immFmt) func([]codegen.Operand) ([]byte, error) {
	return func(ops []codegen.Operand) ([]byte, error) {
		return genericRegImm(
			base,
			class,
			immFmt,
			Register(ops[0].Value().(uint32)),
			Immediate(ops[1].Value().(uint64)),
		)
	}
}

// TODO: addresses not yet implemented in CodeGen.
func gRA(base []byte) func([]codegen.Operand) ([]byte, error) {
	return func(ops []codegen.Operand) ([]byte, error) {
		return genericRegSIBAddr(
			base,
			Register(ops[0].Value().(uint32)),
			ops[1].Value().(*codegen.SIBAddressing),
		)
	}
}

func cRI(base []byte, immFmt immFmt) func([]codegen.Operand) ([]byte, error) {
	return func(ops []codegen.Operand) ([]byte, error) {
		return compressedRegImm(
			base,
			immFmt,
			Register(ops[0].Value().(uint32)),
			Immediate(ops[1].Value().(uint64)),
		)
	}
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

	opcode := genericRegNoPrefix(base, reg1, reg2.Code(), 0b11)

	if (reg1.isRexRequired() || reg2.isRexRequired()) &&
		(reg1.isRexExcluded() || reg2.isRexExcluded()) {
		return nil, errors.New("given register cannot be encoded with a REX prefix")
	}

	return append(prefixRR(reg1, reg2), opcode...), nil
}

func genericRegImm(
	base []byte,
	class byte,
	immFmt immFmt,
	reg Register,
	imm Immediate,
) ([]byte, error) {
	immBytes, err := translateImmByFmt(imm.Value().(uint64), reg, immFmt)
	if err != nil {
		return nil, err
	}

	return append(genericReg(base, reg, class), immBytes...), nil
}

func genericRegSIBAddr(
	base []byte,
	reg Register,
	addr *codegen.SIBAddressing,
) ([]byte, error) {
	// TODO: check @addr.size == reg.size
	// if reg.isRexRequired() && reg.isRexExcluded() {
	// 	return nil, errors.New("given register cannot be encoded with a REX prefix")
	// }

	opcode := genericRegNoPrefix(base, Register(addr.Base), reg.Code(), sibMod(addr))
	if addr.Scale > 1 || addr.Index != 0 {
		// Set ModR/M byte's R/M field to 4 (0b100) as SIB is to be encoded.
		opcode[len(opcode)-1] = (opcode[len(opcode)-1] & 0b11111000) | 0b100

		sib, err := encodeSIB(addr)
		if err != nil {
			return nil, err
		}

		opcode = append(opcode, sib)
	}

	if addr.Displacement != 0 {
		if addr.Displacement <= 0x7f {
			opcode = append(opcode, byte(addr.Displacement))
		} else {
			opcode = binary.LittleEndian.AppendUint32(opcode, uint32(addr.Displacement))
		}
	}

	return append(prefixR(reg), opcode...), nil
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
		opcode[len(opcode)-1] += reg.Code()

		return opcode
	}

	return append(opcode, encodeModRM(mod, class&0b111, reg.Code()))
}

func compressedRegImm(
	base []byte,
	immFmt immFmt,
	reg Register,
	imm Immediate,
) ([]byte, error) {
	immBytes, err := translateImmByFmt(imm.Value().(uint64), reg, immFmt)
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

	if reg1.isRexRequired() || reg2.isRexRequired() {
		prefix = append(prefix, encodeRexRR(reg1, reg2))
	}

	return prefix
}

func prefixR(reg Register) []byte {
	prefix := []byte{}

	if reg.Size() == 16 {
		prefix = []byte{0x66}
	}

	if reg.isRexRequired() {
		prefix = append(prefix, encodeRexR(reg))
	}

	return prefix
}

func translateImmByFmt(imm uint64, reg Register, immFmt immFmt) ([]byte, error) {
	sz := immFmt.getBySize(reg.Size())

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
	return encodeRexRR(reg, NilReg)
}

func encodeRexRR(reg1 Register, reg2 Register) byte {
	var rex byte = 0x40

	if reg1.isRexBRequired() {
		rex |= 0x01
	}

	if reg2.isRexBRequired() {
		rex |= 0x04
	}

	if reg1.Size() == 64 {
		rex |= 0x08
	}

	return rex
}

func encodeSIB(addr *codegen.SIBAddressing) (byte, error) {
	var scale byte
	switch addr.Scale {
	case 1:
		scale = 0b00
	case 2:
		scale = 0b01
	case 4:
		scale = 0b10
	case 8:
		scale = 0b11
	default:
		return 0, errors.New("encodeSIB(): invalid addressing scale")
	}

	return scale<<6 | Register(addr.Index).Code()<<3 | Register(addr.Base).Code(), nil
}

func sibMod(addr *codegen.SIBAddressing) byte {
	if addr.Displacement == 0 {
		return 0
	} else if addr.Displacement <= 0x7f {
		return 0b01
	}

	return 0b10
}
