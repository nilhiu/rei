package x86

import "encoding/binary"

// An Address represents a SIB + displacement encoding of a memory address.
type Address struct {
	Scale        byte
	Index        Register
	Base         Register
	Displacement uint32
}

// EncodeSIB encodes the [Address] as an SIB byte.
func (a Address) EncodeSIB() byte {
	if a.isNil() {
		return 0x25
	}

	var scale byte

	switch a.Scale {
	case 1:
		scale = 0b00
	case 2:
		scale = 0b01
	case 4:
		scale = 0b10
	case 8:
		scale = 0b11
	}

	return (scale << 6) | (a.Index.EncodeByte() << 3) | a.Base.EncodeByte()
}

func (a Address) mod() byte {
	if a.Displacement == 0 {
		return 0b00
	} else if a.Displacement <= 0x7F {
		return 0b01
	}

	return 0b10
}

func (a Address) disp() []byte {
	if a.Displacement <= 0x7F {
		return []byte{byte(a.Displacement)}
	}

	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, a.Displacement)

	return b
}

func (a Address) isNil() bool {
	return a.Scale == 0 && a.Index == NilReg && a.Base == NilReg &&
		a.Displacement == 0
}

func (a Address) isSIB() bool {
	return (a.Scale != 1) || (a.Index != NilReg)
}
