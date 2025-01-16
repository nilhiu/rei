package rasm

import (
	"errors"
	"io"
	"strconv"

	"github.com/nilhiu/rei/x86"
)

// A CodeGen represents an object that turns the expressions parsed by the
// [Parser], to machine code.
type CodeGen struct {
	p       *Parser
	section string
	sectPos map[string]uint64
	labels  map[string]LabelInfo
}

// A LabelInfo represents information about a label.
type LabelInfo struct {
	Section string // the section the label is located in
	Offset  uint64 // the offset from the section the label's at
}

// NewCodeGen creates a new code generator based on the [io.Reader] given to it.
func NewCodeGen(rd io.Reader) *CodeGen {
	return NewCodeGenParser(NewParser(rd))
}

// NewCodeGenParser creates a new code generator based on
// the [Parser] given to it.
func NewCodeGenParser(p *Parser) *CodeGen {
	return &CodeGen{
		p:       p,
		section: ".text",
		sectPos: map[string]uint64{},
		labels:  map[string]LabelInfo{},
	}
}

// Next generates machine code for the next [InstrExpr] expression. It returns
// the machine code itself, the section it's in, and possibly an error. If the
// file has been fully read, Next will always return a nil slice with no error.
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
			cg.addCurrentSectOff(uint64(len(bytes)))

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

// Labels returns a map of names to label information of the encountered
// labels by the [CodeGen].
func (cg *CodeGen) Labels() map[string]LabelInfo {
	return cg.labels
}

func (cg *CodeGen) addLabel(label string) bool {
	_, ok := cg.labels[label]
	if ok {
		return false
	}

	cg.labels[label] = LabelInfo{
		Section: cg.section,
		Offset:  cg.getCurrentSectOff(),
	}

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

func (cg *CodeGen) addCurrentSectOff(off uint64) {
	cg.sectPos[cg.section] += off
}

func (cg CodeGen) getCurrentSectOff() uint64 {
	return cg.sectPos[cg.section]
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
