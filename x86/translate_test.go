package x86_test

import (
	"slices"
	"testing"

	"github.com/nilhiu/rei/x86"
)

func TestTranslate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		mnem    x86.Mnemonic
		ops     []x86.Operand
		want    []byte
		wantErr bool
	}{
		{
			name:    "Translate 'add rcx, rax'",
			mnem:    x86.ADD,
			ops:     []x86.Operand{x86.RCX, x86.RAX},
			want:    []byte{0x48, 0x01, 0xc1},
			wantErr: false,
		},
		{
			name:    "Translate 'add ecx, 0xa4'",
			mnem:    x86.ADD,
			ops:     []x86.Operand{x86.ECX, x86.Immediate(0xa4)},
			want:    []byte{0x81, 0xc1, 0xa4, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'add ax, 0xa4a1'",
			mnem:    x86.ADD,
			ops:     []x86.Operand{x86.AX, x86.Immediate(0xa4a1)},
			want:    []byte{0x66, 0x05, 0xa1, 0xa4},
			wantErr: false,
		},
		{
			name:    "Translate 'add rax, 0xa4'",
			mnem:    x86.ADD,
			ops:     []x86.Operand{x86.RAX, x86.Immediate(0xa4)},
			want:    []byte{0x48, 0x05, 0xa4, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'add rbx, 0x7f' (compressed)",
			mnem:    x86.ADD,
			ops:     []x86.Operand{x86.RBX, x86.Immediate(0x7f)},
			want:    []byte{0x48, 0x83, 0xc3, 0x7f},
			wantErr: false,
		},
		{
			name:    "Translate 'mov rax, 591'",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.RAX, x86.Immediate(591)},
			want:    []byte{0x48, 0xB8, 0x4F, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'mov ecx, 591'",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.ECX, x86.Immediate(591)},
			want:    []byte{0xb9, 0x4f, 0x02, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'mov r15w, r15w'",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.R15W, x86.R15W},
			want:    []byte{0x66, 0x45, 0x89, 0xff},
			wantErr: false,
		},
		{
			name:    "Translate 'mov eax, [rbx]'",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.EAX, x86.Address{1, x86.NilReg, x86.RBX, 0}},
			want:    []byte{0x8b, 0x03},
			wantErr: false,
		},
		{
			name:    "Translate 'mov eax, [rbx+rax]'",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.EAX, x86.Address{1, x86.RAX, x86.RBX, 0}},
			want:    []byte{0x8b, 0x04, 0x03},
			wantErr: false,
		},
		{
			name:    "Translate 'mov eax, [rbx+0x7fffffff]'",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.EAX, x86.Address{1, x86.NilReg, x86.RBX, 0x7fffffff}},
			want:    []byte{0x8b, 0x83, 0xff, 0xff, 0xff, 0x7f},
			wantErr: false,
		},
		{
			name:    "Translate 'mov eax, [rbx+rax+0xff]'",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.EAX, x86.Address{1, x86.RAX, x86.RBX, 0xff}},
			want:    []byte{0x8b, 0x84, 0x03, 0xff, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'mov eax, [rbx+2*rax+0xff]'",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.EAX, x86.Address{2, x86.RAX, x86.RBX, 0xff}},
			want:    []byte{0x8b, 0x84, 0x43, 0xff, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'mov eax, []'",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.EAX, x86.Address{}},
			want:    []byte{0x8b, 0x04, 0x25, 0x00, 0x00, 0x00, 0x00},
			wantErr: false,
		},
		{
			name:    "Translate 'mov r15b, ah' should error",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.R15B, x86.AH},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Translate 'mov r10, ax' should error",
			mnem:    x86.MOV,
			ops:     []x86.Operand{x86.R10, x86.AX},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Should give an error on unknown mnemonic",
			mnem:    x86.Mnemonic(0xdeadbeef),
			ops:     []x86.Operand{},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Should give an error on unsupported operands for mnemonic",
			mnem:    x86.MOV,
			ops:     []x86.Operand{},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := x86.Translate(tt.mnem, tt.ops...)
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
