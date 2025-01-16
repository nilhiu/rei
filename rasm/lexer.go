// The rasm package implements types and methods to provide lexing, parsing,
// and code generation for assembly written in rei's syntax.
//
// As of now, these types are dependent on eachother, as they work concurrently
// to assemble a assembly source. This dependence may later be removed for a
// more "modular" assemblying, such as being able to programatically generate
// code without using the lexer.
package rasm

import (
	"bufio"
	"io"
	"strings"
	"unicode"

	"github.com/nilhiu/rei/x86"
)

// Position is the representation of line and column positioning
// in a source file.
type Position struct {
	Line uint
	Col  uint
}

// A TokenID is the type of a [Token].
type TokenID uint

const (
	EOF         TokenID = iota // represents the end of file
	Illegal                    // represents an illegal/unknown character
	Instruction                // represents an instruction
	Register                   // reqpresents an register
	Section                    // represents the section keyword
	Comma                      // represents the character ','
	Colon                      // represents the character ':'
	Newline                    // represents a newline

	Identifier // represents an identifier/name
	Hex        // represents a hexadecimal number
	Octal      // represents an octal number
	Decimal    // represents a decimal number
)

// Token represents the output of the [Lexer], containing information
// about the lexed input.
type Token struct {
	pos Position
	// id contains the above `TokenID` constants in the first 5 bits,
	// and in the cases of `Instruction` and `Register` the rest contains the
	// instruction/register identifiers.
	id TokenID
	// raw contains the string lexed by the lexer.
	raw string
}

// NewToken creates a new token based on the given parameters.
// Possibly will be removed as it seems quite unnecessary.
func NewToken(pos Position, id TokenID, raw string) Token {
	return Token{pos, id, raw}
}

// Pos returns the [Position] saved in the token.
func (t *Token) Pos() Position {
	return t.pos
}

// ID returns the [TokenID] of the token.
func (t *Token) ID() TokenID {
	return t.id & 0x1f
}

// SpecID returns a "special" ID of the token. Should only be used for [Token]'s of
// type [Instruction] or [Register], otherwise it will, and should, always
// return zero.
func (t *Token) SpecID() uint {
	return (uint(t.id) >> 5) << 5
}

// Raw returns the raw string of the token lexed by the lexer.
func (t *Token) Raw() string {
	return t.raw
}

// A Lexer is object which turns the source file into tokens, which are
// used by the [Parser].
type Lexer struct {
	rd  *bufio.Reader
	pos Position
	sb  strings.Builder
}

// NewLexer create a new [Lexer] based on the [io.Reader] given to it.
func NewLexer(rd io.Reader) *Lexer {
	return &Lexer{
		rd:  bufio.NewReader(rd),
		pos: Position{Line: 1, Col: 0},
		sb:  strings.Builder{},
	}
}

// Next lexes and returns the next token in the source file. If the file
// has been fully lexed, Next will always return a token with the [EOF]
// [TokenID].
func (l *Lexer) Next() Token {
	for {
		pos := l.pos

		r, isEOF := l.read()
		if isEOF {
			return Token{pos: pos, id: EOF, raw: ""}
		}

		switch r {
		case ',':
			return Token{pos: pos, id: Comma, raw: ","}
		case ':':
			return Token{pos: pos, id: Colon, raw: ":"}
		case '0':
			return l.lexZero()
		case '\n':
			l.pos.Line++
			l.pos.Col = 0

			return Token{pos: pos, id: Newline, raw: "\\n"}
		default:
			if unicode.IsSpace(r) {
				continue
			} else if unicode.IsDigit(r) {
				l.unread()

				return l.lexDecimal()
			} else if unicode.IsLetter(r) || r == '_' || r == '.' {
				l.unread()

				return l.lexIdentifier()
			}

			return Token{pos: pos, id: Illegal, raw: string(r)}
		}
	}
}

