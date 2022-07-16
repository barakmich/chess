package bitflip

func QueenAttacks(occupied uint64, location uint64, rank uint64, file uint64, diag uint64, antidiag uint64) uint64 {
	return queenAttacks(occupied, location, rank, file, diag, antidiag)
}

func BishopRookAttacks(occupied uint64, location uint64, rankOrDiag uint64, fileOrAntiDiag uint64) uint64 {
	return bishopRookAttacks(occupied, location, rankOrDiag, fileOrAntiDiag)
}
