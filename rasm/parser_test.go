package rasm_test

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/nilhiu/rei/rasm"
	"github.com/nilhiu/rei/x86"
)

func TestParser_Next(t *testing.T) {
	tests := []struct {
		name string
		rd   io.Reader
		want rasm.Expr
	}{
		{
			name: "Should parse section expression",
			rd:   strings.NewReader("\nsection .bss"),
			want: rasm.Expr{
				rasm.SectionExpr,
				rasm.NewToken(rasm.Position{2, 0}, rasm.Section, "section"),
				[]rasm.Token{
					rasm.NewToken(rasm.Position{2, 8}, rasm.Identifier, ".bss"),
				},
			},
		},
		{
			name: "Should parse instruction expression",
			rd:   strings.NewReader("add"),
			want: rasm.Expr{
				rasm.InstrExpr,
				rasm.NewToken(rasm.Position{1, 0}, rasm.TokenID(x86.ADD)|rasm.Instruction, "add"),
				[]rasm.Token{},
			},
		},
		{
			name: "Should parse instruction expression (with operands)",
			rd:   strings.NewReader("mov eax, 512, 0xff, 0o777, some_ident"),
			want: rasm.Expr{
				rasm.InstrExpr,
				rasm.NewToken(rasm.Position{1, 0}, rasm.TokenID(x86.MOV)|rasm.Instruction, "mov"),
				[]rasm.Token{
					rasm.NewToken(rasm.Position{1, 4}, rasm.TokenID(x86.EAX)|rasm.Register, "eax"),
					rasm.NewToken(rasm.Position{1, 9}, rasm.Decimal, "512"),
					rasm.NewToken(rasm.Position{1, 14}, rasm.Hex, "ff"),
					rasm.NewToken(rasm.Position{1, 20}, rasm.Octal, "777"),
					rasm.NewToken(rasm.Position{1, 27}, rasm.Identifier, "some_ident"),
				},
			},
		},
		{
			name: "Should parse label expression",
			rd:   strings.NewReader("label:"),
			want: rasm.Expr{
				rasm.LabelExpr,
				rasm.NewToken(rasm.Position{1, 0}, rasm.Identifier, "label"),
				nil,
			},
		},
		{
			name: "Should parse EOF",
			rd:   strings.NewReader(""),
			want: rasm.Expr{
				rasm.EOFExpr,
				rasm.NewToken(rasm.Position{1, 0}, rasm.EOF, ""),
				nil,
			},
		},
		{
			name: "Should not parse illegal token",
			rd:   strings.NewReader("\\"),
			want: rasm.Expr{
				rasm.IllegalExpr,
				rasm.NewToken(rasm.Position{1, 0}, rasm.Illegal, "\\"),
				nil,
			},
		},
		{
			name: "Should not parse malformed section expression",
			rd:   strings.NewReader("section :"),
			want: rasm.Expr{
				rasm.IllegalExpr,
				rasm.NewToken(rasm.Position{1, 0}, rasm.Section, "section"),
				[]rasm.Token{
					rasm.NewToken(rasm.Position{0, 0}, 0, "expected identifier"),
					rasm.NewToken(rasm.Position{1, 8}, rasm.Colon, ":"),
				},
			},
		},
		{
			name: "Should not parse malformed instruction expression (expect operand)",
			rd:   strings.NewReader("mov 512,,"),
			want: rasm.Expr{
				rasm.IllegalExpr,
				rasm.NewToken(rasm.Position{1, 0}, rasm.TokenID(x86.MOV)|rasm.Instruction, "mov"),
				[]rasm.Token{
					rasm.NewToken(rasm.Position{0, 0}, 0, "expected operand or '\\n'"),
					rasm.NewToken(rasm.Position{1, 4}, rasm.Decimal, "512"),
					rasm.NewToken(rasm.Position{1, 8}, rasm.Comma, ","),
				},
			},
		},
		{
			name: "Should not parse malformed instruction expression (expect delimiter)",
			rd:   strings.NewReader("mov 512:"),
			want: rasm.Expr{
				rasm.IllegalExpr,
				rasm.NewToken(rasm.Position{1, 0}, rasm.TokenID(x86.MOV)|rasm.Instruction, "mov"),
				[]rasm.Token{
					rasm.NewToken(rasm.Position{0, 0}, 0, "expected '\\n' or ','"),
					rasm.NewToken(rasm.Position{1, 4}, rasm.Decimal, "512"),
					rasm.NewToken(rasm.Position{1, 7}, rasm.Colon, ":"),
				},
			},
		},
		{
			name: "Should not parse malformed label expression",
			rd:   strings.NewReader("label,"),
			want: rasm.Expr{
				rasm.IllegalExpr,
				rasm.NewToken(rasm.Position{1, 0}, rasm.Identifier, "label"),
				[]rasm.Token{
					rasm.NewToken(rasm.Position{0, 0}, 0, "expected ':'"),
					rasm.NewToken(rasm.Position{1, 5}, rasm.Comma, ","),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := rasm.NewParser(tt.rd)

			got := p.Next()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}
