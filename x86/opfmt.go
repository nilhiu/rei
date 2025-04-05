package x86

import "github.com/nilhiu/rei/rasm/codegen"

type translateFunc func([]codegen.Operand) ([]byte, error)

type opFmt struct {
	operands   [][]codegen.OperandType
	translates []translateFunc
	class      byte
}

type immFmt struct {
	forByte  byte
	forWord  byte
	forDWord byte
	forQWord byte
}

const (
	opFmtClassNotChange  = byte(1 << 7)
	opFmtClassCompactReg = byte(1 << 6)
)

var (
	immFmtNative   = immFmt{8, 16, 32, 64}
	immFmtNative32 = immFmt{8, 16, 32, 32}
	immFmtByte     = immFmt{8, 8, 8, 8}
)

func newOpFmt() *opFmt {
	return new(opFmt)
}

func (o *opFmt) withClass(class byte) *opFmt {
	o.class = class

	return o
}

func (o *opFmt) addRI(base []byte, immFmt immFmt) *opFmt {
	o.operands = append(
		o.operands,
		[]codegen.OperandType{codegen.OpTypeRegister, codegen.OpTypeImmediate},
	)
	o.translates = append(o.translates, gRI(base, o.class, immFmt))

	return o
}

func (o *opFmt) addRR(base []byte, mustSameSize bool) *opFmt {
	o.operands = append(
		o.operands,
		[]codegen.OperandType{codegen.OpTypeRegister, codegen.OpTypeRegister},
	)
	o.translates = append(o.translates, gRR(base, mustSameSize))

	return o
}

func (o *opFmt) addRA(base []byte) *opFmt {
	o.operands = append(
		o.operands,
		[]codegen.OperandType{codegen.OpTypeRegister, codegen.OpTypeSIBAddressing},
	)
	o.translates = append(o.translates, gRA(base))

	return o
}

func (o *opFmt) withARegCompressed(base []byte, immFmt immFmt) *opFmt {
	o.translates[len(o.translates)-1] = pIf(
		func(ops []codegen.Operand) bool {
			return Register(ops[0].Value().(uint32)).isARegister()
		},
		cRI(base, immFmt),
		o.translates[len(o.translates)-1],
	)

	return o
}

func (o *opFmt) withByteCompressed(base []byte) *opFmt {
	o.translates[len(o.translates)-1] = pIf(
		func(ops []codegen.Operand) bool {
			return Register(ops[0].Value().(uint32)).Size() != 8 && ops[1].Value().(uint64) <= 0x7F
		},
		gRI(base, o.class|opFmtClassNotChange, immFmtByte),
		o.translates[len(o.translates)-1],
	)

	return o
}

func (i immFmt) getBySize(sz uint) byte {
	switch sz {
	case 8:
		return i.forByte
	case 16:
		return i.forWord
	case 32:
		return i.forDWord
	case 64:
		return i.forQWord
	}

	panic("shouldn't be here")
}
