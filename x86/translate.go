package x86

import (
	"errors"
	"slices"
)

func Translate(mnem Mnemonic, ops ...Operand) ([]byte, error) {
	fmt := mnemToFmt(mnem)
	if fmt == nil {
		return nil, errors.New("unknown mnemonic encountered")
	}
	var opTypes []OpType
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
