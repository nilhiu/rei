package lexer

import (
	"fmt"
	"strings"
)

// ErrorReason is a enum of all errors the lexer can return.
type ErrorReason uint

const (
	ErrReasonEOF ErrorReason = iota
	ErrReasonIllegal
	ErrReasonEmptyHexPrefix
	ErrReasonEmptyOctalPrefix
)

func (r ErrorReason) String() string {
	switch r {
	case ErrReasonEOF:
		return "unexpected end-of-file"
	case ErrReasonIllegal:
		return "illegal/unrecognized character encountered"
	case ErrReasonEmptyHexPrefix:
		return "invalid hex literal"
	case ErrReasonEmptyOctalPrefix:
		return "invalid octal literal"
	}

	return "unknown reason"
}

// Error represents the full error of [Lexer].
type Error struct {
	File     string      // the file where the error was encountered.
	Pos      Position    // the location in the file.
	Reason   ErrorReason // the reason of the error.
	Expected string      // what the lexer expected to encounter.
}

// General list of errors that should only be used for [errors.Is] calls.
var (
	ErrEOF              error = &Error{Reason: ErrReasonEOF}
	ErrIllegal          error = &Error{Reason: ErrReasonIllegal}
	ErrEmptyHexPrefix   error = &Error{Reason: ErrReasonEmptyHexPrefix}
	ErrEmptyOctalPrefix error = &Error{Reason: ErrReasonEmptyOctalPrefix}
)

func (e *Error) Error() string {
	sb := strings.Builder{}

	sb.WriteString(fmt.Sprintf(
		"%s[%d:%d]: %s",
		e.File,
		e.Pos.Line,
		e.Pos.Col,
		e.Reason,
	))

	if e.Expected != "" {
		sb.WriteString(fmt.Sprintf(". expected: %s", e.Expected))
	}

	return sb.String()
}

func (e *Error) Is(target error) bool {
	lerr, ok := target.(*Error)
	if !ok {
		return false
	}

	return e.Reason == lerr.Reason
}
