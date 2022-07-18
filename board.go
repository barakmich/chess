package chess

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math/bits"
	"strconv"
	"strings"
)

const darkSquares uint64 = 0xAA55AA55AA55AA55
const lightSquares uint64 = 0x55AA55AA55AA55AA

// A Board represents a chess board and its relationship between squares and pieces.
type Board struct {
	array         [22]bitboard
	whiteKingSq   Square
	blackKingSq   Square
	occupiedCache bitboard
}

// NewBoard returns a board from a square to piece mapping.
func NewBoard(m map[Square]Piece) *Board {
	b := &Board{}
	for _, p1 := range allPieces {
		bm := make(map[Square]bool)
		for sq, p2 := range m {
			if p1 == p2 {
				bm[sq] = true
			}
		}
		bb := newBitboard(bm)
		b.setBBForPiece(p1, bb)
	}
	b.updateKings(nil)
	return b
}

// SquareMap returns a mapping of squares to pieces.  A square is only added to the map if it is occupied.
func (b *Board) SquareMap() map[Square]Piece {
	m := map[Square]Piece{}
	for sq := 0; sq < numOfSquaresInBoard; sq++ {
		p := b.Piece(Square(sq))
		if p != NoPiece {
			m[Square(sq)] = p
		}
	}
	return m
}

// Rotate rotates the board 90 degrees clockwise.
func (b *Board) Rotate() *Board {
	return b.Flip(UpDown).Transpose()
}

// FlipDirection is the direction for the Board.Flip method
type FlipDirection int

const (
	// UpDown flips the board's rank values
	UpDown FlipDirection = iota
	// LeftRight flips the board's file values
	LeftRight
)

// Flip flips the board over the vertical or hoizontal
// center line.
func (b *Board) Flip(fd FlipDirection) *Board {
	m := map[Square]Piece{}
	for sq := 0; sq < numOfSquaresInBoard; sq++ {
		var mv Square
		switch fd {
		case UpDown:
			file := Square(sq).File()
			rank := Rank(7 - Square(sq).Rank())
			mv = NewSquare(file, rank)
		case LeftRight:
			file := File(7 - Square(sq).File())
			rank := Square(sq).Rank()
			mv = NewSquare(file, rank)
		}
		m[mv] = b.Piece(Square(sq))
	}
	return NewBoard(m)
}

// Transpose flips the board over the A8 to H1 diagonal.
func (b *Board) Transpose() *Board {
	m := map[Square]Piece{}
	for sq := 0; sq < numOfSquaresInBoard; sq++ {
		file := File(7 - Square(sq).Rank())
		rank := Rank(7 - Square(sq).File())
		mv := NewSquare(file, rank)
		m[mv] = b.Piece(Square(sq))
	}
	return NewBoard(m)
}

// Draw returns visual representation of the board useful for debugging.
func (b *Board) Draw() string {
	s := "\n A B C D E F G H\n"
	for r := 7; r >= 0; r-- {
		s += Rank(r).String()
		for f := 0; f < numOfSquaresInRow; f++ {
			p := b.Piece(NewSquare(File(f), Rank(r)))
			if p == NoPiece {
				s += "-"
			} else {
				s += p.String()
			}
			s += " "
		}
		s += "\n"
	}
	return s
}

// String implements the fmt.Stringer interface and returns
// a string in the FEN board format: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
func (b *Board) String() string {
	expandedFen := []byte{
		'1', '1', '1', '1', '1', '1', '1', '1', '/',
		'1', '1', '1', '1', '1', '1', '1', '1', '/',
		'1', '1', '1', '1', '1', '1', '1', '1', '/',
		'1', '1', '1', '1', '1', '1', '1', '1', '/',
		'1', '1', '1', '1', '1', '1', '1', '1', '/',
		'1', '1', '1', '1', '1', '1', '1', '1', '/',
		'1', '1', '1', '1', '1', '1', '1', '1', '/',
		'1', '1', '1', '1', '1', '1', '1', '1',
	}
	offset := 0
	for r := 7; r >= 0; r-- {
		for f := 0; f < numOfSquaresInRow; f++ {
			sq := NewSquare(File(f), Rank(r))
			p := b.Piece(sq)
			if p != NoPiece {
				expandedFen[offset] = fenReverseMap[p]
			}
			offset += 1
		}
		offset += 1
	}
	fen := string(expandedFen)
	for i := 8; i > 1; i-- {
		repeatStr := strings.Repeat("1", i)
		countStr := strconv.Itoa(i)
		fen = strings.Replace(fen, repeatStr, countStr, -1)
	}
	return fen
}

