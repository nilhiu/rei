package codegen

import (
	"encoding/binary"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nilhiu/rei/rasm/lexer"
	"github.com/nilhiu/rei/rasm/parser"
)

type mockTranslator struct{}

func (t *mockTranslator) IDMap() lexer.IdentifierMap {
	return lexer.NewSimpleIdentifierMap(map[string]*lexer.IdentToken{
		"nop": lexer.NewIdentToken(lexer.Instruction, 0),
		"mov": lexer.NewIdentToken(lexer.Instruction, 1),
		"eax": lexer.NewIdentToken(lexer.Register, 1),
	})
}

func (t *mockTranslator) Translate(instrID uint32, ops ...Operand) ([]byte, error) {
	switch instrID {
	case 0:
		return []byte{0x90}, nil
	case 1:
		return t.translateMov(ops)
	default:
		return nil, errors.New("unknown instruction ID")
	}
}

func (t *mockTranslator) translateMov(ops []Operand) ([]byte, error) {
	if len(ops) != 2 {
		return nil, errors.New("expected 2 operands")
	}

	bytes := []byte{0xb8}

	if ops[0].Type() != OpTypeRegister {
		return nil, errors.New("expected first operand of kind register")
	}

	bytes = append(bytes, byte(ops[0].Value().(uint32)))

	if ops[1].Type() != OpTypeImmediate {
		return nil, errors.New("expected second operand of kind immediate")
	}

	return binary.LittleEndian.AppendUint64(bytes, ops[1].Value().(uint64)), nil
}

func TestCodeGen_Next(t *testing.T) {
	translator := &mockTranslator{}

	tests := []struct {
		name       string
		input      string
		want       Code
		wantLabels map[string]LabelInfo
		wantErr    error
	}{
		{
			name:       "Should generate code for a no operand instruction",
			input:      "nop",
			want:       NewCode(".text", []byte{0x90}),
			wantLabels: map[string]LabelInfo{},
			wantErr:    nil,
		},
		{
			name:       "Should generate code for the correct section",
			input:      "section .bss\nsection .text\nsection .data\nnop",
			want:       NewCode(".data", []byte{0x90}),
			wantLabels: map[string]LabelInfo{},
			wantErr:    nil,
		},
		{
			name:  "Should correctly keep track of labels",
			input: "section .bss\nlabel:\nnop",
			want:  NewCode(".bss", []byte{0x90}),
			wantLabels: map[string]LabelInfo{
				"label": {
					Section: ".bss",
					Offset:  0,
				},
			},
			wantErr: nil,
		},
		{
			name:  "Should generate code for a multi-operand instruction",
			input: "mov eax, 512",
			want: NewCode(
				".text",
				[]byte{0xb8, 0x01, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			),
			wantLabels: map[string]LabelInfo{},
			wantErr:    nil,
		},

		// Erroneous tests
		{
			name:       "EOF should return error lexer.ErrEOF",
			input:      "",
			want:       nil,
			wantLabels: map[string]LabelInfo{},
			wantErr:    lexer.ErrEOF,
		},
		{
			name:  "Label reintroduction should give an error",
			input: "label:\nlabel:\n",
			want:  nil,
			wantLabels: map[string]LabelInfo{
				"label": {
					Section: ".text",
					Offset:  0,
				},
			},
			wantErr: ErrLabelExists,
		},
		{
			name:       "Translator errors should be reported soundly (mismatched operand count)",
			input:      "mov eax",
			want:       nil,
			wantLabels: map[string]LabelInfo{},
			wantErr:    ErrTranslator,
		},
		{
			name:       "Translator errors should be reported soundly (mismatched operand type)",
			input:      "mov 512, eax",
			want:       nil,
			wantLabels: map[string]LabelInfo{},
			wantErr:    ErrTranslator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test", strings.NewReader(tt.input), translator.IDMap())
			p := parser.New(l)
			cg := New(p, translator)

			got, err := cg.Next()
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
			} else if tt.wantErr != nil {
				t.Errorf("err = nil, want %v", tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}

			if !reflect.DeepEqual(cg.Labels(), tt.wantLabels) {
				t.Errorf("Labels() = %v, want %v", cg.Labels(), tt.wantLabels)
			}
		})
	}
}
