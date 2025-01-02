package x86

type Mnemonic uint

// Mnemonic constants (Very WIP)
const (
	_            = iota
	ADD Mnemonic = iota << 5
	MOV
)

var MnemonicSearchMap = map[string]Mnemonic{
	"add": ADD,
	"mov": MOV,
}
