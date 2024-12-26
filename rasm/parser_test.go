package rasm_test

import (
	"strings"
	"testing"

	"github.com/nilhiu/rei/rasm"
)

func TestIllegalParsing(t *testing.T) {
	lxr := rasm.NewLexer(
		strings.NewReader("section :\nident,\nmov 0 1\n\n,\nmov :"),
	)
	p := rasm.NewParserLexer(lxr)
	for i := 0; i < 5; i++ {
		expr := p.Next()
		if expr.Id != rasm.IllegalExpr {
			t.Fatalf("Illegal parsing let through. Got: %d", expr.Id)
		}
	}
}

func TestSectionParsing(t *testing.T) {
	p := rasm.NewParser(
		strings.NewReader("section .text\n\n\n\nSECTION .DATA"),
	)
	expectedSectionIdents := []string{".text", ".DATA"}

	for i := 0; i < 2; i++ {
		expr := p.Next()
		if expr.Id != rasm.SectionExpr {
			t.Fatalf("Incorrect parsing detected. Expected section expression")
		}
		if expr.Children[0].Raw() != expectedSectionIdents[i] {
			t.Fatalf("Incorrect parsing detected. Expected: %q, got: %q", expectedSectionIdents[i], expr.Children[0].Raw())
		}
	}
}

func TestInstructionParsing(t *testing.T) {
	p := rasm.NewParser(
		strings.NewReader("mov rax, 42\nmov 0xF74182A, 0o6, eip, r15\nmov\n"),
	)
	expectedOpAmt := []int{2, 4, 0}

	for i := 0; i < 3; i++ {
		expr := p.Next()
		if expr.Id != rasm.InstrExpr {
			t.Fatalf("Incorrect parsing detected. Expected: InstrExpr(1), got: %d", expr.Id)
		}
		if len(expr.Children) != expectedOpAmt[i] {
			t.Fatalf(
				"Incorrect parsing detected. Expected: %d operands, got: %d",
				expectedOpAmt[i],
				len(expr.Children),
			)
		}
	}
}

func TestLabelParsing(t *testing.T) {
	p := rasm.NewParser(
		strings.NewReader("_start:\n\n\nlabel1:\nlabel._2_.:\n"),
	)
	expectedLabelNames := []string{"_start", "label1", "label._2_."}

	for i := 0; i < 3; i++ {
		expr := p.Next()
		if expr.Id != rasm.LabelExpr {
			t.Fatalf("Incorrect parsing detected. Expected: LabelExpr(3), got: %d", expr.Id)
		}
		if expr.Root.Raw() != expectedLabelNames[i] {
			t.Fatalf(
				"Incorrect parsing detected. Expected: %q, got: %q",
				expectedLabelNames[i],
				expr.Root.Raw(),
			)
		}
	}
}
