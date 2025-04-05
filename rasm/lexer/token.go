package lexer

// Position is the representation of line and column positioning in a source file.
type Position struct {
	Line uint
	Col  uint
}

// NewPosition returns a new position with the given parameters.
func NewPosition(line, col uint) Position {
	return Position{
		Line: line,
		Col:  col,
	}
}

// A TokenID represents a token's identifier.
type TokenID uint

const (
	Instruction = iota // represents an instruction
	Register           // reqpresents an register
	Section            // represents the section keyword
	Comma              // represents the character ','
	Colon              // represents the character ':'
	Newline            // represents a newline
	Identifier         // represents an identifier/name
	Hex                // represents a hexadecimal number
	Octal              // represents an octal number
	Decimal            // represents a decimal number
)

// Token represents the output of the [Lexer], containing information
// about the lexed input.
type Token interface {
	// Pos returns the [Position] saved in the token.
	Pos() Position

	// ID returns the [TokenID] of the token.
	ID() TokenID

	// SubID returns a sub-identifier of the token. Used for tokens of types like
	// [Instruction] or [Register], to identify the specific intruction or register.
	SubID() uint32

	// Raw returns the raw string of the token lexed by the lexer.
	Raw() string
}

type token struct {
	// pos contains the position of the token in the file.
	pos Position

	// id contains the token's identifier
	id TokenID

	// subID contains the tokens sub-identifier, like specific register or
	// instruction identifiers.
	subID uint32

	// raw contains the string lexed by the lexer.
	raw string
}

// NewToken creates a new token with the sub-identifier set to 0.
func NewToken(pos Position, id TokenID, raw string) Token {
	return &token{
		pos:   pos,
		id:    id,
		subID: 0,
		raw:   raw,
	}
}

// NewTokenSubID creates a new token.
func NewTokenSubID(pos Position, id TokenID, subID uint32, raw string) Token {
	return &token{
		pos:   pos,
		id:    id,
		subID: subID,
		raw:   raw,
	}
}

func (t token) Pos() Position {
	return t.pos
}

func (t token) ID() TokenID {
	return t.id
}

func (t token) SubID() uint32 {
	return t.subID
}

func (t token) Raw() string {
	return t.raw
}
