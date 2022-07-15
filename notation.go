package chess

import (
	"fmt"
	"regexp"
	"strings"
)

type Notation int

const (
	SANNotation = iota
	UCINotation
	LongAlgebraicNotation
)

func (pos *Position) EncodeMove(m *Move, n Notation) string {
	switch n {
	case SANNotation:
		return pos.EncodeSAN(m)
	case UCINotation:
		return pos.EncodeUCI(m)
	case LongAlgebraicNotation:
		return pos.EncodeLongAlgebraic(m)
	}
	panic("unreachable")
}

func (pos *Position) DecodeMove(s string, n ...Notation) (*Move, error) {
	if len(n) != 0 {
		switch n[0] {
		case SANNotation:
			return pos.DecodeSAN(s)
		case UCINotation:
			return pos.DecodeUCI(s)
		case LongAlgebraicNotation:
			return pos.DecodeLongAlgebraic(s)
		}
	}
	if m, err := pos.DecodeSAN(s); err == nil {
		return m, nil
	}
	if m, err := pos.DecodeLongAlgebraic(s); err == nil {
		return m, nil
	}
	if m, err := pos.DecodeUCI(s); err == nil {
		return m, nil
	}
	return nil, fmt.Errorf(`chess: failed to decode notation text "%s" for position %s`, s, pos)
}

// Encode implements the Encoder interface.
func (pos *Position) EncodeUCI(m *Move) string {
	return m.S1().String() + m.S2().String() + m.Promo().String()
}

// Decode implements the Decoder interface.
func (pos *Position) DecodeUCI(s string) (*Move, error) {
	l := len(s)
	err := fmt.Errorf(`chess: failed to decode long algebraic notation text "%s" for position %s`, s, pos)
	if l < 4 || l > 5 {
		return nil, err
	}
	s1, ok := strToSquareMap[s[0:2]]
	if !ok {
		return nil, err
	}
	s2, ok := strToSquareMap[s[2:4]]
	if !ok {
		return nil, err
	}
	promo := NoPieceType
	if l == 5 {
		promo = pieceTypeFromChar(s[4:5])
		if promo == NoPieceType {
			return nil, err
		}
	}
	m := &Move{s1: s1, s2: s2, promo: promoFromPieceType(promo)}
	if pos == nil {
		return m, nil
	}
	p := pos.Board().Piece(s1)
	m.piece = p
	if p.Type() == King {
		if (s1 == E1 && s2 == G1) || (s1 == E8 && s2 == G8) {
			m.addTag(KingSideCastle)
		} else if (s1 == E1 && s2 == C1) || (s1 == E8 && s2 == C8) {
			m.addTag(QueenSideCastle)
		}
	} else if p.Type() == Pawn && s2 == pos.enPassantSquare {
		m.addTag(EnPassant)
		m.addTag(Capture)
	}
	c1 := p.Color()
	c2 := pos.Board().Piece(s2).Color()
	if c2 != NoColor && c1 != c2 {
		m.addTag(Capture)
	}
	return m, nil
}

func (pos *Position) EncodeSAN(m *Move) string {
	return pos.encodeSANInternal(m, nil)
}

func (pos *Position) encodeSANInternal(m *Move, validMoves []*Move) string {
	checkChar := getCheckChar(pos, m)
	if m.HasTag(KingSideCastle) {
		return "O-O" + checkChar
	} else if m.HasTag(QueenSideCastle) {
		return "O-O-O" + checkChar
	}
	p := m.piece
	if p == NoPiece {
		p = pos.Board().Piece(m.S1())
	}
	pChar := charFromPieceType(p.Type())
	s1Str := formS1(pos, m, validMoves)
	capChar := ""
	if m.HasTag(Capture) || m.HasTag(EnPassant) {
		capChar = "x"
		if p.Type() == Pawn && s1Str == "" {
			capChar = m.s1.File().String() + "x"
		}
	}
	promoText := charForPromo(m.promo)
	var sb strings.Builder
	sb.WriteString(pChar)
	sb.WriteString(s1Str)
	sb.WriteString(capChar)
	sb.WriteString(m.s2.String())
	sb.WriteString(promoText)
	sb.WriteString(checkChar)
	return sb.String()
}

var pgnRegex = regexp.MustCompile(`^(?:([RNBQKP]?)([abcdefgh]?)(\d?)(x?)([abcdefgh])(\d)(=[QRBN])?|(O-O(?:-O)?))([+#!?]|e\.p\.)*$`)

func algebraicNotationParts(s string) ([]string, error) {
	submatches := pgnRegex.FindStringSubmatch(s)
	if len(submatches) == 0 {
		return nil, fmt.Errorf("could not decode algebraic notation %s", s)
	}
	return submatches, nil
}

type moveAndStr struct {
	str  string
	move *Move
}

