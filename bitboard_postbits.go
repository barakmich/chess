package chess

import "math/bits"

// Reverse returns a bitboard where the bit order is reversed.
func (b bitboard) Reverse() bitboard {
	return bitboard(bits.Reverse64(uint64(b)))
}

// Occupied returns true if the square's bitboard position is 1.
func (b bitboard) Occupied(sq Square) bool {
	return (uint64(b) & (0b1 << int(sq))) != 0
}
