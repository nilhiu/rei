package x86

import (
	"encoding/binary"
	"errors"
)

type OpType uint

const (
	OpImmediate OpType = iota
	OpRegister
)

type Operand struct {
	Type OpType
	Data uint
}

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

	if ops[0].Type == OpRegister && ops[1].Type == OpImmediate {
		return translateMovRegImm(Register(ops[0].Data), ops[1].Data)
	} else if ops[0].Type == OpRegister && ops[1].Type == OpRegister {
		return translateMovRegReg(Register(ops[0].Data), Register(ops[1].Data))
	}

	return []byte{}, errors.New("given operands are unsupported by the 'mov' mnemonic")
}

func translateMovRegImm(reg Register, imm uint) ([]byte, error) {
	switch reg {
	case Al, Cl, Dl, Bl, Ah, Ch, Dh, Bh:
		return []byte{}, errors.New("8-bit registers not yet supported")
	case Ax, Cx, Dx, Bx:
		code := []byte{0x66, 0xB8 + encodeRegister(reg)}
		return binary.LittleEndian.AppendUint16(code, uint16(imm)), nil
	case Eax, Ecx, Edx, Ebx:
		code := []byte{0xB8 + encodeRegister(reg)}
		return binary.LittleEndian.AppendUint32(code, uint32(imm)), nil
	}
	return []byte{}, errors.New("unsupported register")
}

func translateMovRegReg(dst Register, src Register) ([]byte, error) {
	if registerSize(dst) != registerSize(src) {
		return []byte{}, errors.New("different sized registers in a register-register instruction")
	}

	switch registerSize(dst) {
	case 16:
		return []byte{0x66, 0x89, encodeModRM(0b11, dst, src)}, nil
	case 32:
		return []byte{0x89, encodeModRM(0b11, dst, src)}, nil
	}

	return []byte{}, errors.New("unsupported register size")
}

func encodeRegister(reg Register) byte {
	switch reg {
	case Al, Ax, Eax, Rax, R8b, R8w, R8d, R8:
		return 0
	case Cl, Cx, Ecx, Rcx, R9b, R9w, R9d, R9:
		return 1
	case Dl, Dx, Edx, Rdx, R10b, R10w, R10d, R10:
		return 2
	case Bl, Bx, Ebx, Rbx, R11b, R11w, R11d, R11:
		return 3
	case Ah, Sp, Esp, Spl, Rsp, R12b, R12w, R12d, R12:
		return 4
	case Ch, Bp, Ebp, Bpl, Rbp, R13b, R13w, R13d, R13:
		return 5
	case Dh, Si, Esi, Sil, Rsi, R14b, R14w, R14d, R14:
		return 6
	case Bh, Di, Edi, Dil, Rdi, R15b, R15w, R15d, R15:
		return 7
	default:
		panic("given register is unsupported ")
	}
}

// TODO: Add REX byte extension to ModR/M.
func encodeModRM(mod byte, dst Register, src Register) byte {
	switch mod {
	case 0b11:
		return (mod << 6) | (encodeRegister(src) << 3) | encodeRegister(dst)
	default:
		panic("any 'mod' value other than 0b11 is unsupported for now")
	}
}

func registerSize(reg Register) uint {
	switch reg {
	case Al, Cl, Dl, Bl, Sil, Dil, Spl, Bpl, R8b, R9b, R10b, R11b, R12b, R13b, R14b, R15b, Ah, Ch, Dh, Bh:
		return 8
	case Ax, Cx, Dx, Bx, Si, Di, Sp, Bp, R8w, R9w, R10w, R11w, R12w, R13w, R14w, R15w:
		return 16
	case Eax, Ecx, Edx, Ebx, Esi, Edi, Esp, Ebp, R8d, R9d, R10d, R11d, R12d, R13d, R14d, R15d:
		return 32
	case Rax, Rcx, Rdx, Rbx, Rsi, Rdi, Rsp, Rbp, R8, R9, R10, R11, R12, R13, R14, R15:
		return 64
	}

	panic("unreachable")
}
