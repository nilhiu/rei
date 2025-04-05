package lexer

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

// A Lexer represents an object which processes input to output tokens.
type Lexer interface {
	// File returns the file that the lexer is currently lexing.
	File() string

	// Next returns the next token in the file. If EOF is encountered, the error
	// [ErrEOF] is returned.
	Next() (Token, error)
}

type lexer struct {
	// file is the source file being lexed
	file string

	// rd is the reader used by the lexer to lex.
	rd *bufio.Reader

	// pos is the current position in file.
	pos Position

	// sb is used to build string efficiently.
	sb strings.Builder

	// idMaps are identifier maps that are provided by the user for
	// architecture specific keywords, like registers and instructions.
	idMaps []IdentifierMap
}

// New create a new [Lexer] based on the [io.Reader] and [IdentifierMap]s
// given to it.
func New(file string, rd io.Reader, idMaps ...IdentifierMap) Lexer {
	return &lexer{
		file:   file,
		rd:     bufio.NewReader(rd),
		pos:    NewPosition(1, 0),
		sb:     strings.Builder{},
		idMaps: idMaps,
	}
}

func (l *lexer) File() string {
	return l.file
}

func (l *lexer) Next() (Token, error) {
	for {
		pos := l.pos

		r, isEOF := l.read()
		if isEOF {
			return nil, &Error{
				File:   l.file,
				Pos:    pos,
				Reason: ErrReasonEOF,
			}
		}

		switch r {
		case ',':
			return NewToken(pos, Comma, ","), nil
		case ':':
			return NewToken(pos, Colon, ":"), nil
		case '0':
			return l.lexZero()
		case '\n':
			l.pos.Line++
			l.pos.Col = 0

			return NewToken(pos, Newline, "\\n"), nil
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

			return nil, &Error{
				File:   l.file,
				Pos:    pos,
				Reason: ErrReasonIllegal,
			}
		}
	}
}

func (l *lexer) unread() {
	if err := l.rd.UnreadRune(); err != nil {
		panic(err)
	}

	l.pos.Col--
}

func (l *lexer) read() (rune, bool) {
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

func (l *lexer) writeStr(r rune) {
	_, err := l.sb.WriteRune(r)
	if err != nil {
		panic(err)
	}
}

func (l *lexer) popStr() string {
	str := l.sb.String()
	l.sb.Reset()

	return str
}

func (l *lexer) lexZero() (Token, error) {
	r, isEOF := l.read()
	if isEOF {
		return NewToken(NewPosition(l.pos.Line, l.pos.Col-2), Decimal, "0"), nil
	}

	switch r {
	case 'x', 'X':
		return l.lexHex()
	case 'o', 'O':
		return l.lexOctal()
	default:
		if unicode.IsDigit(r) {
			l.unread()
			_tok, err := l.lexDecimal()
			if err != nil {
				return nil, err
			}

			return NewToken(
				NewPosition(_tok.Pos().Line, _tok.Pos().Col-1),
				_tok.ID(),
				"0"+_tok.Raw(),
			), nil
		} else {
			l.unread()

			return NewToken(NewPosition(l.pos.Line, l.pos.Col-1), Decimal, "0"), nil
		}
	}
}

func (l *lexer) lexHex() (Token, error) {
	pos := l.pos
	pos.Col -= 2

	for {
		r, isEOF := l.read()
		if isEOF {
			return nil, &Error{
				File:     l.file,
				Pos:      pos,
				Reason:   ErrReasonEmptyHexPrefix,
				Expected: "0-9, a-f, A-F",
			}
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
					return nil, &Error{
						File:     l.file,
						Pos:      pos,
						Reason:   ErrReasonEmptyHexPrefix,
						Expected: "0-9, a-f, A-F",
					}
				}

				return NewToken(pos, Hex, raw), nil
			}
		}
	}
}

func (l *lexer) lexOctal() (Token, error) {
	pos := l.pos
	pos.Col -= 2

	for {
		r, isEOF := l.read()
		if isEOF {
			return nil, &Error{
				File:     l.file,
				Pos:      pos,
				Reason:   ErrReasonEmptyOctalPrefix,
				Expected: "0-7",
			}
		}

		switch r {
		case '0', '1', '2', '3', '4', '5', '6', '7':
			l.writeStr(r)
		default:
			l.unread()

			raw := l.popStr()
			if raw == "" {
				return nil, &Error{
					File:     l.file,
					Pos:      pos,
					Reason:   ErrReasonEmptyOctalPrefix,
					Expected: "0-7",
				}
			}

			return NewToken(pos, Octal, raw), nil
		}
	}
}

func (l *lexer) lexDecimal() (Token, error) {
	pos := l.pos

	for {
		r, isEOF := l.read()
		if isEOF {
			return NewToken(pos, Decimal, l.popStr()), nil
		}

		if unicode.IsDigit(r) {
			l.writeStr(r)
		} else {
			l.unread()
			return NewToken(pos, Decimal, l.popStr()), nil
		}
	}
}

func (l *lexer) lexIdentifier() (Token, error) {
	pos := l.pos

	for {
		r, isEOF := l.read()
		if isEOF {
			raw := l.popStr()
			id, subID := l.getIdentIDs(raw)
			return NewTokenSubID(pos, id, subID, raw), nil
		}

		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '.' {
			l.writeStr(r)
		} else {
			l.unread()
			raw := l.popStr()
			id, subID := l.getIdentIDs(raw)
			return NewTokenSubID(pos, id, subID, raw), nil
		}
	}
}

func (l *lexer) getIdentIDs(id string) (TokenID, uint32) {
	ident := strings.ToLower(id)
	switch ident {
	case "section":
		return Section, 0
	default:
		for _, idMap := range l.idMaps {
			if idTok, ok := idMap.Find(ident); ok {
				return idTok.tID, idTok.subID
			}
		}

		return Identifier, 0
	}
}
