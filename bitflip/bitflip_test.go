package bitflip

import (
	"math/bits"
	"testing"
)

func TestByteFlip(t *testing.T) {
	//	in := uint64(0x0123456789abcdef)

	in := uint64(0x58)
	//in = 0x5555555555555555
	out := Reverse64AVX(in)
	exp := bits.Reverse64(in)
	if out != exp {
		t.Errorf("Bytes didn't match, got %x expected %x", out, exp)
	}
}

func TestAttacks(t *testing.T) {
	sq := 27 // d4
	occ := bbForSquare(sq)
	expo, expd := queenAttack(occ, sq)
	outo, outd := queenAttackAVX(occ, sq)
	if expo != outo {
		t.Errorf("Ortho mismatch: \ngot %064b\nexp %064b\n", outo, expo)
	}
	if expd != outd {
		t.Errorf("Diag mismatch: \ngot %064b\nexp %064b\n", outd, expd)
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

func BenchmarkQueenAttackGo(b *testing.B) {
	sq := sqInt(4, 4)
	occ := bbForSquare(sq)
	for n := 0; n < b.N; n++ {
		queenAttack(occ, sq)
	}
}

func BenchmarkBishopAttackGo(b *testing.B) {
	sq := sqInt(4, 4)
	occ := bbForSquare(sq)
	for n := 0; n < b.N; n++ {
		diaAttack(occ, sq)
	}
}

func BenchmarkQueenAttackAVX(b *testing.B) {
	sq := sqInt(4, 4)
	occ := bbForSquare(sq)
	for n := 0; n < b.N; n++ {
		queenAttackAVX(occ, sq)
	}
}

func queenAttack(occupied uint64, sq int) (uint64, uint64) {
	return hvAttack(occupied, sq), diaAttack(occupied, sq)
}

func queenAttackAVX(occupied uint64, sq int) (uint64, uint64) {
	pos := bbForSquare(sq)
	diagMask := bbDiagonals[sq]
	adMask := bbAntiDiagonals[sq]
	rankMask := bbRanks[sq&0x7]
	fileMask := bbFiles[(sq >> 3)]
	masks := [4]uint64{rankMask, fileMask, diagMask, adMask}
	return CalcAttacks(occupied, pos, masks)
}
