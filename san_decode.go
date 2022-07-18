package chess

import (
	"errors"
	"fmt"
	"math/bits"
	"strings"
	"unicode"
)

func pieceFromChar(c rune) Piece {
	v, ok := fenPieceMap[string(c)]
	if !ok {
		return NoPiece
	}
	return v
}

func parseSAN(s string, pos *Position) (Move, error) {
	if s == "--" {
		return 0, nil
	}
	s = strings.TrimSpace(s)

	var (
		typ      PieceType
		fileHint int = -1
		rankHint int = -1
		toSq     Square
		move     Move
		err      error
	)

	if len(s) < 2 {
		return 0, errors.New("parseSAN: invalid move")
	}

	// Handle castling
	if strings.HasPrefix(s, "O-O-O") || strings.HasPrefix(s, "0-0-0") {
		move = move.addTag(QueenSideCastle)
		if pos.turn == White {
			move = move.setPiece(WhiteKing)
			move = move.setS1(E1)
			move = move.setS2(C1)
		} else {
			move = move.setPiece(BlackKing)
			move = move.setS1(E8)
			move = move.setS2(C8)
		}
		return parseSANTail(move, s[5:])
	}
	if strings.HasPrefix(s, "O-O") || strings.HasPrefix(s, "0-0") {
		move = move.addTag(QueenSideCastle)
		if pos.turn == White {
			move = move.setPiece(WhiteKing)
			move = move.setS1(E1)
			move = move.setS2(G1)
		} else {
			move = move.setPiece(BlackKing)
			move = move.setS1(E8)
			move = move.setS2(G8)
		}
		return parseSANTail(move, s[3:])
	}

	originalMove := s
	// Find the index of the last number.
	lastNum := -1
	for i, c := range s {
		if unicode.IsNumber(c) {
			lastNum = i
		}
	}
	if lastNum == -1 || lastNum < 1 {
		// We didn't find any numbers, this isn't valid.
		return 0, fmt.Errorf("parseSAN: couldn't find a square number in `%s`", s)
	}

	// Split the parts of the move
	head := s[:lastNum-1]
	// Every SAN move contains the full destination square
	toSquareStr := s[lastNum-1 : lastNum+1]
	toSq = strToSquareMap[strings.ToLower(toSquareStr)]
	// These are the extra info parsed at the end
	tail := s[lastNum+1:]

	if strings.Contains(head, "x") {
		// Double check
		move = move.addTag(Capture)
		head = strings.ReplaceAll(head, "x", "")
	}

	switch len(head) {
	case 0:
		typ = Pawn
	case 1:
		// Either a piece move/capture or a pawn capture.
		// Capitalization is important here; consider the conflation of
		// "bxc5" or "Bxc5"
		if head[0] < 0x60 {
			typ = fenPieceMap[head[0:1]].Type()
		} else {
			typ = Pawn
			fileHint = int(head[0] - 0x61)
			if fileHint > 7 {
				// We're assuming a pawn, but thanks to the capitilization rule, this
				// is just incorrect (eg, "nf3")
				return 0, fmt.Errorf("parseSAN: invalid capitalization `%s` for move `%s`", head, originalMove)
			}
		}
	case 2:
		// A disambiguated move. Must contain a piece at the front.
		typ = fenPieceMap[head[0:1]].Type()
		if head[1] > 0x30 && head[1] < 0x3A {
			// It's a number disambiguator.
			rankHint = int(head[1]) - 0x31
		} else {
			// It's a rank disambiguator
			if head[1] < 0x60 {
				// It's uppercase for some reason -- but we can deal.
				fileHint = int(head[1] - 0x41)
			} else {
				fileHint = int(head[1] - 0x61)
			}
		}
	case 3:
		// A fully disambiguated move. Contains all the info.
		typ = fenPieceMap[head[0:1]].Type()
		fromSq := strToSquareMap[strings.ToLower(head[1:])]
		rankHint = int(fromSq.Rank())
		fileHint = int(fromSq.File())
	}
	if typ == NoPieceType {
		return 0, fmt.Errorf("parseSAN: Couldn't deduce a piece type for `%s`", originalMove)
	}
	move = move.setPiece(GetPiece(typ, pos.Turn()))
	if p := pos.board.pieceAt(toSq); p != NoPiece {
		if p.Color() != pos.turn.Other() {
			if p.Type() == Rook && typ == King {
				// This may be a castle by other means.
				if p.Color() == White && pos.board.whiteKingSq == E1 && (toSq == A1 || toSq == H1) {
					move = move.setS1(E1)
					if toSq == A1 {
						move = move.addTag(QueenSideCastle)
						move = move.setS2(C1)
					} else {
						move = move.addTag(KingSideCastle)
						move = move.setS2(G1)
					}
					return parseSANTail(move, tail)
				} else if p.Color() == Black && pos.board.blackKingSq == E8 && (toSq == A8 || toSq == H8) {
					move = move.setS1(E8)
					if toSq == A8 {
						move = move.addTag(QueenSideCastle)
						move = move.setS2(C8)
					} else {
						move = move.addTag(KingSideCastle)
						move = move.setS2(G8)
					}
					return parseSANTail(move, tail)
				}
			}
			return 0, fmt.Errorf("parseSAN: `%s` apparently trying to capture own piece?", originalMove)
		}
		if !move.HasTag(Capture) {
			return 0, fmt.Errorf("parseSAN: `%s` appears to not capture but is moving onto a piece", originalMove)
		}
	}
	s1, err := findAndValidateFromSquare(move.piece(), toSq, fileHint, rankHint, pos)
	if err != nil {
		return 0, fmt.Errorf("parseSAN: for move `%s`, fAVS err: %s", originalMove, err)
	}
	move = move.setS1(s1)
	move = move.setS2(toSq)
	return parseSANTail(move, tail)
}

