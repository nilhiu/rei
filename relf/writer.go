// The relf package implements routines to write ELF object files.
//
// This package uses a combination of its own definitions of ELF structures,
// while also utilizing the [debug/elf] package from the standard library.
// As an example, when writing sections, or symbols, it uses the [Section64],
// or [Symbol64], but internally it's saved as a [elf.Section64], or [elf.Sym64].
package relf

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"io"
	"slices"
	"strings"
)

const (
	Header64Size  = 64 // the size of an ELF header (64-bit)
	Section64Size = 64 // the size of an ELF section header (64-bit)
	Symbol64Size  = 24 // the size of an ELF symbol (64-bit)
)

// Writer implements an ELF file writer.
type Writer struct {
	header elf.Header64

	sections []elf.Section64
	shndx    map[string]uint16
	shstrtab strings.Builder

	symbols []elf.Sym64
	strtab  strings.Builder

	code   bytes.Buffer
	output io.Writer
}

// Header64 represents a minified version of the ELF file header.
type Header64 struct {
	Endian  elf.Data
	ABI     elf.OSABI
	Machine elf.Machine
	Flags   uint32
}

// Section64 represents a extended version of the ELF section.
type Section64 struct {
	Name      string          // the name of the section
	Type      elf.SectionType // the type of the section
	Flags     elf.SectionFlag // the flags of the section
	Addr      uint64          // the virtual address of the beginning of the section
	Link      uint32          // the index of an associated section
	Info      uint32          // extra information about the section
	Addralign uint64          // the required alignment of the section
	Entsize   uint64          // the size, in bytes, of each entry in the section
	Code      []byte          // is the code/bytes associated with the section
}

// Symbol64 represents an ELF symbol.
type Symbol64 struct {
	Name  string
	Type  elf.SymType
	Bind  elf.SymBind
	Shndx uint16
	Value uint64
}

// New returns a new [Writer] to write an ELF file to the given writer.
// The filename is the source assembly file's name. It's needed as it has
// to be encoded into the symbol table.
//
// For now, this function doesn't return an error, but instead panics if
// an error is encountered. This behaviour will most likely change.
func New(filename string, hdr Header64, writer io.Writer) *Writer {
	w := Writer{
		header: elf.Header64{
			Ident: [elf.EI_NIDENT]byte{
				0x7f, 'E', 'L', 'F',
				byte(elf.ELFCLASS64),
				byte(hdr.Endian),
				byte(elf.EV_CURRENT),
				byte(hdr.ABI),
				0, 0, 0, 0, 0, 0, 0,
				elf.EI_NIDENT,
			},
			Type:    uint16(elf.ET_REL),
			Machine: uint16(hdr.Machine),
			Version: uint32(elf.EV_CURRENT),
			// Place the section headers right after header.
			Shoff:     Header64Size,
			Flags:     hdr.Flags,
			Shnum:     1,
			Ehsize:    Header64Size,
			Shentsize: Section64Size,
		},
		sections: []elf.Section64{{}},
		symbols:  []elf.Sym64{{}},
		shndx:    map[string]uint16{"": 0},
		shstrtab: strings.Builder{},
		output:   writer,
	}

	if err := w.shstrtab.WriteByte(0); err != nil {
		panic(err)
	}

	if err := w.strtab.WriteByte(0); err != nil {
		panic(err)
	}

	err := w.WriteSymbol(Symbol64{
		Name:  filename,
		Type:  elf.STT_FILE,
		Bind:  elf.STB_LOCAL,
		Shndx: uint16(elf.SHN_ABS),
	})
	if err != nil {
		panic(err)
	}

	return &w
}

// WriteSection writes the given section internally in the [Writer].
func (w *Writer) WriteSection(sect Section64) error {
	w.sections = append(w.sections, elf.Section64{
		Name:      uint32(w.shstrtab.Len()),
		Type:      uint32(sect.Type),
		Flags:     uint64(sect.Flags),
		Addr:      sect.Addr,
		Off:       uint64(w.code.Len()),
		Size:      uint64(len(sect.Code)),
		Link:      sect.Link,
		Info:      sect.Info,
		Addralign: sect.Addralign,
		Entsize:   sect.Entsize,
	})

	w.shndx[sect.Name] = w.header.Shnum

	err := w.WriteSymbol(Symbol64{
		Type:  elf.STT_SECTION,
		Bind:  elf.STB_LOCAL,
		Shndx: w.header.Shnum,
	})
	if err != nil {
		return err
	}
	w.symbols[len(w.symbols)-1].Name = 0

	w.header.Shnum++

	if _, err := w.code.Write(sect.Code); err != nil {
		return err
	}

	pad := int(sect.Addralign) - (len(sect.Code) % int(sect.Addralign))
	w.code.Grow(pad)
	for i := pad; i > 0; i-- {
		w.code.WriteByte(0)
	}

	return writeNullStr(&w.shstrtab, sect.Name)
}

