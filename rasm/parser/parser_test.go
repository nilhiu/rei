package parser

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/nilhiu/rei/rasm/lexer"
)

func TestNext(t *testing.T) {
	// Identifier map used for testing
	idMap := lexer.NewSimpleIdentifierMap(map[string]*lexer.IdentToken{
		"nop": lexer.NewIdentToken(lexer.Instruction, 1),
		"mov": lexer.NewIdentToken(lexer.Instruction, 2),
		"eax": lexer.NewIdentToken(lexer.Register, 3),
	})

	tests := []struct {
		name    string
		input   string
		want    Expr
		wantErr error
	}{
		{
			name:  "Should parse section expression",
			input: "section .text",
			want: NewExpr(
				Section,
				lexer.NewToken(lexer.NewPosition(1, 8), lexer.Identifier, ".text"),
			),
			wantErr: nil,
		},
		{
			name:  "Should parse instruction expression without operands",
			input: "nop",
			want: NewExpr(
				Instruction,
				lexer.NewTokenSubID(lexer.NewPosition(1, 0), lexer.Instruction, 1, "nop"),
			),
			wantErr: nil,
		},
		{
			name:  "Should parse instruction expression without operands (newline ending)",
			input: "nop\n",
			want: NewExpr(
				Instruction,
				lexer.NewTokenSubID(lexer.NewPosition(1, 0), lexer.Instruction, 1, "nop"),
			),
			wantErr: nil,
		},
		{
			name:  "Should parse instruction expression with operands",
			input: "mov eax, 512, 0xff, 0o766, ident",
			want: NewExpr(
				Instruction,
				lexer.NewTokenSubID(lexer.NewPosition(1, 0), lexer.Instruction, 2, "mov"),
				NewExpr(
					Leaf,
					lexer.NewTokenSubID(lexer.NewPosition(1, 4), lexer.Register, 3, "eax"),
				),
				NewExpr(Leaf, lexer.NewToken(lexer.NewPosition(1, 9), lexer.Decimal, "512")),
				NewExpr(Leaf, lexer.NewToken(lexer.NewPosition(1, 14), lexer.Hex, "ff")),
				NewExpr(Leaf, lexer.NewToken(lexer.NewPosition(1, 20), lexer.Octal, "766")),
				NewExpr(Leaf, lexer.NewToken(lexer.NewPosition(1, 27), lexer.Identifier, "ident")),
			),
			wantErr: nil,
		},
		{
			name:  "Should parse instruction expression with operands (newline ending)",
			input: "mov 512\n",
			want: NewExpr(
				Instruction,
				lexer.NewTokenSubID(lexer.NewPosition(1, 0), lexer.Instruction, 2, "mov"),
				NewExpr(Leaf, lexer.NewToken(lexer.NewPosition(1, 4), lexer.Decimal, "512")),
			),
			wantErr: nil,
		},
		{
			name:  "Should parse label expression",
			input: "label:",
			want: NewExpr(
				Label,
				lexer.NewToken(lexer.NewPosition(1, 0), lexer.Identifier, "label"),
			),
			wantErr: nil,
		},

		// Erroneous tests
		{
			name:    "Should not parse malformed section expression",
			input:   "section :",
			want:    nil,
			wantErr: ErrNoSectionName,
		},
		{
			name:    "Should not parse malformed instruction expression (expect operand)",
			input:   "mov 512,,",
			want:    nil,
			wantErr: ErrIllegal,
		},
		{
			name:    "Should not parse malformed instruction expression (expect delimiter)",
			input:   "mov 512:",
			want:    nil,
			wantErr: ErrIllegal,
		},
		{
			name:    "Should not parse malformed label expression",
			input:   "label,",
			want:    nil,
			wantErr: ErrUnexpectedIdentifier,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New("test", strings.NewReader(tt.input), idMap)
			p := New(l)

			got, err := p.Next()
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
		})
	}
}
