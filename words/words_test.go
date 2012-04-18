package words

import (
	"io"
	"os"
	"testing"
)

var notchMem  = []uint16{
	0x7c01, 0x0030, 0x7de1, 0x1000, 0x0020, 0x7803, 0x1000, 0xc00d,
	0x7dc1, 0x001a, 0xa861, 0x7c01, 0x2000, 0x2161, 0x2000, 0x8463,
	0x806d, 0x7dc1, 0x000d, 0x9031, 0x7c10, 0x0018, 0x7dc1, 0x001a,
	0x9037, 0x61c1, 0x7dc1, 0x001a, 0x0000, 0x0000, 0x0000, 0x0000,
}

func TestReadWriter(t *testing.T) {
	file, err := os.Open("../examples/notch.bin")
	if err != nil {
		t.Fatal(err)
	}

	mem := make([]uint16, len(notchMem))
	rw := NewReadWriter(mem)
	_, err = io.Copy(rw, file)

	for i, w := range(mem) {
		if w != notchMem[i] {
			t.Fatalf("Expected %#04x, got %#04x", notchMem[i], w)
		}
	}

	rw.SeekWords(0, 0)

	file, err = os.Create("../examples/write.bin")
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(file, rw)
	if err != nil {
		t.Fatal(err)
	}
}
