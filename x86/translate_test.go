package x86

import (
	"slices"
	"testing"

	"github.com/nilhiu/rei/rasm/codegen"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		name    string
		mnem    Mnemonic
		ops     []codegen.Operand
		want    []byte
		wantErr bool
	}{
		{
			name:    "Translate 'add rcx, rax'",
			mnem:    ADD,
			ops:     []codegen.Operand{RCX, RAX},
			want:    []byte{0x48, 0x01, 0xc1},
			wantErr: false,
		},
		{
			name:    "Translate 'add ecx, 0xa4'",
			mnem:    ADD,
			ops:     []codegen.Operand{ECX, Immediate(0xa4)},
			want:    []byte{0x81, 0xc1, 0xa4, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'add ax, 0xa4a1'",
			mnem:    ADD,
			ops:     []codegen.Operand{AX, Immediate(0xa4a1)},
			want:    []byte{0x66, 0x05, 0xa1, 0xa4},
			wantErr: false,
		},
		{
			name:    "Translate 'add rax, 0xa4'",
			mnem:    ADD,
			ops:     []codegen.Operand{RAX, Immediate(0xa4)},
			want:    []byte{0x48, 0x05, 0xa4, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'add rbx, 0x7f' (compressed)",
			mnem:    ADD,
			ops:     []codegen.Operand{RBX, Immediate(0x7f)},
			want:    []byte{0x48, 0x83, 0xc3, 0x7f},
			wantErr: false,
		},
		{
			name:    "Translate 'mov rax, 591'",
			mnem:    MOV,
			ops:     []codegen.Operand{RAX, Immediate(591)},
			want:    []byte{0x48, 0xB8, 0x4F, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'mov ecx, 591'",
			mnem:    MOV,
			ops:     []codegen.Operand{ECX, Immediate(591)},
			want:    []byte{0xb9, 0x4f, 0x02, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'mov r15w, r15w'",
			mnem:    MOV,
			ops:     []codegen.Operand{R15W, R15W},
			want:    []byte{0x66, 0x45, 0x89, 0xff},
			wantErr: false,
		},
		{
			name: "Translate 'mov eax, [rbx]'",
			mnem: MOV,
			ops: []codegen.Operand{EAX, codegen.SIBAddressing{
				Scale:        1,
				Index:        NilReg,
				Base:         uint32(RBX),
				Displacement: 0,
			}},
			want:    []byte{0x8b, 0x03},
			wantErr: false,
		},
		{
			name: "Translate 'mov eax, [rbx+rax]'",
			mnem: MOV,
			ops: []codegen.Operand{EAX, codegen.SIBAddressing{
				Scale:        1,
				Index:        uint32(RAX),
				Base:         uint32(RBX),
				Displacement: 0,
			}},
			want:    []byte{0x8b, 0x04, 0x03},
			wantErr: false,
		},
		{
			name: "Translate 'mov eax, [rbx+0x7fffffff]'",
			mnem: MOV,
			ops: []codegen.Operand{EAX, codegen.SIBAddressing{
				Scale:        1,
				Index:        NilReg,
				Base:         uint32(RBX),
				Displacement: 0x7fffffff,
			}},
			want:    []byte{0x8b, 0x83, 0xff, 0xff, 0xff, 0x7f},
			wantErr: false,
		},
		{
			name: "Translate 'mov eax, [rbx+rax+0xff]'",
			mnem: MOV,
			ops: []codegen.Operand{EAX, codegen.SIBAddressing{
				Scale:        1,
				Index:        uint32(RAX),
				Base:         uint32(RBX),
				Displacement: 0xff,
			}},
			want:    []byte{0x8b, 0x84, 0x03, 0xff, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name: "Translate 'mov eax, [rbx+2*rax+0xff]'",
			mnem: MOV,
			ops: []codegen.Operand{EAX, codegen.SIBAddressing{
				Scale:        2,
				Index:        uint32(RAX),
				Base:         uint32(RBX),
				Displacement: 0xff,
			}},
			want:    []byte{0x8b, 0x84, 0x43, 0xff, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name: "Translate 'mov eax, [rbx+0x42]' (byte compression)",
			mnem: MOV,
			ops: []codegen.Operand{EAX, codegen.SIBAddressing{
				Scale:        1,
				Index:        NilReg,
				Base:         uint32(RBX),
				Displacement: 0x42,
			}},
			want:    []byte{0x8b, 0x43, 0x42},
			wantErr: false,
		},

		// Erroneous tests
		{
			name:    "Translate 'mov r15b, ah' should error",
			mnem:    MOV,
			ops:     []codegen.Operand{R15B, AH},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Translate 'mov r10, ax' should error",
			mnem:    MOV,
			ops:     []codegen.Operand{R10, AX},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Should give an error on unknown mnemonic",
			mnem:    Mnemonic(0xdeadbeef),
			ops:     []codegen.Operand{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Should give an error on unsupported operands for mnemonic",
			mnem:    MOV,
			ops:     []codegen.Operand{},
			want:    nil,
			wantErr: true,
		},
	}

	translator := Translator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := translator.Translate(uint32(tt.mnem), tt.ops...)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Translate() failed: %v", gotErr)
				}

				return
			}

			if tt.wantErr {
				t.Fatal("Translate() succeeded unexpectedly")
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}
