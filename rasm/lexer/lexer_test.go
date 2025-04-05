package lexer

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

var idMap = NewSimpleIdentifierMap(map[string]*IdentToken{
	"mov": {
		tID:   Instruction,
		subID: 1,
	},
})

func TestNext(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Token
		wantErr error
	}{
		{
			name:    "Should lex random identifier (EOF)",
			input:   "_lo_._hi_bit",
			want:    NewToken(NewPosition(1, 0), Identifier, "_lo_._hi_bit"),
			wantErr: nil,
		},
		{
			name:    "Should lex random identifier",
			input:   "_lo_._hi_bit ",
			want:    NewToken(NewPosition(1, 0), Identifier, "_lo_._hi_bit"),
			wantErr: nil,
		},
		{
			name:    "Should lex decimal zero (EOF)",
			input:   "0",
			want:    NewToken(NewPosition(1, 0), Decimal, "0"),
			wantErr: nil,
		},
		{
			name:    "Should lex decimal zero",
			input:   "0 ",
			want:    NewToken(NewPosition(1, 0), Decimal, "0"),
			wantErr: nil,
		},
		{
			name:    "Should lex decimal numbers",
			input:   "1234567890",
			want:    NewToken(NewPosition(1, 0), Decimal, "1234567890"),
			wantErr: nil,
		},
		{
			name:    "Should lex decimal numbers starting with 0",
			input:   "0123456789",
			want:    NewToken(NewPosition(1, 0), Decimal, "0123456789"),
			wantErr: nil,
		},
		{
			name:    "Should lex decimal numbers multiple zeros",
			input:   "00000",
			want:    NewToken(NewPosition(1, 0), Decimal, "00000"),
			wantErr: nil,
		},
		{
			name:    "Should lex decimal numbers before letters separately",
			input:   "512hello",
			want:    NewToken(NewPosition(1, 0), Decimal, "512"),
			wantErr: nil,
		},
		{
			name:    "Should lex hexadecimal numbers (x)",
			input:   "0x0123456789AbCdEfGhIjKl",
			want:    NewToken(NewPosition(1, 0), Hex, "0123456789AbCdEf"),
			wantErr: nil,
		},
		{
			name:    "Should lex hexadecimal numbers (X)",
			input:   "0X0123456789aBcDeFgHiJkL",
			want:    NewToken(NewPosition(1, 0), Hex, "0123456789aBcDeF"),
			wantErr: nil,
		},
		{
			name:    "Should lex octal numbers (o)",
			input:   "0o0123456789",
			want:    NewToken(NewPosition(1, 0), Octal, "01234567"),
			wantErr: nil,
		},
		{
			name:    "Should lex octal numbers (O)",
			input:   "0O0123456789",
			want:    NewToken(NewPosition(1, 0), Octal, "01234567"),
			wantErr: nil,
		},
		{
			name:    "Should lex section keyword",
			input:   "sEcTiOn",
			want:    NewToken(NewPosition(1, 0), Section, "sEcTiOn"),
			wantErr: nil,
		},
		{
			name:    "Should lex ','",
			input:   ",",
			want:    NewToken(NewPosition(1, 0), Comma, ","),
			wantErr: nil,
		},
		{
			name:    "Should lex ':'",
			input:   ":",
			want:    NewToken(NewPosition(1, 0), Colon, ":"),
			wantErr: nil,
		},
		{
			name:    "Should lex newline",
			input:   "\n",
			want:    NewToken(NewPosition(1, 0), Newline, "\\n"),
			wantErr: nil,
		},
		{
			name:    "Should lex identifiers in idMaps",
			input:   "mov",
			want:    NewTokenSubID(NewPosition(1, 0), Instruction, 1, "mov"),
			wantErr: nil,
		},

		// Erroneous tests
		{
			name:    "Should give EOF error",
			input:   "",
			want:    nil,
			wantErr: ErrEOF,
		},
		{
			name:    "Should give EOF error for whitespace input",
			input:   "   \t",
			want:    nil,
			wantErr: ErrEOF,
		},
		{
			name:    "Should not lex just the hex prefix (EOF)",
			input:   "0x",
			want:    nil,
			wantErr: ErrEmptyHexPrefix,
		},
		{
			name:    "Should not lex just the hex prefix",
			input:   "0x ",
			want:    nil,
			wantErr: ErrEmptyHexPrefix,
		},
		{
			name:    "Should not lex just the octal prefix (EOF)",
			input:   "0o",
			want:    nil,
			wantErr: ErrEmptyOctalPrefix,
		},
		{
			name:    "Should not lex just the octal prefix",
			input:   "0o ",
			want:    nil,
			wantErr: ErrEmptyOctalPrefix,
		},
		{
			name:    "Should not lex unknown symbols",
			input:   "\\",
			want:    nil,
			wantErr: ErrIllegal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("test", strings.NewReader(tt.input), idMap)

			got, err := l.Next()
			if err != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("err = %v, want %v", err, tt.wantErr)
				}
			} else if tt.wantErr != nil {
				t.Error("expected error, got nil")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Next() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLexerPositioning(t *testing.T) {
	str := "\nname mov,\n0xAAFF0 0o1234 0 0 000\n12418\n\n\nsection\n    label: random_name\n19370 0"
	expected := []Position{
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
	lxr := New("test", strings.NewReader(str))

	for i := 0; i < 14; i++ {
		tok, err := lxr.Next()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if tok.Pos() != expected[i] {
			t.Fatalf(
				`Incorrect positioning detected. Expected: %+v, got: %+v (%q)`,
				expected[i],
				tok.Pos(),
				tok.Raw(),
			)
		}
	}
}
