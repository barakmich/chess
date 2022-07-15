package chess

// NOTE(barakmich):
// Piece, PieceType and Color constant values are carefully chosen
// to allow for bit operations between them.
//
// A Piece has the upper 4 bits as Color and the
// lower 4 bits as PieceType

// Color represents the color of a chess piece.
type Color uint8

const (
	// White represents the color white
	White Color = 0
	// Black represents the color black
	Black Color = 1
	// NoColor represents no color
	NoColor Color = 15
)

// Other returns the opposite color of the receiver.
func (c Color) Other() Color {
	if c == White {
		return Black
	}
	return White
}

// String implements the fmt.Stringer interface and returns
// the color's FEN compatible notation.
func (c Color) String() string {
	switch c {
	case White:
		return "w"
	case Black:
		return "b"
	}
	return "-"
}

// Name returns a display friendly name.
func (c Color) Name() string {
	switch c {
	case White:
		return "White"
	case Black:
		return "Black"
	}
	return "No Color"
}

// PieceType is the type of a piece.
type PieceType uint8

const (
	// King represents a king
	King PieceType = 0
	// Queen represents a queen
	Queen PieceType = 1
	// Rook represents a rook
	Rook PieceType = 2
	// Bishop represents a bishop
	Bishop PieceType = 3
	// Knight represents a knight
	Knight PieceType = 4
	// Pawn represents a pawn
	Pawn PieceType = 5
	// NoPieceType represents a lack of piece type
	NoPieceType PieceType = 15
)

// PromoType is a promotion choice
type PromoType uint8

const (
	NoPromo PromoType = iota
	// Queen represents a queen
	PromoQueen
	// Rook represents a rook
	PromoRook
	// Bishop represents a bishop
	PromoBishop
	// Knight represents a knight
	PromoKnight
)

func (promo PromoType) PieceType() PieceType {
	if promo == NoPromo {
		return NoPieceType
	}
	return PieceType(promo)
}

func promoFromPieceType(p PieceType) PromoType {
	switch p {
	case Queen:
		return PromoQueen
	case Rook:
		return PromoRook
	case Knight:
		return PromoKnight
	case Bishop:
		return PromoBishop
	}
	return NoPromo
}

var allPieceTypes = [6]PieceType{King, Queen, Rook, Bishop, Knight, Pawn}

// PieceTypes returns a slice of all piece types.
func PieceTypes() [6]PieceType {
	return allPieceTypes
}

func (p PieceType) String() string {
	switch p {
	case King:
		return "k"
	case Queen:
		return "q"
	case Rook:
		return "r"
	case Bishop:
		return "b"
	case Knight:
		return "n"
	case Pawn:
		return "p"
	}
	return ""
}

// Piece is a piece type with a color.
type Piece uint8

const (
	// WhiteKing is a white king
	WhiteKing Piece = 0
	// WhiteQueen is a white queen
	WhiteQueen Piece = 1
	// WhiteRook is a white rook
	WhiteRook Piece = 2
	// WhiteBishop is a white bishop
	WhiteBishop Piece = 3
	// WhiteKnight is a white knight
	WhiteKnight Piece = 4
	// WhitePawn is a white pawn
	WhitePawn Piece = 5
	// BlackKing is a black king
	BlackKing Piece = 16
	// BlackQueen is a black queen
	BlackQueen Piece = 17
	// BlackRook is a black rook
	BlackRook Piece = 18
	// BlackBishop is a black bishop
	BlackBishop Piece = 19
	// BlackKnight is a black knight
	BlackKnight Piece = 20
	// BlackPawn is a black pawn
	BlackPawn Piece = 21
	// NoPiece represents no piece
	NoPiece Piece = 255
)

var (
	allPieces = []Piece{
		WhiteKing, WhiteQueen, WhiteRook, WhiteBishop, WhiteKnight, WhitePawn,
		BlackKing, BlackQueen, BlackRook, BlackBishop, BlackKnight, BlackPawn,
	}
)

func GetPiece(t PieceType, c Color) Piece {
	return Piece(uint8(c)<<4 | uint8(t))
}

// Type returns the type of the piece.
func (p Piece) Type() PieceType {
	return PieceType(p & 0xF)
}

// Color returns the color of the piece.
func (p Piece) Color() Color {
	return Color(p >> 4)
}

// String implements the fmt.Stringer interface
func (p Piece) String() string {
	v, ok := pieceUnicodes[p]
	if !ok {
		return " "
	}
	return v
}

var (
	pieceUnicodes = map[Piece]string{
		WhiteKing:   "♔",
		WhiteQueen:  "♕",
		WhiteRook:   "♖",
		WhiteBishop: "♗",
		WhiteKnight: "♘",
		WhitePawn:   "♙",
		BlackKing:   "♚",
		BlackQueen:  "♛",
		BlackRook:   "♜",
		BlackBishop: "♝",
		BlackKnight: "♞",
		BlackPawn:   "♟",
	}
)

func (p Piece) getFENChar() string {
	return string(fenReverseMap[p])
}
