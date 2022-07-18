package chess

import "github.com/barakmich/chess/bitflip"

type engine struct{}

func (engine) CalcMoves(pos *Position, first bool) []*Move {
	// generate possible moves
	moves := standardMoves(pos, first)
	// return moves including castles
	return append(moves, castleMoves(pos)...)
}

func (engine) Status(pos *Position) Method {
	hasMove := false
	if pos.validMoves != nil {
		hasMove = len(pos.validMoves) > 0
	} else {
		hasMove = len(engine{}.CalcMoves(pos, true)) > 0
	}
	if !pos.inCheck && !hasMove {
		return Stalemate
	} else if pos.inCheck && !hasMove {
		return Checkmate
	}
	return NoMethod
}

var (
	promoPieceTypes = []PromoType{PromoQueen, PromoRook, PromoBishop, PromoKnight}
)

func standardMoves(pos *Position, first bool) []*Move {
	// compute allowed destination bitboard
	bbAllowed := ^pos.board.whiteSqs()
	if pos.Turn() == Black {
		bbAllowed = ^pos.board.blackSqs()
	}
	moves := []*Move{}
	// iterate through pieces to find possible moves
	for _, typ := range allPieceTypes {
		p := GetPiece(typ, pos.Turn())
		// iterate through possible starting squares for piece
		s1BB := pos.board.bbForPiece(p)
		if s1BB == 0 {
			continue
		}
		for s1i := 0; s1i < numOfSquaresInBoard; s1i++ {
			if s1BB&bbForSquare(Square(s1i)) == 0 {
				continue
			}
			s1 := Square(s1i)
			// iterate through possible destination squares for piece
			var s2BB bitboard
			if p.Type() == Pawn {
				s2BB = pawnMoves(pos, s1)
			} else {
				s2BB = bbForPossiblePieceMoves(pos.board.occupied(), p.Type(), s1)
			}
			s2BB = s2BB & bbAllowed
			if s2BB == 0 {
				continue
			}
			for s2i := 0; s2i < numOfSquaresInBoard; s2i++ {
				if s2BB&bbForSquare(Square(s2i)) == 0 {
					continue
				}
				s2 := Square(s2i)
				// add promotions if pawn on promo square
				if (p == WhitePawn && s2.Rank() == Rank8) || (p == BlackPawn && s2.Rank() == Rank1) {
					for _, pt := range promoPieceTypes {
						m := &Move{piece: p, s1: s1, s2: s2, promo: pt}
						addTags(m, pos)
						// filter out moves that put king into check
						if !m.HasTag(inCheck) {
							moves = append(moves, m)
							if first {
								return moves
							}
						}
					}
				} else {
					m := &Move{piece: p, s1: s1, s2: s2}
					addTags(m, pos)
					// filter out moves that put king into check
					if !m.HasTag(inCheck) {
						moves = append(moves, m)
						if first {
							return moves
						}
					}
				}
			}
		}
	}
	return moves
}

func addTags(m *Move, pos *Position) {
	p := m.piece
	if p == NoPiece {
		p = pos.board.Piece(m.s1)
	}
	if pos.board.isOccupied(m.s2) {
		m.addTag(Capture)
	} else if m.s2 == pos.enPassantSquare && p.Type() == Pawn {
		m.addTag(EnPassant)
	}
	// determine if in check after move (makes move invalid)
	tmpBoard := pos.tempCopyBoard()
	tmpBoard.update(m)
	if isInCheck(tmpBoard, pos.turn) {
		m.addTag(inCheck)
	}
	// determine if opponent in check after move
	if isInCheck(tmpBoard, pos.turn.Other()) {
		m.addTag(Check)
	}
	pos.finishTempCopy(tmpBoard)
}

func isInCheck(board *Board, turn Color) bool {
	kingSq := board.whiteKingSq
	if turn == Black {
		kingSq = board.blackKingSq
	}
	// king should only be missing in tests / examples
	if kingSq == NoSquare {
		return false
	}
	return squaresAreAttacked(board, turn, kingSq)
}

