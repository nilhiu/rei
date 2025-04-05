package parser

import (
	"errors"

	"github.com/nilhiu/rei/rasm/lexer"
)

// Parser the interface for rei's assmebly parser.
type Parser interface {
	// File returns the file that the parser is parsing.
	File() string

	// Next returns the next expression in the file. If EOF is encountered,
	// the error [lexer.ErrEOF] is returned.
	Next() (Expr, error)
}

type parser struct {
	lxr  lexer.Lexer
	root lexer.Token
}

// New creates a new parser based on the given [Lexer].
func New(lxr lexer.Lexer) Parser {
	return &parser{lxr: lxr}
}

func (p *parser) File() string {
	return p.lxr.File()
}

func (p *parser) Next() (Expr, error) {
	for {
		tok, err := p.lxr.Next()
		if err != nil {
			return nil, &Error{
				File:     p.File(),
				Reason:   ErrReasonLexer,
				lexerErr: err,
			}
		}

		switch tok.ID() {
		case lexer.Newline:
			continue
		case lexer.Section:
			p.root = tok
			return p.parseSection()
		case lexer.Instruction:
			p.root = tok
			return p.parseInstruction()
		case lexer.Identifier:
			p.root = tok
			return p.parseLabel()
		}

		return nil, &Error{
			File:   p.File(),
			Reason: ErrReasonIllegal,
			Token:  tok,
		}
	}
}

func (p *parser) parseInstruction() (Expr, error) {
	children := []Expr{}

	for {
		op, err := p.parseOperand()
		if err != nil {
			if errors.Is(err, lexer.ErrEOF) || errors.Is(err, ErrUnexpectedNewline) {
				return NewExpr(Instruction, p.root, children...), nil
			}

			return nil, err
		}

		children = append(children, op)

		skipped, err := p.skipDelimiter()
		if err != nil {
			var lerr *lexer.Error
			if errors.As(err, &lerr) && lerr.Is(lexer.ErrEOF) {
				return NewExpr(Instruction, p.root, children...), nil
			}

			return nil, err
		}

		if !skipped {
			return NewExpr(Instruction, p.root, children...), nil
		}
	}
}

func (p *parser) parseOperand() (Expr, error) {
	op, err := p.lxr.Next()
	if err != nil {
		return nil, &Error{
			File:     p.File(),
			Reason:   ErrReasonLexer,
			lexerErr: err,
		}
	}

	switch op.ID() {
	case lexer.Newline:
		return nil, &Error{
			File:     p.File(),
			Reason:   ErrReasonUnexpectedNewline,
			Token:    op,
			Expected: "operand",
		}
	case lexer.Identifier, lexer.Decimal, lexer.Hex, lexer.Octal, lexer.Register:
		return NewExpr(Leaf, op), nil
	default:
		return nil, &Error{
			File:     p.File(),
			Reason:   ErrReasonIllegal,
			Token:    op,
			Expected: "operand",
		}
	}
}

func (p *parser) skipDelimiter() (bool, error) {
	tok, err := p.lxr.Next()
	if err != nil {
		return false, &Error{
			File:     p.File(),
			Reason:   ErrReasonLexer,
			lexerErr: err,
		}
	}

	switch tok.ID() {
	case lexer.Newline:
		return false, nil
	case lexer.Comma:
		return true, nil
	default:
		return false, &Error{
			File:     p.File(),
			Reason:   ErrReasonIllegal,
			Token:    tok,
			Expected: "'\\n' or ','",
		}
	}
}

func (p *parser) parseSection() (Expr, error) {
	tok, err := p.lxr.Next()
	if err != nil {
		return nil, &Error{
			File:     p.File(),
			Reason:   ErrReasonLexer,
			lexerErr: err,
		}
	}

	if tok.ID() != lexer.Identifier {
		return nil, &Error{
			File:     p.File(),
			Reason:   ErrReasonNoSectionName,
			Token:    tok,
			Expected: "identifier",
		}
	}

	return NewExpr(Section, tok), nil
}

func (p *parser) parseLabel() (Expr, error) {
	tok, err := p.lxr.Next()
	if err != nil {
		return nil, &Error{
			File:     p.File(),
			Reason:   ErrReasonLexer,
			lexerErr: err,
		}
	}

	if tok.ID() != lexer.Colon {
		return nil, &Error{
			File:     p.File(),
			Reason:   ErrReasonUnexpectedIdentifier,
			Token:    tok,
			Expected: "':'",
		}
	}

	return NewExpr(Label, p.root), nil
}
