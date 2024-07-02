package rasm

import (
	"bufio"
	"io"
	"unicode"
)

type Position struct {
	line uint
	col  uint
}

type TokenId uint

const (
	EOF = iota
	ILLEGAL
	SECTION
	LABEL
	COMMA

	NAME
	HEX   // TODO
	OCTAL // TODO
	DECIMAL
)

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
				return Token{pos: l.pos, id: EOF, raw: ""}
			}
			panic(err)
		}

		l.pos.col++
		switch r {
		case ',':
			return Token{pos: l.pos, id: COMMA, raw: ","}
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

			return Token{pos: l.pos, id: ILLEGAL, raw: string(r)}
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
				return Token{pos: pos, id: DECIMAL, raw: raw}
			}
			panic(err)
		}

		if unicode.IsDigit(r) {
			raw = raw + string(r)
		} else {
			l.unread()
			return Token{pos: pos, id: DECIMAL, raw: raw}
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
			return Token{pos: pos, id: LABEL, raw: raw}
		} else {
			l.unread()
			return Token{pos: pos, id: nameTokenId(raw), raw: raw}
		}
	}
}

func nameTokenId(nm string) TokenId {
	switch nm {
	case "section":
		return SECTION
	default:
		return NAME
	}
}
