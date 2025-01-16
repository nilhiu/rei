// The x86 package implements trnaslation from user-readable assembly to
// x86 instructions.
package x86

import (
	"errors"
	"slices"
)

// Translate translates the provided mnemonic and operands into x86 machine
// code. An error can occur if the given mnemonic is unknown, or if the given
// operands don't match what the mnemonic should be given.
func Translate(mnem Mnemonic, ops ...Operand) ([]byte, error) {
	fmt := mnemToFmt(mnem)
	if fmt == nil {
		return nil, errors.New("unknown mnemonic encountered")
	}

	opTypes := []OpType{}
	for _, op := range ops {
		opTypes = append(opTypes, op.Type())
	}

	ix := ^uint(0)

	for i, typ := range fmt.operands {
		if slices.Compare(typ, opTypes) == 0 {
			ix = uint(i)

			break
		}
	}

	if ix == ^uint(0) {
		return nil, errors.New("given operands for this mnemonic are unsupported")
	}

	return fmt.translates[ix](ops)
}

func mnemToFmt(mnem Mnemonic) *opFmt {
	switch mnem {
	case ADD:
		return newOpFmt().
			withClass(0).
			addRI([]byte{0x80}, immFmtNative32).
			withARegCompressed([]byte{0x04}, immFmtNative32).
			withByteCompressed([]byte{0x83}).
			addRR([]byte{0x00}, true)
	case MOV:
		return newOpFmt().
			withClass(opFmtClassCompactReg).
			addRI([]byte{0xB0}, immFmtNative).
			addRR([]byte{0x88}, true).
			addRA([]byte{0x8A})
	}

	return nil
}