func (l *Lexer) unread() {
	if err := l.rd.UnreadRune(); err != nil {
		panic(err)
	}

	l.pos.Col--
}

func (l *Lexer) read() (rune, bool) {
	l.pos.Col++

	r, _, err := l.rd.ReadRune()
	if err != nil {
		if err == io.EOF {
			return 0, true
		}

		panic(err)
	}

	return r, false
}

func (l *Lexer) writeStr(r rune) {
	_, err := l.sb.WriteRune(r)
	if err != nil {
		panic(err)
	}
}

func (l *Lexer) popStr() string {
	str := l.sb.String()
	l.sb.Reset()

	return str
}

func (l *Lexer) lexZero() Token {
	r, isEOF := l.read()
	if isEOF {
		return Token{pos: Position{Line: l.pos.Line, Col: l.pos.Col - 2}, id: Decimal, raw: "0"}
	}

	switch r {
	case 'x', 'X':
		return l.lexHex()
	case 'o', 'O':
		return l.lexOctal()
	default:
		if unicode.IsDigit(r) {
			l.unread()
			tok := l.lexDecimal()
			tok.raw = string('0') + tok.raw
			tok.pos.Col--

			return tok
		} else {
			l.unread()

			return Token{pos: Position{Line: l.pos.Line, Col: l.pos.Col - 1}, id: Decimal, raw: "0"}
		}
	}
}

func (l *Lexer) lexHex() Token {
	pos := l.pos
	pos.Col -= 2

	for {
		r, isEOF := l.read()
		if isEOF {
			return Token{pos: pos, id: Illegal, raw: "hex prefix without logical continuation"}
		}

		switch r {
		case 'A', 'B', 'C', 'D', 'E', 'F', 'a', 'b', 'c', 'd', 'e', 'f':
			l.writeStr(r)
		default:
			if unicode.IsDigit(r) {
				l.writeStr(r)
			} else {
				l.unread()

				raw := l.popStr()
				if raw == "" {
					return Token{pos: pos, id: Illegal, raw: "hex prefix without logical continuation"}
				}

				return Token{pos: pos, id: Hex, raw: raw}
			}
		}
	}
}

func (l *Lexer) lexOctal() Token {
	pos := l.pos
	pos.Col -= 2

	for {
		r, isEOF := l.read()
		if isEOF {
			return Token{
				pos: pos,
				id:  Illegal,
				raw: "octal prefix without logical continuation",
			}
		}

		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			l.writeStr(r)
		default:
			l.unread()

			raw := l.popStr()
			if raw == "" {
				return Token{
					pos: pos,
					id:  Illegal,
					raw: "octal prefix without logical continuation",
				}
			}

			return Token{pos: pos, id: Octal, raw: raw}
		}
	}
}

func (l *Lexer) lexDecimal() Token {
	pos := l.pos

	for {
		r, isEOF := l.read()
		if isEOF {
			return Token{pos: pos, id: Decimal, raw: l.popStr()}
		}

		if unicode.IsDigit(r) {
			l.writeStr(r)
		} else {
			l.unread()
			return Token{pos: pos, id: Decimal, raw: l.popStr()}
		}
	}
}

func (l *Lexer) lexIdentifier() Token {
	pos := l.pos

	for {
		r, isEOF := l.read()
		if isEOF {
			raw := l.popStr()
			return Token{pos: pos, id: identTokenID(raw), raw: raw}
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' {
			l.writeStr(r)
		} else {
			l.unread()
			raw := l.popStr()
			return Token{pos: pos, id: identTokenID(raw), raw: raw}
		}
	}
}

func identTokenID(id string) TokenID {
	ident := strings.ToLower(id)
	switch ident {
	case "section":
		return Section
	default:
		if instr := x86.MnemonicSearchMap[ident]; instr != 0 {
			return Instruction | TokenID(instr)
		} else if reg := x86.RegisterSearchMap[ident]; reg != 0 {
			return Register | TokenID(reg)
		} else {
			return Identifier
		}
	}
}
