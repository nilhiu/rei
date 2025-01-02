package rasm

import (
	"errors"
	"io"
	"strconv"

	"github.com/nilhiu/rei/x86"
)

type CodeGen struct {
	p       *Parser
	section string
	pos     uint
	labels  map[string]uint
}

func NewCodeGen(rd io.Reader) *CodeGen {
	return NewCodeGenParser(NewParser(rd))
}

func NewCodeGenParser(p *Parser) *CodeGen {
	return &CodeGen{p, ".text", 0, map[string]uint{}}
}

func (cg *CodeGen) Next() ([]byte, string, error) {
	for {
		expr := cg.p.Next()

		switch expr.ID {
		case LabelExpr:
			ok := cg.addLabel(expr.Root.Raw())
			if !ok {
				return nil, cg.section, errors.New("label already exists")
			}

			continue
		case InstrExpr:
			bytes, err := cg.genInstruction(expr)
			cg.pos += uint(len(bytes))

			return bytes, cg.section, err
		case SectionExpr:
			cg.section = expr.Children[0].Raw()

			continue
		case EOFExpr:
			return nil, cg.section, nil
		}

		return nil, cg.section, errors.New("codegen expression not supported")
	}
}

func (cg *CodeGen) Labels() map[string]uint {
	return cg.labels
}

func (cg *CodeGen) addLabel(label string) bool {
	_, ok := cg.labels[label]
	if ok {
		return false
	}

	cg.labels[label] = cg.pos

	return true
}

func (cg *CodeGen) genInstruction(expr Expr) ([]byte, error) {
	ops := []x86.Operand{}

	for _, t := range expr.Children {
		op, err := toOperand(t)
		if err != nil {
			return nil, err
		}

		ops = append(ops, op)
	}

	return x86.Translate(x86.Mnemonic(expr.Root.SpecID()), ops...)
}

func toOperand(t Token) (x86.Operand, error) {
	// TODO: Add other operands.
	switch t.ID() {
	case Decimal:
		i, err := strconv.ParseUint(t.Raw(), 10, 64)

		return x86.Immediate(i), err
	case Register:
		return x86.Register(t.SpecID()), nil
	}

	return nil, errors.New("not supported operand")
}