func squaresAreAttacked(board *Board, turn Color, sqs ...Square) bool {
	otherColor := turn.Other()
	occ := board.occupied()
	for _, sq := range sqs {
		// hot path check to see if attack vector is possible
		dia := diaAttack(occ, sq)
		hv := hvAttack(occ, sq)
		s2BB := board.blackSqs()
		if turn == Black {
			s2BB = board.whiteSqs()
		}
		if ((dia|hv)&s2BB)|(bbKnightMoves[sq]&s2BB) == 0 {
			continue
		}
		// check queen attack vector
		queenBB := board.bbForPiece(GetPiece(Queen, otherColor))
		bb := (dia | hv) & queenBB
		if bb != 0 {
			return true
		}
		// check rook attack vector
		rookBB := board.bbForPiece(GetPiece(Rook, otherColor))
		bb = hv & rookBB
		if bb != 0 {
			return true
		}
		// check bishop attack vector
		bishopBB := board.bbForPiece(GetPiece(Bishop, otherColor))
		bb = dia & bishopBB
		if bb != 0 {
			return true
		}
		// check knight attack vector
		knightBB := board.bbForPiece(GetPiece(Knight, otherColor))
		bb = bbKnightMoves[sq] & knightBB
		if bb != 0 {
			return true
		}
		// check pawn attack vector
		if turn == White {
			capLeft := (board.bbForPiece(BlackPawn) & ^bbFileH & ^bbRank1) >> 7
			capRight := (board.bbForPiece(BlackPawn) & ^bbFileA & ^bbRank1) >> 9
			bb = (capRight | capLeft) & bbForSquare(sq)
			if bb != 0 {
				return true
			}
		} else {
			capLeft := (board.bbForPiece(WhitePawn) & ^bbFileH & ^bbRank8) << 9
			capRight := (board.bbForPiece(WhitePawn) & ^bbFileA & ^bbRank8) << 7
			bb = (capRight | capLeft) & bbForSquare(sq)
			if bb != 0 {
				return true
			}
		}
		// check king attack vector
		kingBB := board.bbForPiece(GetPiece(King, otherColor))
		bb = bbKingMoves[sq] & kingBB
		if bb != 0 {
			return true
		}
	}
	return false
}

func bbForPossiblePieceMoves(occupied bitboard, pt PieceType, sq Square) bitboard {
	switch pt {
	case King:
		return bbKingMoves[sq]
	case Queen:
		return diaAttack(occupied, sq) | hvAttack(occupied, sq)
	case Rook:
		return hvAttack(occupied, sq)
	case Bishop:
		return diaAttack(occupied, sq)
	case Knight:
		return bbKnightMoves[sq]
	}
	return bitboard(0)
}

// TODO can calc isInCheck twice
func castleMoves(pos *Position) []*Move {
	moves := []*Move{}
	kingSide := pos.castleRights.CanCastle(pos.Turn(), KingSide)
	queenSide := pos.castleRights.CanCastle(pos.Turn(), QueenSide)
	occupied := pos.board.occupied()
	// white king side
	if pos.turn == White && kingSide &&
		(occupied&(bbForSquare(F1)|bbForSquare(G1))) == 0 &&
		!squaresAreAttacked(pos.board, pos.turn, F1, G1) &&
		!pos.inCheck {
		m := &Move{piece: WhiteKing, s1: E1, s2: G1}
		m.addTag(KingSideCastle)
		addTags(m, pos)
		moves = append(moves, m)
	}
	// white queen side
	if pos.turn == White && queenSide &&
		(occupied&(bbForSquare(B1)|bbForSquare(C1)|bbForSquare(D1))) == 0 &&
		!squaresAreAttacked(pos.board, pos.turn, C1, D1) &&
		!pos.inCheck {
		m := &Move{piece: WhiteKing, s1: E1, s2: C1}
		m.addTag(QueenSideCastle)
		addTags(m, pos)
		moves = append(moves, m)
	}
	// black king side
	if pos.turn == Black && kingSide &&
		(occupied&(bbForSquare(F8)|bbForSquare(G8))) == 0 &&
		!squaresAreAttacked(pos.board, pos.turn, F8, G8) &&
		!pos.inCheck {
		m := &Move{piece: BlackKing, s1: E8, s2: G8}
		m.addTag(KingSideCastle)
		addTags(m, pos)
		moves = append(moves, m)
	}
	// black queen side
	if pos.turn == Black && queenSide &&
		(occupied&(bbForSquare(B8)|bbForSquare(C8)|bbForSquare(D8))) == 0 &&
		!squaresAreAttacked(pos.board, pos.turn, C8, D8) &&
		!pos.inCheck {
		m := &Move{piece: BlackKing, s1: E8, s2: C8}
		m.addTag(QueenSideCastle)
		addTags(m, pos)
		moves = append(moves, m)
	}
	return moves
}