// WriteSymbol writes the given symbol internally in the [Writer].
func (w *Writer) WriteSymbol(symb Symbol64) error {
	w.symbols = append(w.symbols, elf.Sym64{
		Name:  uint32(w.strtab.Len()),
		Info:  byte(symb.Bind<<4) | byte(symb.Type),
		Shndx: symb.Shndx,
		Value: symb.Value,
		Size:  0,
	})

	return writeNullStr(&w.strtab, symb.Name)
}

// Flush compiles the written ELF file in the [Writer] to bytes and writes
// it into the output.
func (w *Writer) Flush() error {
	if err := w.makeSymbolTable(); err != nil {
		return err
	}

	if err := w.writeStringTable(".strtab", &w.strtab); err != nil {
		return err
	}

	if err := w.writeShstrtab(); err != nil {
		return err
	}

	w.finalizeSectionOffsets()

	if err := binary.Write(w.output, binary.LittleEndian, w.header); err != nil {
		return err
	}

	if err := w.writeSectionHeaders(); err != nil {
		return err
	}

	_, err := w.code.WriteTo(w.output)
	return err
}

func (w *Writer) makeSymbolTable() error {
	slices.SortFunc(w.symbols, func(a, b elf.Sym64) int {
		return int(a.Info>>4) - int(b.Info>>4)
	})

	buf := bytes.NewBuffer(make([]byte, 0, Symbol64Size*len(w.symbols)))
	var firstGlobalIx uint32
	for i, symb := range w.symbols {
		if firstGlobalIx == 0 && (symb.Info>>4) == uint8(elf.STB_GLOBAL) {
			firstGlobalIx = uint32(i)
		}

		if err := writeSymbolToBuffer(buf, symb); err != nil {
			return err
		}
	}

	return w.WriteSection(Section64{
		Name:      ".symtab",
		Type:      elf.SHT_SYMTAB,
		Link:      uint32(len(w.sections)) + 1,
		Info:      firstGlobalIx,
		Addralign: 8,
		Entsize:   Symbol64Size,
		Code:      buf.Bytes(),
	})
}

func (w *Writer) writeShstrtab() error {
	w.header.Shstrndx = uint16(len(w.sections))
	if err := w.writeStringTable(".shstrtab", &w.shstrtab); err != nil {
		return err
	}
	w.sections[len(w.sections)-1].Size += 1

	return w.code.WriteByte(0)
}

func (w *Writer) writeStringTable(name string, strings *strings.Builder) error {
	if _, err := w.shstrtab.WriteString(name); err != nil {
		return err
	}

	err := w.WriteSection(Section64{
		Type:      elf.SHT_STRTAB,
		Addralign: 1,
		Code:      []byte(strings.String()),
	})
	if err != nil {
		return err
	}

	w.sections[len(w.sections)-1].Name -= uint32(len(name))

	return nil
}

func (w *Writer) finalizeSectionOffsets() {
	offset := Header64Size + Section64Size*len(w.sections)
	for i := range w.sections[1:] {
		w.sections[i+1].Off += uint64(offset)
	}
}

func (w *Writer) writeSectionHeaders() error {
	sectHdrs := bytes.NewBuffer(make([]byte, 0, Section64Size*len(w.sections)))
	for _, sect := range w.sections {
		if err := writeSectToBuffer(sectHdrs, sect); err != nil {
			return err
		}
	}

	_, err := sectHdrs.WriteTo(w.output)
	return err
}

func writeNullStr(builder *strings.Builder, str string) error {
	if _, err := builder.WriteString(str); err != nil {
		return err
	}

	return builder.WriteByte(0)
}

func writeSectToBuffer(buf *bytes.Buffer, sect elf.Section64) error {
	sectBytes, err := section64ToBytes(sect)
	if err != nil {
		return err
	}

	_, err = buf.Write(sectBytes)
	return err
}

func writeSymbolToBuffer(buf *bytes.Buffer, symb elf.Sym64) error {
	sectBytes, err := sym64ToBytes(symb)
	if err != nil {
		return err
	}

	_, err = buf.Write(sectBytes)
	return err
}

func section64ToBytes(sect elf.Section64) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, Section64Size))
	err := binary.Write(buf, binary.LittleEndian, sect)

	return buf.Bytes(), err
}

func sym64ToBytes(symb elf.Sym64) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, Symbol64Size))
	err := binary.Write(buf, binary.LittleEndian, symb)

	return buf.Bytes(), err
}
