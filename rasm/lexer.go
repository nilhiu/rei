package rasm

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

type Position struct {
	line uint
	col  uint
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
	Hex   // TODO
	Octal // TODO
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

func (t *Token) Pos() Position { return t.pos }
func (t *Token) Id() TokenId   { return t.id }
func (t *Token) Raw() string   { return t.raw }

type Lexer struct {
	rd  *bufio.Reader
	pos Position
}

func NewLexer(rd io.Reader) *Lexer {
	return &Lexer{
		rd:  bufio.NewReader(rd),
		pos: Position{line: 1, col: 0},
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

		l.pos.col++
		switch r {
		case ',':
			return Token{pos: l.pos, id: Comma, raw: ","}
		default:
			if unicode.IsSpace(r) {
				continue
			} else if unicode.IsDigit(r) {
				l.unread()
				return l.lexDecimal()
			} else if unicode.IsLetter(r) {
				l.unread()
				return l.lexName()
			}

			return Token{pos: l.pos, id: Illegal, raw: string(r)}
		}
	}
}

func (l *Lexer) unread() {
	if err := l.rd.UnreadRune(); err != nil {
		panic(err)
	}
	l.pos.col--
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

		if unicode.IsDigit(r) {
			raw = raw + string(r)
		} else {
			l.unread()
			return Token{pos: pos, id: Decimal, raw: raw}
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
				return Token{pos: pos, id: nameTokenId(raw), raw: raw}
			}
			panic(err)
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' {
			raw = raw + string(r)
		} else if r == ':' {
			// Keywords can also be labels. May change later.
			return Token{pos: pos, id: Label, raw: raw}
		} else {
			l.unread()
			return Token{pos: pos, id: nameTokenId(raw), raw: raw}
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