func pawnMoves(pos *Position, sq Square) bitboard {
	bb := bbForSquare(sq)
	occupied := pos.board.occupied()
	noccupied := ^occupied
	var bbEnPassant bitboard
	if pos.enPassantSquare != NoSquare {
		bbEnPassant = bbForSquare(pos.enPassantSquare)
	}
	if pos.Turn() == White {
		capRight := ((bb & ^bbFileH & ^bbRank8) << 9) & (pos.board.blackSqs() | bbEnPassant)
		capLeft := ((bb & ^bbFileA & ^bbRank8) << 7) & (pos.board.blackSqs() | bbEnPassant)
		upOne := ((bb & ^bbRank8) << 8) & noccupied
		upTwo := ((upOne & bbRank3) << 8) & noccupied
		return capRight | capLeft | upOne | upTwo
	}
	capRight := ((bb & ^bbFileH & ^bbRank1) >> 7) & (pos.board.whiteSqs() | bbEnPassant)
	capLeft := ((bb & ^bbFileA & ^bbRank1) >> 9) & (pos.board.whiteSqs() | bbEnPassant)
	upOne := ((bb & ^bbRank1) >> 8) & (noccupied)
	upTwo := ((upOne & bbRank6) >> 8) & (noccupied)
	return capRight | capLeft | upOne | upTwo
}

func diaAttack(occupied bitboard, sq Square) bitboard {
	pos := bbForSquare(sq)
	dMask := bbDiagonals[int(sq)]
	adMask := bbAntiDiagonals[int(sq)]
	return bitboard(bitflip.BishopRookAttacks(uint64(occupied), uint64(pos), uint64(dMask), uint64(adMask)))
}

func hvAttack(occupied bitboard, sq Square) bitboard {
	pos := bbForSquare(sq)
	rankMask := bbRanks[Square(sq).Rank()]
	fileMask := bbFiles[Square(sq).File()]
	return bitboard(bitflip.BishopRookAttacks(uint64(occupied), uint64(pos), uint64(rankMask), uint64(fileMask)))
}

func queenAttack(occupied bitboard, sq Square) bitboard {
	pos := bbForSquare(sq)
	rankMask := bbRanks[Square(sq).Rank()]
	fileMask := bbFiles[Square(sq).File()]
	dMask := bbDiagonals[int(sq)]
	adMask := bbAntiDiagonals[int(sq)]
	return bitboard(bitflip.QueenAttacks(
		uint64(occupied),
		uint64(pos),
		uint64(rankMask),
		uint64(fileMask),
		uint64(dMask),
		uint64(adMask),
	))
}
func linearAttack(occupied, pos, mask bitboard) bitboard {
	oInMask := occupied & mask
	return ((oInMask - (pos << 1)) ^ (oInMask.Reverse() - (pos.Reverse() << 1)).Reverse()) & mask
}

const (
	bbFileA bitboard = 72340172838076673
	bbFileB bitboard = 144680345676153346
	bbFileC bitboard = 289360691352306692
	bbFileD bitboard = 578721382704613384
	bbFileE bitboard = 1157442765409226768
	bbFileF bitboard = 2314885530818453536
	bbFileG bitboard = 4629771061636907072
	bbFileH bitboard = 9259542123273814144

	bbRank1 bitboard = 255
	bbRank2 bitboard = 65280
	bbRank3 bitboard = 16711680
	bbRank4 bitboard = 4278190080
	bbRank5 bitboard = 1095216660480
	bbRank6 bitboard = 280375465082880
	bbRank7 bitboard = 71776119061217280
	bbRank8 bitboard = 18374686479671623680
)

