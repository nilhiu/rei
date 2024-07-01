package rasm_test

import (
	"strings"
	"testing"

	"github.com/nilhiu/rei/rasm"
)

func TestNameLexing(t *testing.T) {
	expected := []string{"mov", "raw", "r1", "r31", "lo_bit", "open.win", "slo_._mo"}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Raw() != expected[i] {
			t.Fatalf(`Incorrect lexing detected. Expected %q, got: %q`, expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.EOF {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}

func TestDecimalLexing(t *testing.T) {
	expected := []string{"12345678", "2024", "7", "1", "000000", "389", "471208939810923819716278"}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Raw() != expected[i] {
			t.Fatalf(`Incorrect lexing detected. Expected %q, got: %q`, expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.EOF {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}

func TestIllegalLexing(t *testing.T) {
	expected := []string{"_", "\\", "{", "}"}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Id() != rasm.ILLEGAL && tok.Raw() != expected[i] {
			t.Fatalf(`Illegal lexing let through. Expected: %q, got: %q`, expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.EOF {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}