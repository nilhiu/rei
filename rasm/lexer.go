package rasm

import (
	"bufio"
	"io"
	"strings"
	"unicode"

	"github.com/nilhiu/rei/x86"
)

type Position struct {
	Line uint
	Col  uint
}

type TokenId uint

const (
	Eof TokenId = iota
	Illegal
	Instruction
	Register
	Section
	Comma
	Colon
	Newline

	Identifier
	Hex
	Octal
	Decimal
)

// The `id` field contains the above `TokenId` constants in the first 5 bits,
// and in the cases of `Instruction` and `Register` the upper 27 (or 59 if 64-bit)
// bits contains the instruction/register identifiers.
type Token struct {
	pos Position
	id  TokenId
	raw string
}

func (t *Token) Pos() Position {
	return t.pos
}

func (t *Token) Id() TokenId {
	return t.id & 0x1f
}

func (t *Token) SpecId() uint {
	return (uint(t.id) >> 5) << 5
}

func (t *Token) Raw() string {
	return t.raw
}

type Lexer struct {
	rd  *bufio.Reader
	pos Position
}

func NewLexer(rd io.Reader) *Lexer {
	return &Lexer{
		rd:  bufio.NewReader(rd),
		pos: Position{Line: 1, Col: 0},
	}
}

func (l *Lexer) Next() Token {
	for {
		r, _, err := l.rd.ReadRune()
		if err != nil {
			if err == io.EOF {
				return Token{pos: l.pos, id: Eof, raw: ""}
			}
			panic(err)
		}

		pos := l.pos
		l.pos.Col++
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

func (l *Lexer) lexZero() Token {
	r, _, err := l.rd.ReadRune()
	if err != nil {
		if err == io.EOF {
			return Token{pos: Position{Line: l.pos.Line, Col: l.pos.Col - 1}, id: Decimal, raw: "0"}
		}
		panic(err)
	}

	l.pos.Col++
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
	var raw string
	for {
		r, _, err := l.rd.ReadRune()
		if err != nil {
			if err == io.EOF {
				return Token{pos: pos, id: Hex, raw: raw}
			}
			panic(err)
		}

		l.pos.Col++
		switch r {
		case 'A', 'B', 'C', 'D', 'E', 'F':
			raw = raw + string(r)
		default:
			if unicode.IsDigit(r) {
				raw = raw + string(r)
			} else {
				l.unread()
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
	var raw string
	for {
		r, _, err := l.rd.ReadRune()
		if err != nil {
			if err == io.EOF {
				return Token{pos: pos, id: Octal, raw: raw}
			}
			panic(err)
		}

		l.pos.Col++
		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			raw = raw + string(r)
		default:
			l.unread()
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
	var raw string
	for {
		r, _, err := l.rd.ReadRune()
		if err != nil {
			if err == io.EOF {
				return Token{pos: pos, id: Decimal, raw: raw}
			}
			panic(err)
		}

		l.pos.Col++
		if unicode.IsDigit(r) {
			raw = raw + string(r)
		} else {
			l.unread()
			return Token{pos: pos, id: Decimal, raw: raw}
		}
	}
}

func (l *Lexer) lexIdentifier() Token {
	pos := l.pos
	var raw string
	for {
		r, _, err := l.rd.ReadRune()
		if err != nil {
			if err == io.EOF {
				return Token{pos: pos, id: identTokenId(raw), raw: raw}
			}
			panic(err)
		}

		l.pos.Col++
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' {
			raw = raw + string(r)
		} else {
			l.unread()
			return Token{pos: pos, id: identTokenId(raw), raw: raw}
		}
	}
}

func identTokenId(id string) TokenId {
	ident := strings.ToLower(id)
	switch ident {
	case "section":
		return Section
	default:
		if instr := x86.MnemonicSearchMap[ident]; instr != 0 {
			return TokenId(uint(Instruction) | uint(instr))
		} else if reg := x86.RegisterSearchMap[ident]; reg != 0 {
			return TokenId(uint(Register) | uint(reg))
		} else {
			return Identifier
		}
	}
}
