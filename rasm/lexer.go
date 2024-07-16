package rasm

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

type Position struct {
	Line uint
	Col  uint
}

type TokenId uint

const (
	Eof = iota
	Illegal
	Instruction
	Register
	Section
	Label
	Comma

	Name
	Hex
	Octal
	Decimal
)

// The `id` field contains the above `TokenId` constants in the first 5 bits,
// and in the cases of `Instruction` and `Register` the upper 27 (or 59 if 64-bit)
// bits contains the instruction/register identifiers.
type Token struct {
	Pos Position
	Id  TokenId
	Raw string
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
				return Token{Pos: l.pos, Id: Eof, Raw: ""}
			}
			panic(err)
		}

		pos := l.pos
		l.pos.Col++
		switch r {
		case ',':
			return Token{Pos: pos, Id: Comma, Raw: ","}
		case '0':
			return l.lexZero()
		case '\n':
			l.pos.Line++
			l.pos.Col = 0
			continue
		default:
			if unicode.IsSpace(r) {
				continue
			} else if unicode.IsDigit(r) {
				l.unread()
				return l.lexDecimal()
			} else if unicode.IsLetter(r) || r == '_' || r == '.' {
				l.unread()
				return l.lexName()
			}

			return Token{Pos: pos, Id: Illegal, Raw: string(r)}
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
			return Token{Pos: Position{Line: l.pos.Line, Col: l.pos.Col - 1}, Id: Decimal, Raw: "0"}
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
			tok.Raw = string('0') + tok.Raw
			tok.Pos.Col--
			return tok
		} else {
			l.unread()
			return Token{Pos: Position{Line: l.pos.Line, Col: l.pos.Col - 1}, Id: Decimal, Raw: "0"}
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
				return Token{Pos: pos, Id: Hex, Raw: raw}
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
					return Token{Pos: pos, Id: Illegal, Raw: "hex prefix without logical continuation"}
				}
				return Token{Pos: pos, Id: Hex, Raw: raw}
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
				return Token{Pos: pos, Id: Octal, Raw: raw}
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
				return Token{Pos: pos, Id: Illegal, Raw: "octal prefix without logical continuation"}
			}
			return Token{Pos: pos, Id: Octal, Raw: raw}
		}
	}
}

func (l *Lexer) lexDecimal() Token {
	pos := l.pos
	var raw string
	for {
		r, _, err := l.rd.ReadRune()
		if err != nil {
			panic(err)
		}

		l.pos.Col++
		if unicode.IsDigit(r) {
			raw = raw + string(r)
		} else {
			l.unread()
			return Token{Pos: pos, Id: Decimal, Raw: raw}
		}
	}
}

func (l *Lexer) lexName() Token {
	pos := l.pos
	var raw string
	for {
		r, _, err := l.rd.ReadRune()
		if err != nil {
			if err == io.EOF {
				return Token{Pos: pos, Id: nameTokenId(raw), Raw: raw}
			}
			panic(err)
		}

		l.pos.Col++
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' {
			raw = raw + string(r)
		} else if r == ':' {
			// Keywords can also be labels. May change later.
			return Token{Pos: pos, Id: Label, Raw: raw}
		} else {
			l.unread()
			return Token{Pos: pos, Id: nameTokenId(raw), Raw: raw}
		}
	}
}

func nameTokenId(nm string) TokenId {
	name := strings.ToLower(nm)
	switch name {
	case "section":
		return Section
	default:
		if instr := x86InstrSearchMap[name]; instr != 0 {
			return TokenId(Instruction | instr)
		} else if reg := x86RegisterSearchMap[name]; reg != 0 {
			return TokenId(Register | reg)
		} else {
			return Name
		}
	}
}
