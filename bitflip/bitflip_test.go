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

func TestQueenAttacks(t *testing.T) {
	for i := 0; i < 64; i++ {
		occ := bbForSquare(27) | bbForSquare(i)
		exp := queenAttack(occ, i)
		out := queenAttackAVX(occ, i)
		if exp != out {
			t.Errorf("Queen Attack mismatch %d: \ngot %064b\nexp %064b\n", i, out, exp)
		}
	}
}

func TestBishopAttacks(t *testing.T) {
	for i := 0; i < 64; i++ {
		occ := bbForSquare(i) | bbForSquare(27)
		exp := diaAttack(occ, i)
		out := bishopAttackAVX(occ, i)
		if exp != out {
			t.Errorf("Bishop mismatch: \ngot %064b\nexp %064b\n", out, exp)
		}
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

func BenchmarkRookAttackGo(b *testing.B) {
	sq := sqInt(4, 4)
	occ := bbForSquare(sq)
	for n := 0; n < b.N; n++ {
		hvAttack(occ, sq)
	}
}

func BenchmarkQueenAttackAVX(b *testing.B) {
	sq := sqInt(4, 4)
	occ := bbForSquare(sq)
	for n := 0; n < b.N; n++ {
		queenAttackAVX(occ, sq)
	}
}

func BenchmarkBishopAttackAVX(b *testing.B) {
	sq := sqInt(4, 4)
	occ := bbForSquare(sq)
	for n := 0; n < b.N; n++ {
		bishopAttackAVX(occ, sq)
	}
}

func queenAttack(occupied uint64, sq int) uint64 {
	return hvAttack(occupied, sq) | diaAttack(occupied, sq)
}

func queenAttackAVX(occupied uint64, sq int) uint64 {
	pos := bbForSquare(sq)
	diagMask := bbDiagonals[sq]
	adMask := bbAntiDiagonals[sq]
	rankMask := bbRanks[sq&0x7]
	fileMask := bbFiles[(sq >> 3)]
	return QueenAttacks(occupied, pos, rankMask, fileMask, diagMask, adMask)
}

func bishopAttackAVX(occupied uint64, sq int) uint64 {
	pos := bbForSquare(sq)
	diagMask := bbDiagonals[sq]
	adMask := bbAntiDiagonals[sq]
	return BishopRookAttacks(occupied, pos, diagMask, adMask)
}
