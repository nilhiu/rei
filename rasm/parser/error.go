package parser

import (
	"fmt"
	"strings"

	"github.com/nilhiu/rei/rasm/lexer"
)

// ErrorReason is a enum of all errors the parser can return.
type ErrorReason uint

const (
	ErrReasonLexer ErrorReason = iota
	ErrReasonIllegal
	ErrReasonUnexpectedNewline
	ErrReasonNoSectionName
	ErrReasonUnexpectedIdentifier
)

func (r ErrorReason) String() string {
	switch r {
	case ErrReasonLexer:
		return "lexer error"
	case ErrReasonIllegal:
		return "illegal syntax encountered"
	case ErrReasonUnexpectedNewline:
		return "unexpected newline"
	case ErrReasonNoSectionName:
		return "no section name"
	case ErrReasonUnexpectedIdentifier:
		return "unexpected identifier"
	}

	return "unknown reason"
}

// Error represents the full error of [Parser]. It provides everything to
// track down exactly what and where the error happend in the input file.
type Error struct {
	File     string
	Reason   ErrorReason
	Token    lexer.Token
	lexerErr error
	Expected string
}

// General list of errors that should only be used for [errors.Is] calls.
var (
	ErrLexer                error = &Error{Reason: ErrReasonLexer}
	ErrIllegal              error = &Error{Reason: ErrReasonIllegal}
	ErrUnexpectedNewline    error = &Error{Reason: ErrReasonUnexpectedNewline}
	ErrNoSectionName        error = &Error{Reason: ErrReasonNoSectionName}
	ErrUnexpectedIdentifier error = &Error{Reason: ErrReasonUnexpectedIdentifier}
)

func (e *Error) Unwrap() error {
	return e.lexerErr
}

func (e *Error) Error() string {
	sb := strings.Builder{}

	if e.Token == nil {
		sb.WriteString(e.Reason.String())
	} else {
		sb.WriteString(fmt.Sprintf(
			"%s[%d:%d]: %s",
			e.File,
			e.Token.Pos().Line,
			e.Token.Pos().Col,
			e.Reason,
		))
	}

	if e.lexerErr != nil {
		sb.WriteString(": ")
		sb.WriteString(e.lexerErr.Error())
	}

	return sb.String()
}

func (e *Error) Is(target error) bool {
	perr, ok := target.(*Error)
	if !ok {
		return false
	}

	return e.Reason == perr.Reason
}
