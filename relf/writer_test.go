package relf_test

import (
	"bytes"
	"debug/elf"
	"testing"

	"github.com/nilhiu/rei/relf"
)

func TestWriterCorrectness(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	// TODO: Maybe think of a better way to test this...
	expected := []byte{
		// HEADER
		0x7f, 'E', 'L', 'F', byte(elf.ELFCLASS64), byte(elf.ELFDATA2LSB), 1,
		byte(elf.ELFOSABI_NONE), 0, 0, 0, 0, 0, 0, 0, 16, byte(elf.ET_REL), 0,
		byte(elf.EM_X86_64), 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 64,
		0, 3, 0, 2, 0,
		// SECTION HEADER null
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		// SECTION HEADER .text
		1, 0, 0, 0, byte(elf.SHT_PROGBITS), 0, 0, 0, byte(elf.SHF_EXECINSTR | elf.SHF_ALLOC),
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 7, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		// SECTION HEADER .shstrtab
		7, 0, 0, 0, byte(elf.SHT_STRTAB), 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 7, 1, 0, 0, 0, 0, 0, 0, 17, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		// SECTION DATA .text
		0x48, 0xc7, 0xc0, 0xff, 0xff, 0x00, 0x00,
		// SECTION DATA .shstrtab
		0, '.', 't', 'e', 'x', 't', 0, '.', 's', 'h', 's', 't', 'r', 't', 'a', 'b', 0,
	}

	w := relf.New("test.S", relf.Header64{
		Endian:  elf.ELFDATA2LSB,
		ABI:     elf.ELFOSABI_NONE,
		Machine: elf.EM_X86_64,
	}, buf)

	err := w.WriteSection(relf.Section64{
		Name:      ".text",
		Code:      []byte{0x48, 0xC7, 0xC0, 0xFF, 0xFF, 0x00, 0x00},
		Type:      elf.SHT_PROGBITS,
		Addralign: 16,
		Entsize:   0,
		Flags:     elf.SHF_EXECINSTR | elf.SHF_ALLOC,
	})
	if err != nil {
		t.Fatalf("failed to write section, %s", err)
	}

	if err := w.Flush(); err != nil {
		t.Fatalf("failed to flush to output, %s", err)
	}

	got := buf.Bytes()
	if !bytes.Equal(got, expected) {
		t.Fatalf("incorrect output, expected:\n%X\ngot:\n%X", expected, got)
	}
}
