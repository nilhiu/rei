package x86

import "encoding/binary"

type Address struct {
	Scale        byte
	Index        Register
	Base         Register
	Displacement uint32
}

func (a Address) EncodeSib() byte {
	var scale byte = 0
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
	} else {
		return 0b10
	}
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
	return a.Scale == 0 && a.Index == Register(0) && a.Base == Register(0) &&
		a.Displacement == 0
}

func (a Address) isSIB() bool {
	return (a.Scale != 1) || (a.Index != Register(0))
}
