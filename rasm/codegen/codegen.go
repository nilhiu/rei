package codegen

import (
	"github.com/nilhiu/rei/rasm/parser"
)

// A CodeGen represents an object that turns the expressions parsed by the
// [Parser], to machine code.
type CodeGen interface {
	// File returns the current file being worked on.
	File() string

	// Next generates machine code for the next [parser.Instruction] expression.
	// The machine code generation follows from the [Translator] passed to [New].
	// If the file has been fully read, Next will always return a error [lexer.ErrEOF].
	Next() (Code, error)

	// Labels returns a map from the assembler labels that have been encountered
	// during the code generation to additional information about them.
	Labels() map[string]LabelInfo
}

type codeGen struct {
	p          parser.Parser
	section    string
	sectPos    map[string]uint64
	labels     map[string]LabelInfo
	translator Translator
}

// A LabelInfo represents information about a label.
type LabelInfo struct {
	Section string // the section the label is located in
	Offset  uint64 // the offset from the section the label's at
}

// New creates a new code generator based on the [Parser] given to it.
func New(p parser.Parser, translator Translator) CodeGen {
	return &codeGen{
		p:          p,
		section:    ".text",
		sectPos:    map[string]uint64{},
		labels:     map[string]LabelInfo{},
		translator: translator,
	}
}

func (cg *codeGen) File() string {
	return cg.p.File()
}

func (cg *codeGen) Next() (Code, error) {
	for {
		expr, err := cg.p.Next()
		if err != nil {
			return nil, &Error{
				File:   cg.File(),
				Reason: ErrReasonParser,
				err:    err,
			}
		}

		switch expr.ID() {
		case parser.Label:
			ok := cg.addLabel(expr.Root().Raw())
			if !ok {
				return nil, &Error{
					File:   cg.File(),
					Reason: ErrReasonLabelExists,
					Expr:   expr,
				}
			}

			continue
		case parser.Instruction:
			bytes, err := cg.genInstruction(expr)
			if err != nil {
				return nil, err
			}

			cg.addCurrentSectOff(uint64(len(bytes)))

			return NewCode(cg.section, bytes), err
		case parser.Section:
			cg.section = expr.Root().Raw()
			continue
		}

		return nil, &Error{
			File:   cg.File(),
			Reason: ErrReasonUnsupportedExpression,
			Expr:   expr,
		}
	}
}

func (cg *codeGen) Labels() map[string]LabelInfo {
	return cg.labels
}

func (cg *codeGen) addLabel(label string) bool {
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

func (cg *codeGen) genInstruction(expr parser.Expr) ([]byte, error) {
	ops := []Operand{}

	for _, expr := range expr.Children() {
		op := NewOperand(expr.Root())
		if op == nil {
			return nil, &Error{
				File:   cg.File(),
				Reason: ErrReasonUnrecognizedOperand,
				Expr:   expr,
			}
		}

		ops = append(ops, op)
	}

	bytes, err := cg.translator.Translate(expr.Root().SubID(), ops...)
	if err != nil {
		return nil, &Error{
			File:   cg.File(),
			Reason: ErrReasonTranslator,
			err:    err,
		}
	}

	return bytes, nil
}

func (cg *codeGen) addCurrentSectOff(off uint64) {
	cg.sectPos[cg.section] += off
}

func (cg *codeGen) getCurrentSectOff() uint64 {
	return cg.sectPos[cg.section]
}
