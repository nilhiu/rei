package codegen

// Translator is the interface that has to be implemented by separate ISA
// packages to be used with the [CodeGen] to generate machine code.
type Translator interface {
	// Translate returns machine code for the given instruction with the
	// provided operands. It can return any type of error, which will later
	// be wrapped by [CodeGen] with [Error].
	Translate(instrID uint32, ops ...Operand) ([]byte, error)
}
