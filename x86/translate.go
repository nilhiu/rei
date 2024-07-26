package x86

import (
	"encoding/binary"
	"errors"
)

func Translate(mnem Mnemonic, ops ...Operand) ([]byte, error) {
	switch mnem {
	case Mov:
		return translateMov(ops)
	}

	return []byte{}, errors.New("unknown mnemonic encountered")
}

// TODO: Add translation for 8-bit and 64-bit registers.
func translateMov(ops []Operand) ([]byte, error) {
	if len(ops) != 2 {
		return []byte{}, errors.New("the 'mov' mnemonic must only have 2 operands")
	}

	if ops[0].Type() == OpRegister && ops[1].Type() == OpImmediate {
		return translateMovRegImm(ops[0].(Register), ops[1].Value())
	} else if ops[0].Type() == OpRegister && ops[1].Type() == OpRegister {
		return translateMovRegReg(ops[0].(Register), ops[1].(Register))
	}

	return []byte{}, errors.New("given operands are unsupported by the 'mov' mnemonic")
}

func translateMovRegImm(reg Register, imm uint) ([]byte, error) {
	switch reg {
	case Al, Cl, Dl, Bl, Ah, Ch, Dh, Bh:
		return []byte{}, errors.New("8-bit registers not yet supported")
	case Ax, Cx, Dx, Bx:
		code := []byte{0x66, 0xB8 + reg.EncodeByte()}
		return binary.LittleEndian.AppendUint16(code, uint16(imm)), nil
	case Eax, Ecx, Edx, Ebx:
		code := []byte{0xB8 + reg.EncodeByte()}
		return binary.LittleEndian.AppendUint32(code, uint32(imm)), nil
	}
	return []byte{}, errors.New("unsupported register")
}

func translateMovRegReg(dst Register, src Register) ([]byte, error) {
	if dst.Size() != src.Size() {
		return []byte{}, errors.New("different sized registers in a register-register instruction")
	}

	switch dst.Size() {
	case 16:
		return []byte{0x66, 0x89, encodeModRM(0b11, dst, src)}, nil
	case 32:
		return []byte{0x89, encodeModRM(0b11, dst, src)}, nil
	}

	return []byte{}, errors.New("unsupported register size")
}

// TODO: Add REX byte extension to ModR/M.
func encodeModRM(mod byte, dst Register, src Register) byte {
	switch mod {
	case 0b11:
		return (mod << 6) | (src.EncodeByte() << 3) | dst.EncodeByte()
	default:
		panic("any 'mod' value other than 0b11 is unsupported for now")
	}
}
