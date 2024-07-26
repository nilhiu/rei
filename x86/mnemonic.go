package x86

type Mnemonic uint

// Mnemonic constants (Very WIP)
const (
	_            = iota
	Mov Mnemonic = iota << 5
)

var MnemonicSearchMap = map[string]Mnemonic{
	"mov": Mov,
}
