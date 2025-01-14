package relf_test

import (
	"bytes"
	"debug/elf"
	"testing"

	"github.com/nilhiu/rei/relf"
)

func TestELFWriter(t *testing.T) {
	hdr := relf.Header64{
		Endian:  elf.ELFDATA2LSB,
		ABI:     elf.ELFOSABI_NONE,
		Machine: elf.EM_X86_64,
		Flags:   0,
	}
	buf := bytes.NewBuffer([]byte{})

	tests := []struct {
		name      string
		wantSects []relf.Section64
		wantSymbs []relf.Symbol64
		wantErr   bool
	}{
		{
			name:      "Should generate correctful empty ELF file",
			wantSects: []relf.Section64{},
			wantSymbs: []relf.Symbol64{},
			wantErr:   false,
		},
		{
			name: "Should generate correctful ELF file with a single section",
			wantSects: []relf.Section64{
				{
					Name:      ".text",
					Code:      []byte{0x48, 0xC7, 0xC0, 0xFF, 0xFF, 0x00, 0x00},
					Type:      elf.SHT_PROGBITS,
					Addralign: 16,
					Entsize:   0,
					Flags:     elf.SHF_EXECINSTR | elf.SHF_ALLOC,
				},
			},
			wantSymbs: []relf.Symbol64{},
			wantErr:   false,
		},
		{
			name: "Should generate correctful ELF file with multiple sections",
			wantSects: []relf.Section64{
				{
					Name:      ".text",
					Code:      []byte{0x48, 0xC7, 0xC0, 0xFF, 0xFF, 0x00, 0x00},
					Type:      elf.SHT_PROGBITS,
					Addralign: 16,
					Entsize:   0,
					Flags:     elf.SHF_EXECINSTR | elf.SHF_ALLOC,
				},
				{
					Name:      ".rodata",
					Code:      []byte{0xFF},
					Type:      elf.SHT_PROGBITS,
					Addralign: 4,
					Entsize:   0,
					Flags:     elf.SHF_ALLOC,
				},
				{
					Name:      ".data",
					Code:      []byte{0xAA},
					Type:      elf.SHT_PROGBITS,
					Addralign: 4,
					Entsize:   0,
					Flags:     elf.SHF_WRITE | elf.SHF_ALLOC,
				},
			},
			wantSymbs: []relf.Symbol64{},
			wantErr:   false,
		},
		{
			name: "Should generate correctful ELF file with multiple sections and symbols",
			wantSects: []relf.Section64{
				{
					Name:      ".text",
					Code:      []byte{0x48, 0xC7, 0xC0, 0xFF, 0xFF, 0x00, 0x00},
					Type:      elf.SHT_PROGBITS,
					Addralign: 16,
					Entsize:   0,
					Flags:     elf.SHF_EXECINSTR | elf.SHF_ALLOC,
				},
				{
					Name:      ".rodata",
					Code:      []byte{0xFF},
					Type:      elf.SHT_PROGBITS,
					Addralign: 4,
					Entsize:   0,
					Flags:     elf.SHF_ALLOC,
				},
				{
					Name:      ".data",
					Code:      []byte{0xAA},
					Type:      elf.SHT_PROGBITS,
					Addralign: 4,
					Entsize:   0,
					Flags:     elf.SHF_WRITE | elf.SHF_ALLOC,
				},
			},
			wantSymbs: []relf.Symbol64{
				{
					Name:  "label1",
					Type:  elf.STT_NOTYPE,
					Bind:  elf.STB_LOCAL,
					Shndx: 2,
					Value: 0,
				},
				{
					Name:  "label2",
					Type:  elf.STT_NOTYPE,
					Bind:  elf.STB_LOCAL,
					Shndx: 3,
					Value: 0,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := relf.New("test.S", hdr, buf)
			for _, sect := range tt.wantSects {
				if err := w.WriteSection(sect); err != nil {
					t.Fatalf("w.WriteSection(sect) failed to write section: %v", err)
				}
			}
			for _, symb := range tt.wantSymbs {
				if err := w.WriteSymbol(symb); err != nil {
					t.Fatalf("w.WriteSymbol(sect) failed to write symbol: %v", err)
				}
			}
			w.Flush()

			gotFile, err := elf.NewFile(bytes.NewReader(buf.Bytes()))
			if err != nil {
				t.Fatalf("elf.NewFile(...) failed to read generated file: %v", err)
			}
			defer buf.Reset()

			if tt.wantErr {
				t.Fatal("Next() succeeded unexpectedly")
			}

			for _, wantSect := range tt.wantSects {
				gotSect := gotFile.Section(wantSect.Name)
				if gotSect == nil {
					t.Errorf("gotFile.Section(%s) returned nil, doesn't exist", wantSect.Name)
					continue
				}

				if !equalSect(wantSect, *gotSect) {
					t.Errorf("gotSymb = %v, want %v", *gotSect, wantSect)
				}
			}

			gotSymbs, err := gotFile.Symbols()
			if err != nil {
				t.Errorf("gotFile.Symbols() failed to return symbols: %v", err)
			}

			for _, wantSymb := range tt.wantSymbs {
				gotSymb := findSymb(wantSymb.Name, gotSymbs)
				if gotSymb == nil {
					t.Errorf("gotFile.Symbols() didn't have wanted symbol \"%s\"", wantSymb.Name)
					continue
				}

				if !equalSymb(wantSymb, *gotSymb) {
					t.Errorf("gotSymb = %v, want %v", *gotSymb, wantSymb)
				}
			}
		})
	}
}

func equalSect(relfSect relf.Section64, elfSect elf.Section) bool {
	if relfSect.Type != elfSect.Type {
		return false
	}

	if relfSect.Flags != elfSect.Flags {
		return false
	}

	if relfSect.Addr != elfSect.Addr {
		return false
	}

	if relfSect.Link != elfSect.Link {
		return false
	}

	if relfSect.Info != elfSect.Info {
		return false
	}

	if relfSect.Addralign != elfSect.Addralign {
		return false
	}

	if relfSect.Entsize != elfSect.Entsize {
		return false
	}

	return true
}

func equalSymb(relfSymb relf.Symbol64, elfSymb elf.Symbol) bool {
	if ((byte(relfSymb.Bind) << 4) | byte(relfSymb.Type)) != elfSymb.Info {
		return false
	}

	if relfSymb.Shndx != uint16(elfSymb.Section) {
		return false
	}

	if relfSymb.Value != elfSymb.Value {
		return false
	}

	return true
}

func findSymb(name string, symbs []elf.Symbol) *elf.Symbol {
	for _, symb := range symbs {
		if symb.Name == name {
			return &symb
		}
	}

	return nil
}
