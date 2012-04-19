package assembler

import (
	"testing"
)

func expect(t *testing.T, got, exp []uint16) {
	if len(got) != len(exp) {
		t.Fatalf("length differs: expected %#04x, got %#04x", len(exp), len(got))
	}
	for i, v := range(exp) {
		if i >= len(got) {
		} else if got[i] != v {
			t.Fatalf("at %#04x: expected %#04x, got %#04x", i, v, got[i])
		}
	}
}

func TestBasic(t *testing.T) {
	gen, err := Assemble([]byte(`
;Try some basic stuff
		SET A, 0x30				; 7c01 0030
		SET [0x1000], 0x20		; 7de1 1000 0020
		SUB A, [0x1000]			; 7803 1000
		IFN A, 0x10				; c00d 
		SET I, 10               ; a861
		SET [0x2000+I], [A]		; 2161 2000
	`))
	if err != nil {
		t.Fatal(err)
	}
	expect(t, gen, []uint16{
		0x7c01, 0x0030,
		0x7de1, 0x1000, 0x0020,
		0x7803, 0x1000,
		0xc00d, 0xa861,
		0x2161, 0x2000,
	})
}
