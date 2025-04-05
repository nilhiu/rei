package x86

import "github.com/nilhiu/rei/rasm/lexer"

type IDMap struct{}

func (m *IDMap) Find(ident string) (*lexer.IdentToken, bool) {
	if r, ok := RegisterSearchMap[ident]; ok {
		return lexer.NewIdentToken(lexer.Register, uint32(r)), true
	} else if i, ok := MnemonicSearchMap[ident]; ok {
		return lexer.NewIdentToken(lexer.Instruction, uint32(i)), true
	}

	return nil, false
}