// Eq returns whether this board is the same as the other board
func (b *Board) Eq(other *Board) bool {
	for i, v := range b.array {
		if v != other.array[i] {
			return false
		}
	}
	return true
}

// FEN returns the FEN representation of the board.
func (b *Board) FEN() string {
	return b.String()
}

// Piece returns the piece for the given square.
func (b *Board) Piece(sq Square) Piece {
	if !b.isOccupied(sq) {
		return NoPiece
	}
	for _, p := range allPieces {
		bb := b.bbForPiece(p)
		if bb.Occupied(sq) {
			return p
		}
	}
	return NoPiece
}

// MarshalText implements the encoding.TextMarshaler interface and returns
// a string in the FEN board format: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
func (b *Board) MarshalText() (text []byte, err error) {
	return []byte(b.String()), nil
}

// UnmarshalText implements the encoding.TextUnarshaler interface and takes
// a string in the FEN board format: rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR
func (b *Board) UnmarshalText(text []byte) error {
	cp, err := fenBoard(string(text))
	if err != nil {
		return err
	}
	*b = *cp
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface and returns
// the bitboard representations as a array of bytes.  Bitboads are encoded
// in the following order: WhiteKing, WhiteQueen, WhiteRook, WhiteBishop, WhiteKnight
// WhitePawn, BlackKing, BlackQueen, BlackRook, BlackBishop, BlackKnight, BlackPawn
func (b *Board) MarshalBinary() (data []byte, err error) {
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, b.array[:6])
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.BigEndian, b.array[16:])
	return buf.Bytes(), err
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface and parses
// the bitboard representations as a array of bytes.  Bitboads are decoded
// in the following order: WhiteKing, WhiteQueen, WhiteRook, WhiteBishop, WhiteKnight
// WhitePawn, BlackKing, BlackQueen, BlackRook, BlackBishop, BlackKnight, BlackPawn
func (b *Board) UnmarshalBinary(data []byte) error {
	if len(data) != 96 {
		return errors.New("chess: invalid number of bytes for board unmarshal binary")
	}
	for i := 0; i < 48; i += 8 {
		b.array[i>>3] = bitboard(binary.BigEndian.Uint64(data[i : i+8]))
	}
	for i := 0; i < 48; i += 8 {
		b.array[(i>>3)+16] = bitboard(binary.BigEndian.Uint64(data[i+48 : i+48+8]))
	}
	b.updateKings(nil)
	return nil
}

func (b *Board) update(m *Move) {
	p1 := m.piece
	if p1 == NoPiece {
		p1 = b.Piece(m.s1)
	}
	s1BB := bbForSquare(m.s1)
	s2BB := bbForSquare(m.s2)

	// move s1 piece to s2
	for _, p := range allPieces {
		bb := b.bbForPiece(p)
		// remove what was at s2
		b.setBBForPiece(p, bb & ^s2BB)
	}

	bb := b.bbForPiece(p1)
	b.setBBForPiece(p1, (bb & ^s1BB)|s2BB)

	// check promotion
	if m.promo != NoPromo {
		newPiece := GetPiece(m.promo.PieceType(), p1.Color())
		// remove pawn
		bbPawn := b.bbForPiece(p1)
		b.setBBForPiece(p1, bbPawn & ^s2BB)
		// add promo piece
		bbPromo := b.bbForPiece(newPiece)
		b.setBBForPiece(newPiece, bbPromo|s2BB)
	}
	// remove captured en passant piece
	if m.HasTag(EnPassant) {
		if p1.Color() == White {
			b.setBBForPiece(BlackPawn, ^(bbForSquare(m.s2)>>8)&b.bbForPiece(BlackPawn))
		} else {
			b.setBBForPiece(WhitePawn, ^(bbForSquare(m.s2)<<8)&b.bbForPiece(WhitePawn))
		}
	}
	// move rook for castle
	if p1.Color() == White && m.HasTag(KingSideCastle) {
		b.setBBForPiece(WhiteRook, (b.bbForPiece(WhiteRook) & ^bbForSquare(H1) | bbForSquare(F1)))
	} else if p1.Color() == White && m.HasTag(QueenSideCastle) {
		b.setBBForPiece(WhiteRook, (b.bbForPiece(WhiteRook) & ^bbForSquare(A1))|bbForSquare(D1))
	} else if p1.Color() == Black && m.HasTag(KingSideCastle) {
		b.setBBForPiece(BlackRook, b.bbForPiece(BlackRook) & ^bbForSquare(H8) | bbForSquare(F8))
	} else if p1.Color() == Black && m.HasTag(QueenSideCastle) {
		b.setBBForPiece(BlackRook, (b.bbForPiece(BlackRook) & ^bbForSquare(A8))|bbForSquare(D8))
	}
	b.updateKings(m)
	b.occupiedCache = 0
}

