package codegen

// Code represents the machine code that [CodeGen] has generated.
type Code interface {
	// Section returns the section where the code was generated.
	Section() string

	// Bytes returns the machine code generated.
	Bytes() []byte
}

type code struct {
	section string
	bytes   []byte
}

// NewCode creates a new [Code] with the given parameters.
func NewCode(section string, bytes []byte) Code {
	return &code{
		section: section,
		bytes:   bytes,
	}
}

func (c *code) Section() string {
	return c.section
}

func (c *code) Bytes() []byte {
	return c.bytes
}
