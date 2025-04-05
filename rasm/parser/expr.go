package parser

import "github.com/nilhiu/rei/rasm/lexer"

// A ExprID represents the type of an expression emitted by the [Parser].
type ExprID uint

const (
	Instruction = iota // represents an instruction expression
	Section            // represents a section expression
	Label              // represents a label expression
	Leaf               // represents a leaf node (operands)
)

// Expr represents the expression parsed by the [Parser].
type Expr interface {
	ID() ExprID
	Root() lexer.Token
	Children() []Expr
}

// NewExpr creates a new expression with the given parameters.
func NewExpr(id ExprID, root lexer.Token, children ...Expr) Expr {
	if len(children) == 0 {
		children = nil
	}

	return &expr{
		id:       id,
		root:     root,
		children: children,
	}
}

type expr struct {
	id       ExprID
	root     lexer.Token
	children []Expr
}

func (e *expr) ID() ExprID {
	return e.id
}

func (e *expr) Root() lexer.Token {
	return e.root
}

func (e *expr) Children() []Expr {
	return e.children
}