func (b *Board) updateKings(m *Move) {
	if m == nil {
		b.whiteKingSq = NoSquare
		b.blackKingSq = NoSquare

		for sq := 0; sq < numOfSquaresInBoard; sq++ {
			sqr := Square(sq)
			if b.array[WhiteKing].Occupied(sqr) {
				b.whiteKingSq = sqr
			} else if b.array[BlackKing].Occupied(sqr) {
				b.blackKingSq = sqr
			}
		}
	} else if m.s1 == b.whiteKingSq {
		b.whiteKingSq = m.s2
	} else if m.s1 == b.blackKingSq {
		b.blackKingSq = m.s2
	}
}

func (b *Board) copyInto(other *Board) {
	for i := 0; i < 22; i++ {
		other.array[i] = b.array[i]
	}
	other.whiteKingSq = b.whiteKingSq
	other.blackKingSq = b.blackKingSq
}

func (b *Board) whiteSqs() bitboard {
	var total uint64
	for i := WhiteKing; i <= WhitePawn; i++ {
		total = total | uint64(b.array[i])
	}
	return bitboard(total)
}

func (b *Board) blackSqs() bitboard {
	var total uint64
	for i := BlackKing; i <= BlackPawn; i++ {
		total = total | uint64(b.array[i])
	}
	return bitboard(total)
}

func (b *Board) occupied() bitboard {
	if b.occupiedCache == 0 {
		var total uint64
		for i := 0; i < 22; i++ {
			total = total | uint64(b.array[i])
		}
		b.occupiedCache = bitboard(total)
	}
	return b.occupiedCache
}

func (b *Board) isOccupied(sq Square) bool {
	mask := uint64(0b1 << int(sq))
	total := uint64(b.occupied())
	return (total & mask) != 0
}

func (b *Board) hasSufficientMaterial() bool {
	// queen, rook, or pawn exist
	if (b.array[WhiteQueen] | b.array[WhiteRook] | b.array[WhitePawn] |
		b.array[BlackQueen] | b.array[BlackRook] | b.array[BlackPawn]) != 0 {
		return true
	}
	// if king is missing then it is a test
	if b.array[WhiteKing] == 0 || b.array[BlackKing] == 0 {
		return true
	}
	count := map[PieceType]int{}

	for i := 0; i < 6; i++ {
		count[PieceType(i)] += bits.OnesCount64(uint64(b.array[i]))
		count[PieceType(i)] += bits.OnesCount64(uint64(b.array[i+16]))
	}

	// 	king versus king
	if count[Bishop] == 0 && count[Knight] == 0 {
		return false
	}
	// king and bishop versus king
	if count[Bishop] == 1 && count[Knight] == 0 {
		return false
	}
	// king and knight versus king
	if count[Bishop] == 0 && count[Knight] == 1 {
		return false
	}
	// king and bishop(s) versus king and bishop(s) with the bishops on the same colour.
	if count[Knight] == 0 {
		bishops := uint64(b.array[WhiteBishop] | b.array[BlackBishop])
		lightCount := bits.OnesCount64(bishops & lightSquares)
		darkCount := bits.OnesCount64(bishops & darkSquares)
		if lightCount == 0 || darkCount == 0 {
			return false
		}
	}
	return true
}

func (b *Board) bbForPiece(p Piece) bitboard {
	return b.array[p]
}

func (b *Board) setBBForPiece(p Piece, bb bitboard) {
	b.array[p] = bb
}

func (b *Board) pieceAt(sq Square) Piece {
	mask := uint64(0b1) << int(sq)
	for i, bb := range b.array {
		if uint64(bb)&mask != 0 {
			return Piece(i)
		}
	}
	return NoPiece
}
