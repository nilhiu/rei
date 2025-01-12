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
	Header64Size  = 64
	Section64Size = 64
	Symbol64Size  = 24
)

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

type Header64 struct {
	Endian  elf.Data
	ABI     elf.OSABI
	Machine elf.Machine
	Flags   uint32
}

type Section64 struct {
	Name      string
	Type      elf.SectionType
	Flags     elf.SectionFlag
	Addr      uint64
	Link      uint32
	Info      uint32
	Addralign uint64
	Entsize   uint64
	Code      []byte
}

type Symbol64 struct {
	Name  string
	Type  elf.SymType
	Bind  elf.SymBind
	Shndx uint16
	Value uint64
}

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

	if _, err := w.shstrtab.WriteString(sect.Name); err != nil {
		return err
	}

	return w.shstrtab.WriteByte(0)
}

func (w *Writer) WriteSymbol(symb Symbol64) error {
	w.symbols = append(w.symbols, elf.Sym64{
		Name:  uint32(w.strtab.Len()),
		Info:  byte(symb.Bind<<4) | byte(symb.Type),
		Shndx: symb.Shndx,
		Value: symb.Value,
		Size:  0,
	})

	if _, err := w.strtab.WriteString(symb.Name); err != nil {
		return err
	}

	return w.strtab.WriteByte(0)
}

func (w *Writer) Flush() error {
	if err := w.makeSymbolTable(); err != nil {
		return err
	}

	if err := w.writeStrtab(); err != nil {
		return err
	}

	if err := w.writeShstrtab(); err != nil {
		return err
	}

	w.finalizeSectionOffsets()

	if err := w.writeHeader(); err != nil {
		return err
	}

	if err := w.writeSectionHeaders(); err != nil {
		return err
	}

	return w.writeCode()
}

func (w *Writer) makeSymbolTable() error {
	slices.SortFunc(w.symbols, func(a, b elf.Sym64) int {
		return int(a.Info>>4) - int(b.Info>>4)
	})

	buf := bytes.NewBuffer([]byte{})
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

// func (w *Writer) writeSectionsToSymTab() error {
//   for sect : range w.sections {
//
//   }
// }

func (w *Writer) writeSectionStr(str string) error {
	if _, err := w.shstrtab.WriteString(str); err != nil {
		return err
	}

	return w.shstrtab.WriteByte(0)
}

func (w *Writer) writeHeader() error {
	return binary.Write(w.output, binary.LittleEndian, w.header)
}

func (w *Writer) writeShstrtab() error {
	w.header.Shstrndx = uint16(len(w.sections))

	if err := w.writeSectionStr(".shstrtab"); err != nil {
		return err
	}

	err := w.WriteSection(Section64{
		Type:      elf.SHT_STRTAB,
		Addralign: 1,
		Code:      []byte(w.shstrtab.String()),
	})
	if err != nil {
		return err
	}

	w.sections[len(w.sections)-1].Name -= 10

	return nil
}

func (w *Writer) writeStrtab() error {
	if err := w.writeSectionStr(".strtab"); err != nil {
		return err
	}

	err := w.WriteSection(Section64{
		Type:      elf.SHT_STRTAB,
		Addralign: 1,
		Code:      []byte(w.strtab.String()),
	})
	if err != nil {
		return err
	}

	w.sections[len(w.sections)-1].Name -= 8

	return nil
}

func (w *Writer) finalizeSectionOffsets() {
	offset := Header64Size + Section64Size*len(w.sections)
	for i := range w.sections[1:] {
		w.sections[i+1].Off += uint64(offset)
	}
}

func (w *Writer) writeSectionHeaders() error {
	sectHdrs := bytes.NewBuffer([]byte{})
	for _, sect := range w.sections {
		if err := writeSectToBuffer(sectHdrs, sect); err != nil {
			return err
		}
	}

	_, err := sectHdrs.WriteTo(w.output)
	return err
}

func (w *Writer) writeCode() error {
	_, err := w.code.WriteTo(w.output)
	return err
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
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.LittleEndian, sect)

	return buf.Bytes(), err
}

func sym64ToBytes(symb elf.Sym64) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.LittleEndian, symb)

	return buf.Bytes(), err
}
