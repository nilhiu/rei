package rasm_test

import (
	"io"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/nilhiu/rei/rasm"
)

func TestCodeGen_Next(t *testing.T) {
	tests := []struct {
		name    string
		rd      io.Reader
		labels  map[string]uint
		want    []byte
		want2   string
		wantErr bool
	}{
		{
			name:   "Generating code for a simple instruction",
			rd:     strings.NewReader("mov rax, 50123"),
			labels: map[string]uint{},
			want: []byte{
				0x48, 0xB8, 0xCB, 0xC3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			want2:   ".text",
			wantErr: false,
		},
		{
			name:   "Generating code for the correct section",
			rd:     strings.NewReader("section .bss\nsection .text\nsection .data\nmov rax, 50123"),
			labels: map[string]uint{},
			want: []byte{
				0x48, 0xB8, 0xCB, 0xC3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			want2:   ".data",
			wantErr: false,
		},
		{
			name:   "Keeps track of labels",
			rd:     strings.NewReader("section .bss\nlabel:\nmov rax, 50123"),
			labels: map[string]uint{"label": 0},
			want: []byte{
				0x48, 0xB8, 0xCB, 0xC3, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			},
			want2:   ".bss",
			wantErr: false,
		},
		{
			name:    "EOF should give nil slice without error",
			rd:      strings.NewReader(""),
			labels:  map[string]uint{},
			want:    nil,
			want2:   ".text",
			wantErr: false,
		},
		{
			name:    "Illegal expression should give an error",
			rd:      strings.NewReader("just_identifier_error"),
			labels:  map[string]uint{},
			want:    nil,
			want2:   ".text",
			wantErr: true,
		},
		{
			name:    "Label redefinition should give an error",
			rd:      strings.NewReader("label:\nlabel:\n"),
			labels:  map[string]uint{"label": 0},
			want:    nil,
			want2:   ".text",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cg := rasm.NewCodeGen(tt.rd)
			got, got2, gotErr := cg.Next()
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Next() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Next() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(tt.labels, cg.Labels()) {
				t.Errorf("cg.Labels() = %v, want %v", cg.Labels(), tt.labels)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
			if got2 != tt.want2 {
				t.Errorf("Next() = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestCodeGen(t *testing.T) {
	prog := `
  mov ebx, 1
  section .data
  mov_code:
    mov eax, 25
  section .bss
  add_code:
    add edx, ebx
  section .text
  _start:
    mov eax, 60
    mov ebx, 0`
	wantSects := []string{".text", ".data", ".bss", ".text", ".text"}
	wantLabels := map[string]uint{
		"mov_code": 5,
		"add_code": 10,
		"_start":   12,
	}
	wantCode := []byte{
		0xbb, 0x01, 0x00, 0x00, 0x00,
		0xb8, 0x19, 0x00, 0x00, 0x00,
		0x01, 0xda,
		0xb8, 0x3c, 0x00, 0x00, 0x00,
		0xbb, 0x00, 0x00, 0x00, 0x00,
	}
	cg := rasm.NewCodeGen(strings.NewReader(prog))

	gotCode := []byte{}
	for _, wantSect := range wantSects {
		bytes, gotSect, err := cg.Next()
		if err != nil {
			t.Fatal("Next() failed with an error: ", err)
		}

		gotCode = append(gotCode, bytes...)

		if wantSect != gotSect {
			t.Errorf("Next() = %v, want %v", gotSect, wantSect)
		}
	}

	if !reflect.DeepEqual(wantLabels, cg.Labels()) {
		t.Errorf("cg.Labels() = %v, want %v", cg.Labels(), wantLabels)
	}

	if !slices.Equal(wantCode, gotCode) {
		t.Errorf("Next() = %x, want %x", gotCode, wantCode)
	}
}
