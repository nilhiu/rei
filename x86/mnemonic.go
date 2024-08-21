package x86

type Mnemonic uint

// Mnemonic constants (Very WIP)
const (
	_            = iota
	Add Mnemonic = iota << 5
	Mov
)

var MnemonicSearchMap = map[string]Mnemonic{
	"add": Add,
	"mov": Mov,
}
