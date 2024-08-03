package x86_test

import (
	"testing"

	"github.com/nilhiu/rei/x86"
)

// TODO: Need to create much more efficient unit tests.

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
