package rasm

import "io"

// A ExprID represents the type of an expression emitted by the [Parser].
type ExprID uint

const (
	EOFExpr     ExprID = iota // represents an end of file
	InstrExpr                 // represents an instruction expression
	SectionExpr               // represents a section expression
	LabelExpr                 // represents a label expression
	IllegalExpr               // represents an illegal/unknown expression
)

// Expr represents the expression parsed by the [Parser].
type Expr struct {
	ID       ExprID
	Root     Token
	Children []Token
}

// A Parser is an object that takes the [Token]s emitted by the [Lexer]
// to create known assembly expressions.
type Parser struct {
	lxr  *Lexer
	root Token
}

// NewParser creates a new parser based on the given [io.Reader].
func NewParser(rd io.Reader) *Parser {
	return &Parser{lxr: NewLexer(rd)}
}

// NewParserLexer creates a new parser based on the given [Lexer].
func NewParserLexer(lxr *Lexer) *Parser {
	return &Parser{lxr: lxr}
}

// Next parses and returns the next expression. If the file has been fully
// read, Next will always return a [EOFExpr] expression.
func (p *Parser) Next() Expr {
	for {
		tok := p.lxr.Next()
		switch tok.ID() {
		case Newline:
			continue
		case Section:
			p.root = tok
			return p.parseSection()
		case Instruction:
			p.root = tok
			return p.parseInstruction()
		case Identifier:
			p.root = tok
			return p.parseLabel()
		case EOF:
			return Expr{EOFExpr, tok, nil}
		}

		return Expr{ID: IllegalExpr, Root: tok}
	}
}

func (p *Parser) parseInstruction() Expr {
	children := []Token{}

	for {
		op := p.lxr.Next()
		switch op.ID() {
		case Newline, EOF:
			return Expr{ID: InstrExpr, Root: p.root, Children: children}
		case Identifier, Decimal, Hex, Octal, Register:
			children = append(children, op)
		default:
			children = append(children, op)
			return Expr{
				ID:       IllegalExpr,
				Root:     p.root,
				Children: append([]Token{{raw: "expected operand or '\\n'"}}, children...),
			}
		}

		delim := p.lxr.Next()
		switch delim.ID() {
		case Newline, EOF:
			return Expr{ID: InstrExpr, Root: p.root, Children: children}
		case Comma:
			continue
		default:
			children = append(children, delim)
			return Expr{
				ID:       IllegalExpr,
				Root:     p.root,
				Children: append([]Token{{raw: "expected '\\n' or ','"}}, children...),
			}
		}
	}
}

func (p *Parser) parseSection() Expr {
	ident := p.lxr.Next()
	if ident.ID() != Identifier {
		return Expr{
			ID:       IllegalExpr,
			Root:     p.root,
			Children: []Token{{raw: "expected identifier"}, ident},
		}
	}
	return Expr{ID: SectionExpr, Root: p.root, Children: []Token{ident}}
}

func (p *Parser) parseLabel() Expr {
	colon := p.lxr.Next()
	if colon.ID() != Colon {
		return Expr{ID: IllegalExpr, Root: p.root, Children: []Token{{raw: "expected ':'"}, colon}}
	}
	return Expr{ID: LabelExpr, Root: p.root, Children: nil}
}
