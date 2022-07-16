package bitflip

import (
	"math/bits"
	"testing"
)

func TestByteFlip(t *testing.T) {
	in := uint64(0x0123456789abcdef)
	//in = 0x5555555555555555
	out := Reverse64AVX(in)
	exp := bits.Reverse64(in)
	if out != exp {
		t.Errorf("Bytes didn't match, got %x expected %x", out, exp)
	}
}

const benchin = 0x0123456789abcdef

func BenchmarkReverse64Go(b *testing.B) {
	for n := 0; n < b.N; n++ {
		bits.Reverse64(benchin)
	}
}

func BenchmarkReverse64AVX(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Reverse64AVX(benchin)
	}
}
