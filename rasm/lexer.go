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
				pos := l.pos
				l.unread()
				raw := l.lexDecimal()
				return Token{pos: pos, id: DECIMAL, raw: raw}
			} else if unicode.IsLetter(r) {
				pos := l.pos
				l.unread()
				raw := l.lexName()
				return Token{pos: pos, id: NAME, raw: raw}
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

func (l *Lexer) lexDecimal() string {
	var raw string
	for {
		r, _, err := l.rd.ReadRune()
		if err != nil {
			if err == io.EOF {
				return raw
			}
			panic(err)
		}

		if unicode.IsDigit(r) {
			raw = raw + string(r)
		} else {
			l.unread()
			return raw
		}
	}
}

func (l *Lexer) lexName() string {
	var raw string
	for {
		r, _, err := l.rd.ReadRune()
		if err != nil {
			if err == io.EOF {
				return raw
			}
			panic(err)
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' {
			raw = raw + string(r)
		} else {
			l.unread()
			return raw
		}
	}
}
