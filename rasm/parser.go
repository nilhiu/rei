package rasm

import "io"

type ExprId uint

const (
	EofExpr ExprId = iota
	InstrExpr
	SectionExpr
	LabelExpr
	IllegalExpr
)

type Expr struct {
	Id       ExprId
	Root     Token
	Children []Token
}

type Parser struct {
	lxr  *Lexer
	root Token
}

func NewParser(rd io.Reader) *Parser {
	return &Parser{lxr: NewLexer(rd)}
}

func NewParserLexer(lxr *Lexer) *Parser {
	return &Parser{lxr: lxr}
}

func (p *Parser) Next() Expr {
	for {
		tok := p.lxr.Next()
		switch tok.Id() {
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
		case Eof:
			return Expr{EofExpr, tok, nil}
		}

		return Expr{Id: IllegalExpr, Root: tok}
	}
}

func (p *Parser) parseInstruction() Expr {
	children := []Token{}
	for {
		op := p.lxr.Next()
		switch op.Id() {
		case Newline, Eof:
			return Expr{Id: InstrExpr, Root: p.root, Children: children}
		case Identifier, Decimal, Hex, Octal, Register:
			children = append(children, op)
		default:
			children = append(children, op)
			return Expr{
				Id:       IllegalExpr,
				Root:     p.root,
				Children: append([]Token{{raw: "expected operand or '\\n'"}}, children...),
			}
		}

		delim := p.lxr.Next()
		switch delim.Id() {
		case Newline, Eof:
			return Expr{Id: InstrExpr, Root: p.root, Children: children}
		case Comma:
			continue
		default:
			children = append(children, delim)
			return Expr{
				Id:       IllegalExpr,
				Root:     p.root,
				Children: append([]Token{{raw: "expected '\\n' or ','"}}, children...),
			}
		}
	}
}

func (p *Parser) parseSection() Expr {
	ident := p.lxr.Next()
	if ident.Id() != Identifier {
		return Expr{
			Id:       IllegalExpr,
			Root:     p.root,
			Children: []Token{{raw: "expected identifier"}, ident},
		}
	}
	return Expr{Id: SectionExpr, Root: p.root, Children: []Token{ident}}
}

func (p *Parser) parseLabel() Expr {
	colon := p.lxr.Next()
	if colon.Id() != Colon {
		return Expr{Id: IllegalExpr, Root: p.root, Children: []Token{{raw: "expected ':'"}, colon}}
	}
	return Expr{Id: LabelExpr, Root: p.root, Children: []Token{}}
}
