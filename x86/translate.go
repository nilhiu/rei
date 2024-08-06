package x86

import (
	"errors"
)

func Translate(mnem Mnemonic, ops ...Operand) ([]byte, error) {
	switch mnem {
	case Mov:
		return translateGenericMnemonicOp2(
			OpcodeEncoding{0xB0, 0xB8, 0xB8, 0xB8},
			OpcodeEncoding{0x88, 0x89, 0x89, 0x89},
			ops,
			false,
			0,
		)
	}

	return nil, errors.New("unknown mnemonic encountered")
}
