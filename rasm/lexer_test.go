package rasm_test

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/nilhiu/rei/rasm"
	"github.com/nilhiu/rei/x86"
)

func TestLexer_Next(t *testing.T) {
	pos0 := rasm.Position{1, 0}
	tests := []struct {
		name string
		rd   io.Reader
		want rasm.Token
	}{
		{
			name: "Should lex EOF",
			rd:   strings.NewReader(""),
			want: rasm.NewToken(pos0, rasm.EOF, ""),
		},
		{
			name: "Should lex spaces",
			rd:   strings.NewReader("   "),
			want: rasm.NewToken(rasm.Position{1, 3}, rasm.EOF, ""),
		},
		{
			name: "Should lex random identifier (EOF)",
			rd:   strings.NewReader("_lo_._hi_bit"),
			want: rasm.NewToken(pos0, rasm.Identifier, "_lo_._hi_bit"),
		},
		{
			name: "Should lex random identifier",
			rd:   strings.NewReader("_lo_._hi_bit "),
			want: rasm.NewToken(pos0, rasm.Identifier, "_lo_._hi_bit"),
		},
		{
			name: "Should lex x86 mnemonics",
			rd:   strings.NewReader("mOv"),
			want: rasm.NewToken(pos0, rasm.TokenID(x86.MOV)|rasm.Instruction, "mOv"),
		},
		{
			name: "Should lex x86 register",
			rd:   strings.NewReader("bPl"),
			want: rasm.NewToken(pos0, rasm.TokenID(x86.BPL)|rasm.Register, "bPl"),
		},
		{
			name: "Should lex decimal zero (EOF)",
			want: rasm.NewToken(pos0, rasm.Decimal, "0"),
			rd:   strings.NewReader("0"),
		},
		{
			name: "Should lex decimal zero",
			want: rasm.NewToken(pos0, rasm.Decimal, "0"),
			rd:   strings.NewReader("0 "),
		},
		{
			name: "Should lex decimal numbers",
			want: rasm.NewToken(pos0, rasm.Decimal, "1234567890"),
			rd:   strings.NewReader("1234567890"),
		},
		{
			// This test case may change as keeping the beginning '0' is useless
			name: "Should lex decimal numbers starting with 0",
			rd:   strings.NewReader("0123456789"),
			want: rasm.NewToken(pos0, rasm.Decimal, "0123456789"),
		},
		{
			name: "Should lex decimal number zero",
			rd:   strings.NewReader("0"),
			want: rasm.NewToken(pos0, rasm.Decimal, "0"),
		},
		{
			name: "Should lex decimal numbers multiple zeros",
			rd:   strings.NewReader("00000"),
			want: rasm.NewToken(pos0, rasm.Decimal, "00000"),
		},
		{
			name: "Should lex decimal numbers before letters separately",
			rd:   strings.NewReader("512hello"),
			want: rasm.NewToken(pos0, rasm.Decimal, "512"),
		},
		{
			name: "Should lex hexadecimal numbers (x)",
			rd:   strings.NewReader("0x0123456789AbCdEfGhIjKl"),
			want: rasm.NewToken(pos0, rasm.Hex, "0123456789AbCdEf"),
		},
		{
			name: "Should lex hexadecimal numbers (X)",
			rd:   strings.NewReader("0X0123456789AbCdEfGhIjKl"),
			want: rasm.NewToken(pos0, rasm.Hex, "0123456789AbCdEf"),
		},
		{
			name: "Should lex octal numbers (o)",
			rd:   strings.NewReader("0o0123456789"),
			want: rasm.NewToken(pos0, rasm.Octal, "01234567"),
		},
		{
			name: "Should lex octal numbers (O)",
			rd:   strings.NewReader("0O0123456789"),
			want: rasm.NewToken(pos0, rasm.Octal, "01234567"),
		},
		{
			name: "Should lex section keyword",
			rd:   strings.NewReader("sEcTiOn"),
			want: rasm.NewToken(pos0, rasm.Section, "sEcTiOn"),
		},
		{
			name: "Should lex ','",
			rd:   strings.NewReader(","),
			want: rasm.NewToken(pos0, rasm.Comma, ","),
		},
		{
			name: "Should lex ':'",
			rd:   strings.NewReader(":"),
			want: rasm.NewToken(pos0, rasm.Colon, ":"),
		},
		{
			name: "Should lex newline",
			rd:   strings.NewReader("\n"),
			want: rasm.NewToken(pos0, rasm.Newline, "\\n"),
		},
		{
			name: "Should not lex just the hex prefix (EOF)",
			rd:   strings.NewReader("0x"),
			want: rasm.NewToken(pos0, rasm.Illegal, "hex prefix without logical continuation"),
		},
		{
			name: "Should not lex just the hex prefix",
			rd:   strings.NewReader("0x "),
			want: rasm.NewToken(pos0, rasm.Illegal, "hex prefix without logical continuation"),
		},
		{
			name: "Should not lex just the octal prefix (EOF)",
			rd:   strings.NewReader("0o"),
			want: rasm.NewToken(pos0, rasm.Illegal, "octal prefix without logical continuation"),
		},
		{
			name: "Should not lex just the octal prefix",
			rd:   strings.NewReader("0o "),
			want: rasm.NewToken(pos0, rasm.Illegal, "octal prefix without logical continuation"),
		},
		{
			name: "Should not lex unknown symbols",
			rd:   strings.NewReader("\\"),
			want: rasm.NewToken(pos0, rasm.Illegal, "\\"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := rasm.NewLexer(tt.rd)
			got := l.Next()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLexerPositioning(t *testing.T) {
	str := "\nname mov,\n0xAAFF0 0o1234 0 0 000\n12418\n\n\nsection\n    label: random_name\n19370 0"
	expected := []rasm.Position{
		{Line: 1, Col: 0},
		{Line: 2, Col: 0},
		{Line: 2, Col: 5},
		{Line: 2, Col: 8},
		{Line: 2, Col: 9},
		{Line: 3, Col: 0},
		{Line: 3, Col: 8},
		{Line: 3, Col: 15},
		{Line: 3, Col: 17},
		{Line: 3, Col: 19},
		{Line: 3, Col: 22},
		{Line: 4, Col: 0},
		{Line: 4, Col: 5},
		{Line: 5, Col: 0},
		{Line: 6, Col: 0},
		{Line: 7, Col: 0},
		{Line: 7, Col: 7},
		{Line: 8, Col: 4},
		{Line: 8, Col: 9},
		{Line: 8, Col: 11},
		{Line: 8, Col: 22},
		{Line: 9, Col: 0},
		{Line: 9, Col: 6},
	}
	lxr := rasm.NewLexer(strings.NewReader(str))

	for i := 0; i < 14; i++ {
		tok := lxr.Next()
		if tok.Pos() != expected[i] {
			t.Fatalf(
				`Incorrect lexer positioning detected. Expected: %+v, got: %+v (%q)`,
				expected[i],
				tok.Pos(),
				tok.Raw(),
			)
		}
	}
}
