package relf

import (
	"bytes"
	"debug/elf"
	"encoding/binary"
	"io"
	"strings"
)

const (
	Header64Size  = 64
	Section64Size = 64
)

type Writer struct {
	header   elf.Header64
	sections []elf.Section64
	shndx    map[string]uint16
	code     bytes.Buffer
	shstrtab strings.Builder
	output   io.Writer
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

func NewWriter(hdr Header64, writer io.Writer) *Writer {
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
		shndx:    map[string]uint16{"": 0},
		shstrtab: strings.Builder{},
		output:   writer,
	}

	if err := w.shstrtab.WriteByte(0); err != nil {
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
	w.header.Shnum++

	if _, err := w.code.Write(sect.Code); err != nil {
		return err
	}

	if _, err := w.shstrtab.WriteString(sect.Name); err != nil {
		return err
	}

	return w.shstrtab.WriteByte(0)
}

func (w *Writer) Flush() error {
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

func section64ToBytes(sect elf.Section64) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.LittleEndian, sect)

	return buf.Bytes(), err
}
