package x86_test

import (
	"testing"

	"github.com/nilhiu/rei/x86"
)

// TODO: Need to create much more efficient unit tests.

func TestAddRegReg(t *testing.T) {
	bytes := testTranslate(t, x86.Add, x86.Rcx, x86.Rax)
	expected := []byte{0x48, 0x01, 0xC1}
	testBytes(t, bytes, expected)
}

func TestAddRegImm(t *testing.T) {
	bytes0 := testTranslate(t, x86.Add, x86.Ecx, x86.Immediate(0xA4))
	bytes1 := testTranslate(t, x86.Add, x86.Ax, x86.Immediate(0xA4A1))
	bytes2 := testTranslate(t, x86.Add, x86.Rax, x86.Immediate(0xA4))
	bytes3 := testTranslate(t, x86.Add, x86.Rbx, x86.Immediate(0x7F))
	all := append(append(append(bytes0, bytes1...), bytes2...), bytes3...)
	expected := []byte{
		0x81, 0xC1, 0xA4, 0x00, 0x00, 0x00,
		0x66, 0x05, 0xA1, 0xA4,
		0x48, 0x05, 0xA4, 0x00, 0x00, 0x00,
		0x48, 0x83, 0xC3, 0x7F,
	}
	testBytes(t, all, expected)
}

func TestMovRegImm(t *testing.T) {
	bytes := testTranslate(t, x86.Mov, x86.Ecx, x86.Immediate(591))
	expected := []byte{0xB9, 0x4F, 0x02, 0x00, 0x00}
	testBytes(t, bytes, expected)
}

func TestMovRegReg(t *testing.T) {
	bytes := testTranslate(t, x86.Mov, x86.R15w, x86.R15w)
	expected := []byte{0x66, 0x45, 0x89, 0xFF}
	testBytes(t, bytes, expected)
}

func testTranslate(t *testing.T, mnem x86.Mnemonic, ops ...x86.Operand) []byte {
	bytes, err := x86.Translate(mnem, ops...)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func testBytes(t *testing.T, b []byte, expect []byte) {
	for i := 0; i < len(expect); i++ {
		if b[i] != expect[i] {
			t.Fatalf("Incorrect x86 translation detected. Byte #%d\nExpected: %X\n     Got: %X",
				i, expect, b)
		}
	}
}

func TestMovRegImm(t *testing.T) {
	bytes, err := x86.Translate(x86.Mov, x86.Ecx, x86.Immediate(591))
	expectedBytes := []byte{0xB9, 0x4F, 0x02, 0x00, 0x00}
	if err != nil {
		t.Fatal(err)
	}

	if len(expectedBytes) != len(bytes) {
		t.Fatalf("Incorrect x86 translation detected. Expected byte count: %d, got: %d",
			len(expectedBytes), len(bytes))
	}

	for i := 0; i < len(expectedBytes); i++ {
		if bytes[i] != expectedBytes[i] {
			t.Fatalf("Incorrect x86 translation detected. Byte #%d, expected: %X, got: %X",
				i, expectedBytes[i], bytes[i])
		}
	}
}

func TestMovRegReg(t *testing.T) {
	bytes, err := x86.Translate(x86.Mov, x86.R15w, x86.R15w)
	expectedBytes := []byte{0x66, 0x45, 0x89, 0xFF}
	if err != nil {
		t.Fatal(err)
	}

	if len(expectedBytes) != len(bytes) {
		t.Fatalf("Incorrect x86 translation detected. Expected byte count: %d, got: %d",
			len(expectedBytes), len(bytes))
	}

	for i := 0; i < len(expectedBytes); i++ {
		if bytes[i] != expectedBytes[i] {
			t.Fatalf("Incorrect x86 translation detected. Byte #%d, expected: %X, got: %X",
				i, expectedBytes[i], bytes[i])
		}
	}
}
