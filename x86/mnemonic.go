package x86

// A Mnemonic is a x86 assembly mnemonic/instruction
type Mnemonic uint

const (
	_            = iota
	ADD Mnemonic = iota << 5
	MOV
)

// MnemonicSearchMap maps the string representation of mnemonics to their
// [Mnemonic] counterparts.
var MnemonicSearchMap = map[string]Mnemonic{
	"add": ADD,
	"mov": MOV,
}