// TODO make method on Square
func bbForSquare(sq Square) bitboard {
	return bitboard(0b1 << sq)
}

func bbGetFirstSquare(bb bitboard) Square {
	mask := bitboard(0b1)
	for i := 0; i < 64; i++ {
		if mask&bb != 0 {
			return Square(i)
		}
		mask = mask << 1
	}
	return NoSquare
}

var (
	bbFiles = [8]bitboard{bbFileA, bbFileB, bbFileC, bbFileD, bbFileE, bbFileF, bbFileG, bbFileH}
	bbRanks = [8]bitboard{bbRank1, bbRank2, bbRank3, bbRank4, bbRank5, bbRank6, bbRank7, bbRank8}

	bbDiagonals = [64]bitboard{9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 2151686160, 8405024, 32832, 128, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 2151686160, 8405024, 32832, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 2151686160, 8405024, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 2151686160, 577588855528488960, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 288794425616760832, 577588855528488960, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 144396663052566528, 288794425616760832, 577588855528488960, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 72057594037927936, 144396663052566528, 288794425616760832, 577588855528488960, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745}

	bbAntiDiagonals = [64]bitboard{1, 258, 66052, 16909320, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 258, 66052, 16909320, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 66052, 16909320, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 16909320, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 1161999622361579520, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 1161999622361579520, 2323998145211531264, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 1161999622361579520, 2323998145211531264, 4647714815446351872, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 1161999622361579520, 2323998145211531264, 4647714815446351872, 9223372036854775808}

	bbKnightMoves = [64]bitboard{132096, 329728, 659712, 1319424, 2638848, 5277696, 10489856, 4202496, 33816580, 84410376, 168886289, 337772578, 675545156, 1351090312, 2685403152, 1075839008, 8657044482, 21609056261, 43234889994, 86469779988, 172939559976, 345879119952, 687463207072, 275414786112, 2216203387392, 5531918402816, 11068131838464, 22136263676928, 44272527353856, 88545054707712, 175990581010432, 70506185244672, 567348067172352, 1416171111120896, 2833441750646784, 5666883501293568, 11333767002587136, 22667534005174272, 45053588738670592, 18049583422636032, 145241105196122112, 362539804446949376, 725361088165576704, 1450722176331153408, 2901444352662306816, 5802888705324613632, 11533718717099671552, 4620693356194824192, 288234782788157440, 576469569871282176, 1224997833292120064, 2449995666584240128, 4899991333168480256, 9799982666336960512, 1152939783987658752, 2305878468463689728, 1128098930098176, 2257297371824128, 4796069720358912, 9592139440717824, 19184278881435648, 38368557762871296, 4679521487814656, 9077567998918656}

	bbKingMoves = [64]bitboard{770, 1797, 3594, 7188, 14376, 28752, 57504, 49216, 197123, 460039, 920078, 1840156, 3680312, 7360624, 14721248, 12599488, 50463488, 117769984, 235539968, 471079936, 942159872, 1884319744, 3768639488, 3225468928, 12918652928, 30149115904, 60298231808, 120596463616, 241192927232, 482385854464, 964771708928, 825720045568, 3307175149568, 7718173671424, 15436347342848, 30872694685696, 61745389371392, 123490778742784, 246981557485568, 211384331665408, 846636838289408, 1975852459884544, 3951704919769088, 7903409839538176, 15806819679076352, 31613639358152704, 63227278716305408, 54114388906344448, 216739030602088448, 505818229730443264, 1011636459460886528, 2023272918921773056, 4046545837843546112, 8093091675687092224, 16186183351374184448, 13853283560024178688, 144959613005987840, 362258295026614272, 724516590053228544, 1449033180106457088, 2898066360212914176, 5796132720425828352, 11592265440851656704, 4665729213955833856}
)