// Decode implements the Decoder interface.
func (pos *Position) DecodeSAN(s string) (*Move, error) {

	pos.ensureValidMoves()
	validMoveStrings := make([]string, len(pos.validMoves))
	for i, m := range pos.validMoves {
		moveStr := pos.encodeSANInternal(m, pos.validMoves)
		validMoveStrings[i] = moveStr
	}

	for i, moveStr := range validMoveStrings {
		if strings.HasPrefix(moveStr, s) {
			return pos.validMoves[i].copy(), nil
		}
	}

	submatches, err := algebraicNotationParts(s)
	if err != nil {
		return nil, fmt.Errorf("chess: %+v for position %s", err, pos.String())
	}
	piece := submatches[1]
	originFile := submatches[2]
	originRank := submatches[3]
	capture := submatches[4]
	file := submatches[5]
	rank := submatches[6]
	promotes := submatches[7]
	castles := submatches[8]

	var sb strings.Builder
	sb.WriteString(piece)
	sb.WriteString(originFile)
	sb.WriteString(originRank)
	sb.WriteString(capture)
	sb.WriteString(file)
	sb.WriteString(rank)
	sb.WriteString(promotes)
	sb.WriteString(castles)
	cleaned := sb.String()

	for i, move := range validMoveStrings {
		if strings.HasPrefix(move, cleaned) {
			return pos.validMoves[i].copy(), nil
		}
	}

	// Try and remove the disambiguators and see if it parses. Sometimes they
	// get extraneously added.
	options := []string{}

	if piece != "" {
		options = append(options, piece+capture+file+rank+promotes+castles)            // no origin
		options = append(options, piece+originRank+capture+file+rank+promotes+castles) // no origin file
		options = append(options, piece+originFile+capture+file+rank+promotes+castles) // no origin rank
	} else {
		if capture != "" {
			// Possibly a pawn capture. In order to parse things like d4xe5, we need
			// to try parsing without the rank.
			options = append(options, piece+originFile+capture+file+rank+promotes+castles) // no origin rank
		}
		if originFile != "" && originRank != "" {
			options = append(options, piece+capture+file+rank+promotes+castles) // no origin
		}
	}

	for i, move := range validMoveStrings {
		for _, opt := range options {
			if strings.HasPrefix(move, opt) {
				return pos.validMoves[i].copy(), nil
			}
		}
	}

	return nil, fmt.Errorf("chess: could not decode algebraic notation %s for position %s", s, pos.String())
}

func (pos *Position) EncodeLongAlgebraic(m *Move) string {
	checkChar := getCheckChar(pos, m)
	if m.HasTag(KingSideCastle) {
		return "O-O" + checkChar
	} else if m.HasTag(QueenSideCastle) {
		return "O-O-O" + checkChar
	}
	p := m.piece
	if p == NoPiece {
		p = pos.Board().Piece(m.S1())
	}
	pChar := charFromPieceType(p.Type())
	s1Str := m.s1.String()
	capChar := ""
	if m.HasTag(Capture) || m.HasTag(EnPassant) {
		capChar = "x"
		if p.Type() == Pawn && s1Str == "" {
			capChar = m.s1.File().String() + "x"
		}
	}
	promoText := charForPromo(m.promo)
	return pChar + s1Str + capChar + m.s2.String() + promoText + checkChar
}

// Decode implements the Decoder interface.
func (pos *Position) DecodeLongAlgebraic(s string) (*Move, error) {
	return pos.DecodeSAN(s)
}

func getCheckChar(pos *Position, move *Move) string {
	if !move.HasTag(Check) {
		return ""
	}
	nextPos := pos.Update(move)
	if nextPos.Status() == Checkmate {
		return "#"
	}
	return "+"
}

func formS1(pos *Position, m *Move, moves []*Move) string {
	p := m.piece
	if p == NoPiece {
		p = pos.board.Piece(m.s1)
	}
	if p.Type() == Pawn || p.Type() == King {
		return ""
	}

	var req, fileReq, rankReq bool
	if moves == nil {
		moves = pos.ValidMoves()
	}

	for _, mv := range moves {
		otherPiece := mv.piece
		if mv.s1 != m.s1 && mv.s2 == m.s2 && p == otherPiece {
			req = true

			if mv.s1.File() == m.s1.File() {
				rankReq = true
			}

			if mv.s1.Rank() == m.s1.Rank() {
				fileReq = true
			}
		}
	}

	var s1 = ""

	if fileReq || !rankReq && req {
		s1 = m.s1.File().String()
	}

	if rankReq {
		s1 += m.s1.Rank().String()
	}

	return s1
}

func charForPromo(p PromoType) string {
	switch p {
	case PromoBishop:
		return "=B"
	case PromoKnight:
		return "=N"
	case PromoQueen:
		return "=Q"
	case PromoRook:
		return "=R"
	}
	return ""
}

func charFromPieceType(p PieceType) string {
	switch p {
	case King:
		return "K"
	case Queen:
		return "Q"
	case Rook:
		return "R"
	case Bishop:
		return "B"
	case Knight:
		return "N"
	}
	return ""
}

func pieceTypeFromChar(c string) PieceType {
	switch c {
	case "q":
		return Queen
	case "r":
		return Rook
	case "b":
		return Bishop
	case "n":
		return Knight
	}
	return NoPieceType
}
