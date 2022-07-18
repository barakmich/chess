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
type Move uint64

type testMove struct {
	piece Piece
	s1    Square
	s2    Square
	promo PromoType
	tags  MoveTag
}

func (tm *testMove) toMove() Move {
	m := NewMove(tm.s1, tm.s2, tm.promo, tm.piece)
	m = m.setTags(tm.tags)
	return m
}

func NewMove(s1 Square, s2 Square, promo PromoType, piece ...Piece) Move {
	var m Move
	m = m.setS1(s1)
	m = m.setS2(s2)
	m = m.setPromo(promo)
	m = m.setPiece(NoPiece)
	if len(piece) != 0 {
		m = m.setPiece(piece[0])
	}
	return m
}

// String returns a string useful for debugging.  String doesn't return
// algebraic notation.
func (m Move) String() string {
	return fmt.Sprintf("%s%s%s", m.S1().String(), m.S2().String(), m.Promo().PieceType().String())
}

const moveS1Mask = 0xFF
const moveS2Mask = 0xFF00
const movePieceMask = 0xFF0000

// S1 returns the origin square of the move.
func (m Move) S1() Square {
	return Square(m & moveS1Mask)
}

func (m Move) setS1(sq Square) Move {
	return m.setField(moveS1Mask, 0, uint64(sq))
}

// S2 returns the destination square of the move.
func (m Move) S2() Square {
	return Square((m & moveS2Mask) >> 8)
}

func (m Move) setS2(sq Square) Move {
	return m.setField(moveS2Mask, 8, uint64(sq))
}

func (m Move) setPiece(p Piece) Move {
	return m.setField(movePieceMask, 16, uint64(p))
}

func (m Move) piece() Piece {
	return Piece((m & movePieceMask) >> 16)
}

const movePromoMask = 0xFF000000

// Promo returns promotion piece type of the move.
func (m Move) Promo() PromoType {
	promotype := (m & movePromoMask) >> 24
	return PromoType(promotype)
}

func (m Move) setPromoPiece(typ PieceType) Move {
	p := promoFromPieceType(typ)
	return m.setPromo(p)
}

func (m Move) setPromo(typ PromoType) Move {
	return m.setField(movePromoMask, 24, uint64(typ))
}

func (m Move) setField(mask uint64, offset int, value uint64) Move {
	tmp := uint64(m) & ^mask
	return Move(tmp | mask&(value<<offset))
}

func (m Move) Eq(other Move) bool {
	toCompare := Move(movePromoMask | moveS1Mask | moveS2Mask)
	return m&toCompare == other&toCompare
}

const moveTagMask = 0xFFFF00000000

// HasTag returns true if the move contains the MoveTag given.
func (m Move) HasTag(tag MoveTag) bool {
	return (tag & m.tags()) > 0
}

func (m Move) addTag(tag MoveTag) Move {
	return m.setTags(m.tags() | tag)
}

func (m Move) tags() MoveTag {
	return MoveTag((m & moveTagMask) >> 32)
}

func (m Move) setTags(tag MoveTag) Move {
	return m.setField(moveTagMask, 32, uint64(tag))
}
