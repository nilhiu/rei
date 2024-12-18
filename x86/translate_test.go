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

func TestMovRegAddrB(t *testing.T) {
	// mov eax, [rbx]
	addr := x86.Address{1, x86.NilReg, x86.Rbx, 0}
	bytes := testTranslate(t, x86.Mov, x86.Eax, addr)
	expectedBytes := []byte{0x8B, 0x03}
	testBytes(t, bytes, expectedBytes)
}

func TestMovRegAddrIB(t *testing.T) {
	// mov eax, [rbx+rax]
	addr := x86.Address{1, x86.Rax, x86.Rbx, 0}
	bytes := testTranslate(t, x86.Mov, x86.Eax, addr)
	expectedBytes := []byte{0x8B, 0x04, 0x03}
	testBytes(t, bytes, expectedBytes)
}

func TestMovRegAddrBD(t *testing.T) {
	// mov eax, [rbx+0x7FFFFFFF]
	addr := x86.Address{1, x86.NilReg, x86.Rbx, 0x7FFFFFFF}
	bytes := testTranslate(t, x86.Mov, x86.Eax, addr)
	expectedBytes := []byte{0x8B, 0x83, 0xFF, 0xFF, 0xFF, 0x7F}
	testBytes(t, bytes, expectedBytes)
}

func TestMovRegAddrIBD(t *testing.T) {
	// mov eax, [rbx+rax+0xFF]
	addr := x86.Address{1, x86.Rax, x86.Rbx, 0xFF}
	bytes := testTranslate(t, x86.Mov, x86.Eax, addr)
	expectedBytes := []byte{0x8B, 0x84, 0x03, 0xFF, 0x00, 0x00, 0x00}
	testBytes(t, bytes, expectedBytes)
}

func TestMovRegAddrSIBD(t *testing.T) {
	// mov eax, [rbx+2*rax+0xFF]
	addr := x86.Address{2, x86.Rax, x86.Rbx, 0xFF}
	bytes := testTranslate(t, x86.Mov, x86.Eax, addr)
	expectedBytes := []byte{0x8B, 0x84, 0x43, 0xFF, 0x00, 0x00, 0x00}
	testBytes(t, bytes, expectedBytes)
}

func TestMovRegAddrNil(t *testing.T) {
	bytes := testTranslate(t, x86.Mov, x86.Eax, x86.Address{})
	expectedBytes := []byte{0x8B, 0x04, 0x25, 0x00, 0x00, 0x00, 0x00}
	testBytes(t, bytes, expectedBytes)
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
