package x86

import (
	"encoding/binary"
	"errors"
)

// This might not be the best decision, but for now it's okay.
type OpcodeEncoding struct {
	For8Bit  byte
	For16Bit byte
	For32Bit byte
	For64Bit byte
}

func (o OpcodeEncoding) getForReg(reg Register) byte {
	switch reg.Size() {
	case 8:
		return o.For8Bit
	case 16:
		return o.For16Bit
	case 32:
		return o.For32Bit
	case 64:
		return o.For64Bit
	}
	panic("unreachable")
}

func Translate(mnem Mnemonic, ops ...Operand) ([]byte, error) {
	switch mnem {
	case Mov:
		return translateMov(ops)
	}

	return nil, errors.New("unknown mnemonic encountered")
}

func translateMov(ops []Operand) ([]byte, error) {
	if len(ops) != 2 {
		return nil, errors.New("the 'mov' mnemonic must only have 2 operands")
	}

	if ops[0].Type() == OpRegister && ops[1].Type() == OpImmediate {
		return translateMovRegImm(ops[0].(Register), ops[1].Value())
	} else if ops[0].Type() == OpRegister && ops[1].Type() == OpRegister {
		return translateMovRegReg(ops[0].(Register), ops[1].(Register))
	}

	return nil, errors.New("given operands are unsupported by the 'mov' mnemonic")
}

func translateMovRegImm(reg Register, imm uint) ([]byte, error) {
	return translateGenericRegImm(OpcodeEncoding{0xB0, 0xB8, 0xB8, 0xB8}, reg, imm, false, 0)
}

func translateMovRegReg(dst Register, src Register) ([]byte, error) {
	return translateGenericRegReg(OpcodeEncoding{0x88, 0x89, 0x89, 0x89}, dst, src)
}

func translateGenericRegImm(opEnc OpcodeEncoding, reg Register, imm uint, isModRM bool, regDigit byte) ([]byte, error) {
	immBytes, err := translateImmToRegNative(imm, reg)
	if err != nil {
		return nil, err
	}
	opcode := opEnc.getForReg(reg)
	return append(encodeOpcodeRegImm(opcode, reg, isModRM, regDigit), immBytes...), nil
}

func translateGenericRegReg(opEnc OpcodeEncoding, dst Register, src Register) ([]byte, error) {
	if dst.Size() != src.Size() {
		return nil, errors.New("different sized registers in a register-register instruction")
	}
	opcode := opEnc.getForReg(dst)
	return encodeOpcodeRegReg(opcode, dst, src)
}

func encodeOpcodeRegImm(opcode byte, reg Register, isModRM bool, regDigit byte) []byte {
	opBytes := []byte{}
	if reg.Size() == 16 {
		opBytes = []byte{0x66}
	}
	if reg.IsRex() {
		opBytes = append(opBytes, encodeRegRex(reg))
	}
	if isModRM {
		return append(opBytes, opcode, encodeModRM(0b11, Register(regDigit), reg))
	} else {
		return append(opBytes, opcode+reg.EncodeByte())
	}
}

func encodeOpcodeRegReg(opcode byte, reg Register, regOpt Register) ([]byte, error) {
	opBytes := []byte{}
	if reg.Size() == 16 {
		opBytes = []byte{0x66}
	}
	if reg.IsRex() || regOpt.IsRex() {
		if reg.IsRexExcluded() || regOpt.IsRexExcluded() {
			return nil, errors.New("given register cannot be encoded with REX byte present")
		}
		opBytes = append(opBytes, encodeRegRegRex(reg, regOpt))
	}
	return append(opBytes, opcode, encodeModRM(0b11, reg, regOpt)), nil
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

func encodeModRM(mod byte, dst Register, src Register) byte {
	switch mod {
	case 0b11:
		return (mod << 6) | (src.EncodeByte() << 3) | dst.EncodeByte()
	default:
		panic("any 'mod' value other than 0b11 is unsupported for now")
	}
}

func encodeRegRex(reg Register) byte {
	return encodeRegRegRex(reg, Register(0))
}

func encodeRegRegRex(reg1 Register, reg2 Register) byte {
	var rex byte = 0b0100_0000
	if reg1.IsRexB() {
		rex |= 0b0000_0001
	}
	if reg2.IsRexB() {
		rex |= 0b0000_0100
	}
	if reg1.Size() == 64 {
		rex |= 0b0000_1000
	}
	return rex
}
