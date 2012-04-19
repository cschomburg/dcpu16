// Package words provides functions for manipulating word slices.
package words

import (
	"errors"
	"io"
	"fmt"
)

// BytesToWord converts two bytes into a single word
func BytesToWord(a, b byte) (w uint16) {
	return (uint16(a) << 8) | uint16(b)
}

// WordToBytes converts a single word into two bytes.
func WordToBytes(w uint16) (a, b byte) {
	return byte(w >> 8), byte(w & 0xff)
}

// CopyFromBytes copies bytes from src into words of dest.
// Returns the number of bytes written
func CopyFromBytes(dest []uint16, src []byte) int {
	offs := 0
	for i := 0; i < len(src); i += 2 {
		a, b := src[i], byte(0)
		if i < len(src)-1 {
			b = src[i+1]
		}
		if offs >= len(dest) {
			return offs*2
		}
		dest[offs] = BytesToWord(a, b)
		offs++
	}

	return offs*2
}

// CopyToBytes copies words from dest into bytes of src.
// Returns the number of bytes written.
func CopyToBytes(dest []byte, src []uint16) int {
	offs := 0
	for i := 0; i < len(src); i++ {
		if offs >= len(dest) {
			return offs
		}
		dest[offs], dest[offs+1] = WordToBytes(src[i])
		offs += 2
	}
	return offs
}

// A ReadWriter for words that implements Reader and Writer interfaces.
type ReadWriter struct {
	s []uint16
	i int
}

func (rw *ReadWriter) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if rw.i >= len(rw.s) {
		return 0, io.EOF
	}
	n = CopyToBytes(b, rw.s[rw.i:])
	rw.i += n/2
	return
}

func (rw *ReadWriter) Write(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if rw.i >= len(rw.s) {
		return 0, io.EOF
	}
	n = CopyFromBytes(rw.s[rw.i:], b)
	rw.i += n/2
	return
}

func (rw *ReadWriter) SeekWords(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case 0:
		abs = offset
	case 1:
		abs = int64(rw.i) + offset
	case 2:
		abs = int64(len(rw.s)) + offset
	default:
		return 0, errors.New("words: invalid whence")
	}
	if abs < 0 {
		return 0, errors.New("words: negative position")
	}
	if abs >= 1<<31 {
		return 0, errors.New("words: position out of range")
	}
	rw.i = int(abs)
	return abs, nil
}

func (rw *ReadWriter) Words() []uint16 {
	return rw.s
}

// NewReadWriter returns a new Reader reading from w.
func NewReadWriter(w []uint16) * ReadWriter {
	return &ReadWriter{w, 0}
}

// Hexdump displays the word slice in a readable format.
func Hexdump(src []uint16, dest io.Writer) {
	for l := 0; l < (len(src) / 8); l++ {
		lineNull := true
		for c := 0; c < 8; c++ {
			if src[l*8+c] != 0 {
				lineNull = false
				break
			}
		}
		if lineNull {
			continue
		}

		fmt.Fprintf(dest, "0x%04x:    ", l*8)
		for c := 0; c < 8; c++ {
			fmt.Fprintf(dest, "0x%04x ", src[l*8+c])
		}
		fmt.Fprintln(dest)
	}
}
