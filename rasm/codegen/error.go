package codegen

import (
	"fmt"
	"strings"

	"github.com/nilhiu/rei/rasm/parser"
)

// ErrorReason is a enum of all errors the code generator can return.
type ErrorReason uint

const (
	ErrReasonLabelExists ErrorReason = iota
	ErrReasonUnsupportedExpression
	ErrReasonUnrecognizedOperand
	ErrReasonTranslator
	ErrReasonParser
)

func (r ErrorReason) String() string {
	switch r {
	case ErrReasonLabelExists:
		return "label with the same name already exists"
	case ErrReasonUnsupportedExpression:
		return "the code generation of the given expression not supported"
	case ErrReasonUnrecognizedOperand:
		return "the given operand is not a recognized operand"
	case ErrReasonTranslator:
		return "translator error"
	case ErrReasonParser:
		return "parser error"
	}

	return "unknown reason"
}

// Error represents the full error of [CodeGen].
type Error struct {
	File   string
	Reason ErrorReason
	Expr   parser.Expr
	err    error
}

// General list of errors that should only be used for [errors.Is] calls.
var (
	ErrLabelExists           error = &Error{Reason: ErrReasonLabelExists}
	ErrUnsupportedExpression error = &Error{Reason: ErrReasonUnsupportedExpression}
	ErrUnrecognizedOperand   error = &Error{Reason: ErrReasonUnrecognizedOperand}
	ErrTranslator            error = &Error{Reason: ErrReasonTranslator}
	ErrParser                error = &Error{Reason: ErrReasonParser}
)

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Error() string {
	sb := strings.Builder{}

	if e.Expr == nil {
		sb.WriteString(e.Reason.String())
	} else {
		sb.WriteString(fmt.Sprintf(
			"%s[%d:%d]: %s",
			e.File,
			e.Expr.Root().Pos().Line,
			e.Expr.Root().Pos().Col,
			e.Reason,
		))
	}

	if e.err != nil {
		sb.WriteString(": ")
		sb.WriteString(e.err.Error())
	}

	return sb.String()
}

func (e *Error) Is(target error) bool {
	cerr, ok := target.(*Error)
	if !ok {
		return false
	}

	return e.Reason == cerr.Reason
}