func parseSANQuality(s string) string {
	// TODO(barakmich): Perhaps add move comments about the quality of the move.
	// But for now, drop it.
	s = strings.ReplaceAll(s, "!", "")
	s = strings.ReplaceAll(s, "?", "")
	return s
}

func parseSANTail(move Move, s string) (Move, error) {
	if s == "" {
		return move, nil
	}
	s = parseSANQuality(s)
	if s == "" {
		return move, nil
	}
	if strings.Contains(s, "ep") {
		move = move.addTag(EnPassant)
		s = strings.ReplaceAll(s, "ep", "")
	}
	if strings.Contains(s, "e.p.") {
		move = move.addTag(EnPassant)
		s = strings.ReplaceAll(s, "e.p.", "")
	}
	if strings.Contains(s, "+") {
		move = move.addTag(Check)
		s = strings.ReplaceAll(s, "+", "")
	}
	if strings.Contains(s, "#") {
		move = move.addTag(Check)
		s = strings.ReplaceAll(s, "#", "")
		// TODO(barakmich): Validate checkmate iff we see this symbol
	}
	if idx := strings.Index(s, "="); idx != -1 {
		switch s[idx : idx+2] {
		case "=Q", "=q":
			move = move.setPromo(PromoQueen)
		case "=R", "=r":
			move = move.setPromo(PromoRook)
		case "=B", "=b":
			move = move.setPromo(PromoBishop)
		case "=N", "=n":
			move = move.setPromo(PromoKnight)
		default:
			return 0, fmt.Errorf("parseSANTail: detected promotion but can't parse `%s`", s)
		}
		s = s[:idx] + s[idx+2:]
	}
	if s != "" {
		return 0, fmt.Errorf("parseSANTail: Remaining tail characters: `%s`", s)
	}
	return move, nil
}

func findAndValidateFromSquare(p Piece, toSq Square, fileHint, rankHint int, pos *Position) (Square, error) {
	if p.Type() == Pawn {
		// Pawns can't move backwards, we have to know more.
		return findAndValidatePawnSquare(p, toSq, fileHint, pos)
	}
	currentPieces := pos.board.bbForPiece(p)
	if fileHint != -1 {
		currentPieces = currentPieces & bbFiles[fileHint]
	}
	if rankHint != -1 {
		currentPieces = currentPieces & bbRanks[rankHint]
	}
	if bits.OnesCount64(uint64(currentPieces)) == 1 {
		return validateFromBB(currentPieces, toSq)
	}
	occupied := pos.board.occupied()
	thisSq := bbForSquare(toSq)
	mask := bitboard(0b1)
	for i := 0; i < 64; i++ {
		if mask&currentPieces != 0 {
			moves := bbForPossiblePieceMoves(occupied, p.Type(), Square(i))
			if thisSq&moves != 0 {
				return validateFromBB(mask, toSq)
			}
		}
		mask = mask << 1
	}
	return NoSquare, errors.New("Can't find a potential piece to move")
}

func validateFromBB(fromSquareBB bitboard, toSq Square) (Square, error) {
	if fromSquareBB == 0 {
		return NoSquare, fmt.Errorf("validateFromBB: couldn't find any potential pieces to move to %s", toSq)
	}
	if bits.OnesCount64(uint64(fromSquareBB)) != 1 {
		return NoSquare, fmt.Errorf("validateFromBB: More than one apparent potential from piece moving to %s", toSq)
	}
	return bbGetFirstSquare(fromSquareBB), nil
}

func findAndValidatePawnSquare(p Piece, toSq Square, fileHint int, pos *Position) (Square, error) {
	currentPawns := pos.board.bbForPiece(p)
	var file File
	if fileHint == -1 {
		// This can't be a capture, according to SAN. So the file must be obvious.
		file = toSq.File()
	} else {
		file = File(fileHint)
	}
	var pawnBB bitboard
	if file > 7 {
		return Square(0), errors.New("How?")
	}
	filePawns := currentPawns & bbFiles[file]
	if pos.Turn() == White {
		pawnBB = filePawns & bbRanks[toSq.Rank()-1]
		if pawnBB == 0 && toSq.Rank() == Rank4 {
			pawnBB = filePawns & bbRank2
		}
	} else {
		pawnBB = filePawns & bbRanks[toSq.Rank()+1]
		if pawnBB == 0 && toSq.Rank() == Rank5 {
			pawnBB = filePawns & bbRank7
		}
	}
	return validateFromBB(pawnBB, toSq)
}
