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
;Notch's examples
;Should compile fine, by default

;Try some basic stuff
		SET A, 0x30				; 7c01 0030
		SET [0x1000], 0x20		; 7de1 1000 0020
		SUB A, [0x1000]			; 7803 1000
		IFN A, 0x10				; c00d 
		SET PC, crash         	; 7dc1 001a [*]
                      
;Do a loopy thing
		SET I, 10               ; a861
		SET A, 0x2000           ; 7c01 2000
:loop	SET [0x2000+I], [A]		; 2161 2000
		SUB I, 1				; 8463
		IFN I, 0				; 806d
		SET PC, loop			; 7dc1 000d [*]
        
;Call a subroutine
		SET X, 0x4				; 9031
		JSR testsub				; 7c10 0018 [*]
		SET PC, crash			; 7dc1 001a [*]
        
:testsub
		SHL X, 4				; 9037
		SET PC, POP				; 61c1
                        
								
;Hang forever. X should now be 0x40 if everything went right.
:crash	SET PC, crash			; 7dc1 001a [*]
	`))

	if err != nil {
		t.Fatal(err)
	}
	expect(t, gen, []uint16{
		0x7c01, 0x0030, 0x7de1, 0x1000, 0x0020, 0x7803, 0x1000, 0xc00d,
		0x7dc1, 0x001a, 0xa861, 0x7c01, 0x2000, 0x2161, 0x2000, 0x8463,
		0x806d, 0x7dc1, 0x000d, 0x9031, 0x7c10, 0x0018, 0x7dc1, 0x001a,
		0x9037, 0x61c1, 0x7dc1, 0x001a,
	})
}
