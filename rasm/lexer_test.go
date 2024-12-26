package rasm_test

import (
	"strings"
	"testing"

	"github.com/nilhiu/rei/rasm"
)

func TestIdentifierLexing(t *testing.T) {
	expected := []string{"mov", "raw", "r1", "r31", "lo_bit", "open.win", "slo_._mo", ".text", "_.o"}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Raw() != expected[i] {
			t.Fatalf(`Incorrect lexing detected. Expected: %q, got: %q`, expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.Eof {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}

func TestDecimalLexing(t *testing.T) {
	expected := []string{"12345678", "2024", "7", "1", "000000", "389", "471208939810923819716278", "0", "0"}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Raw() != expected[i] {
			t.Fatalf(`Incorrect lexing detected. Expected: %q, got: %q`, expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.Eof {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}

func TestHexLexing(t *testing.T) {
	expected := []string{"0x2830FA", "0X000A", "0xABCDEF0123456789", "0x123456789", "0XABFEF7510"}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Raw() != expected[i] && tok.Id() != rasm.Hex {
			t.Fatalf("Incorrect lexing detected. Expected: %q, got: %q", expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.Eof {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}

func TestIllegalHexLexing(t *testing.T) {
	expected := []string{"0x", "0xL", "0XX"}
	expectedId := []rasm.TokenId{rasm.Illegal, rasm.Illegal, rasm.Identifier, rasm.Illegal, rasm.Identifier}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Raw() != expected[i] && tok.Id() != expectedId[i] {
			t.Fatalf("Illegal lexing not detected. Expected: %q, got: %q", expected[i], tok.Raw())
		}
	}
}

func TestOctalLexing(t *testing.T) {
	expected := []string{"0o273012", "0O1234", "0o01234567", "0O1234567", "0o12541237"}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Raw() != expected[i] && tok.Id() != rasm.Octal {
			t.Fatalf("Incorrect lexing detected. Expected: %q, got: %q", expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.Eof {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}

func TestIllegalOctalLexing(t *testing.T) {
	expected := []string{"0o", "0oL", "0OO"}
	expectedId := []rasm.TokenId{rasm.Illegal, rasm.Illegal, rasm.Identifier, rasm.Illegal, rasm.Identifier}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Raw() != expected[i] && tok.Id() != expectedId[i] {
			t.Fatalf("Illegal lexing not detected. Expected: %q, got: %q", expected[i], tok.Raw())
		}
	}
}

func TestIllegalLexing(t *testing.T) {
	expected := []string{"_", "\\", "{", "}"}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Id() != rasm.Illegal && tok.Raw() != expected[i] {
			t.Fatalf(`Illegal lexing let through. Expected: %q, got: %q`, expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.Eof {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}

func TestGeneralKeywordLexing(t *testing.T) {
	expected := []string{"section", "sEcTiOn", "SECTION", "sectionnot", "section_", ",", ":"}
	expectedId := []rasm.TokenId{
		rasm.Section, rasm.Section, rasm.Section, rasm.Identifier, rasm.Identifier, rasm.Comma,
		rasm.Colon,
	}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected); i++ {
		tok := lxr.Next()
		if tok.Id() != expectedId[i] && tok.Raw() != expected[i] {
			t.Fatalf(`Keyword incorrectly lexed. Expected: %q, got: %q`, expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.Eof {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}

func TestX86KeywordLexing(t *testing.T) {
	expected := []string{
		"rax", "rcx", "rdx", "rBx", "rsi", "rDi", "rsp", "rbp", "r9", "r10", "r11", "R12", "r13", "r14",
		"r15", "eax", "ecx", "edX", "ebx", "esI", "edi", "esp", "ebp", "r9d", "r10d", "r11d", "r12d",
		"r13D", "r14d", "r15d", "Ax", "Cx", "Dx", "bx", "si", "di", "sp", "bp", "r9W", "r10w", "r11w",
		"r12w", "r13W", "r14w", "R15w", "al", "cl", "dl", "bL", "sil", "diL", "sPl", "bpl", "r9b",
		"r10b", "r11b", "r12b", "r13b", "r14B", "r15b", "ah", "Ch", "dh", "bH", "mOv",
	}
	lxr := rasm.NewLexer(
		strings.NewReader(strings.Join(expected, " ")),
	)

	for i := 0; i < len(expected)-1; i++ {
		tok := lxr.Next()
		if tok.Id() != rasm.Register && tok.Raw() != expected[i] {
			t.Fatalf(`Keyword incorrectly lexed. Expected: %q, got: %q`, expected[i], tok.Raw())
		}
	}

	if tok := lxr.Next(); tok.Id() != rasm.Instruction {
		t.Fatalf(`Keyword incorrectly lexed. Expected: %q, got: %q`, expected[len(expected)-1], tok.Raw())
	}

	if tok := lxr.Next(); tok.Id() != rasm.Eof {
		t.Fatalf(`End-of-file expected but got valid token. Got: %q`, tok.Raw())
	}
}

func TestLexPositioning(t *testing.T) {
	str := "\nname mov,\n0xAAFF0 0o1234 0 0 000\n12418\n\n\nsection\n    label: random_name\n19370 0"
	expected := []rasm.Position{
		{Line: 1, Col: 0},
		{Line: 2, Col: 0},
		{Line: 2, Col: 5},
		{Line: 2, Col: 8},
		{Line: 2, Col: 9},
		{Line: 3, Col: 0},
		{Line: 3, Col: 8},
		{Line: 3, Col: 15},
		{Line: 3, Col: 17},
		{Line: 3, Col: 19},
		{Line: 3, Col: 22},
		{Line: 4, Col: 0},
		{Line: 4, Col: 5},
		{Line: 5, Col: 0},
		{Line: 6, Col: 0},
		{Line: 7, Col: 0},
		{Line: 7, Col: 7},
		{Line: 8, Col: 4},
		{Line: 8, Col: 9},
		{Line: 8, Col: 11},
		{Line: 8, Col: 22},
		{Line: 9, Col: 0},
		{Line: 9, Col: 6},
	}
	lxr := rasm.NewLexer(strings.NewReader(str))

	for i := 0; i < 14; i++ {
		tok := lxr.Next()
		if tok.Pos() != expected[i] {
			t.Fatalf(`Incorrect lexer positioning detected. Expected: %+v, got: %+v (%q)`, expected[i], tok.Pos(), tok.Raw())
		}
	}
}
