package chess

import "fmt"

// A MoveTag represents a notable consequence of a move.
type MoveTag uint16

const (
	// KingSideCastle indicates that the move is a king side castle.
	KingSideCastle MoveTag = 1 << iota
	// QueenSideCastle indicates that the move is a queen side castle.
	QueenSideCastle
	// Capture indicates that the move captures a piece.
	Capture
	// EnPassant indicates that the move captures via en passant.
	EnPassant
	// Check indicates that the move puts the opposing player in check.
	Check
	// inCheck indicates that the move puts the moving player in check and
	// is therefore invalid.
	inCheck
	// IsCheckmate indicates that the move puts the opposing player in checkmate.
	IsCheckmate
)

// A Move is the movement of a piece from one square to another.
type Move struct {
	piece Piece
	s1    Square
	s2    Square
	promo PromoType
	tags  MoveTag
}

// String returns a string useful for debugging.  String doesn't return
// algebraic notation.
func (m *Move) String() string {
	return fmt.Sprintf("%s%s%s", m.s1.String(), m.s2.String(), m.promo.PieceType().String())

}

// S1 returns the origin square of the move.
func (m *Move) S1() Square {
	return m.s1
}

// S2 returns the destination square of the move.
func (m *Move) S2() Square {
	return m.s2
}

// Promo returns promotion piece type of the move.
func (m *Move) Promo() PieceType {
	return m.promo.PieceType()
}

func (m *Move) Eq(other *Move) bool {
	if m.s1 != other.s1 {
		return false
	}
	if m.s2 != other.s2 {
		return false
	}
	if m.promo != other.promo {
		return false
	}
	return true
}

func (m *Move) copy() *Move {
	return &Move{
		piece: m.piece,
		s1:    m.s1,
		s2:    m.s2,
		promo: m.promo,
		tags:  m.tags,
	}
}

// HasTag returns true if the move contains the MoveTag given.
func (m *Move) HasTag(tag MoveTag) bool {
	return (tag & m.tags) > 0
}

func (m *Move) addTag(tag MoveTag) {
	m.tags = m.tags | tag
}
