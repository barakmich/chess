package bitflip

import (
	"math/bits"
)

func sqInt(rank, file int) int {
	return file<<3 | rank
}

func diaAttack(occupied uint64, sq int) uint64 {
	pos := bbForSquare(sq)
	dMask := bbDiagonals[sq]
	adMask := bbAntiDiagonals[sq]
	return linearAttack(occupied, pos, dMask) | linearAttack(occupied, pos, adMask)
}

func hvAttack(occupied uint64, sq int) uint64 {
	pos := bbForSquare(sq)
	rankMask := bbRanks[sq&0x7]
	fileMask := bbFiles[(sq >> 3)]
	return linearAttack(occupied, pos, rankMask) | linearAttack(occupied, pos, fileMask)
}

func linearAttack(occupied, pos, mask uint64) uint64 {
	oInMask := occupied & mask
	shiftedpos := pos << 1
	first := (oInMask - shiftedpos)
	revPosShifted := bits.Reverse64(pos) << 1
	revoInMask := bits.Reverse64(oInMask)
	sub := revoInMask - revPosShifted
	unrev := bits.Reverse64(sub)
	xored := first ^ unrev
	out := xored & mask
	//fmt.Printf("oInMaskout \t%016x\n", oInMask)
	//fmt.Printf("shiftedpos\t%016x\n", shiftedpos)
	//fmt.Printf("firstSubtract\t%016x\n", first)
	//fmt.Printf("revPosShifted\t%016x\n", revPosShifted)
	//fmt.Printf("revoInMask\t%016x\n", revoInMask)
	//fmt.Printf("subtractrev\t%016x\n", sub)
	//fmt.Printf("unreversed\t%016x\n", unrev)
	//fmt.Printf("out\t%016x\n", out)
	return out
}

const (
	bbFileA uint64 = 72340172838076673
	bbFileB uint64 = 144680345676153346
	bbFileC uint64 = 289360691352306692
	bbFileD uint64 = 578721382704613384
	bbFileE uint64 = 1157442765409226768
	bbFileF uint64 = 2314885530818453536
	bbFileG uint64 = 4629771061636907072
	bbFileH uint64 = 9259542123273814144

	bbRank1 uint64 = 255
	bbRank2 uint64 = 65280
	bbRank3 uint64 = 16711680
	bbRank4 uint64 = 4278190080
	bbRank5 uint64 = 1095216660480
	bbRank6 uint64 = 280375465082880
	bbRank7 uint64 = 71776119061217280
	bbRank8 uint64 = 18374686479671623680
)

// TODO make method on Square
func bbForSquare(sq int) uint64 {
	return uint64(0b1 << sq)
}

var (
	bbFiles = [8]uint64{bbFileA, bbFileB, bbFileC, bbFileD, bbFileE, bbFileF, bbFileG, bbFileH}
	bbRanks = [8]uint64{bbRank1, bbRank2, bbRank3, bbRank4, bbRank5, bbRank6, bbRank7, bbRank8}

	bbDiagonals = [64]uint64{9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 2151686160, 8405024, 32832, 128, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 2151686160, 8405024, 32832, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 2151686160, 8405024, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 2151686160, 577588855528488960, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 550831656968, 288794425616760832, 577588855528488960, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 141012904183812, 144396663052566528, 288794425616760832, 577588855528488960, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745, 36099303471055874, 72057594037927936, 144396663052566528, 288794425616760832, 577588855528488960, 1155177711073755136, 2310355422147575808, 4620710844295151872, 9241421688590303745}

	bbAntiDiagonals = [64]uint64{1, 258, 66052, 16909320, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 258, 66052, 16909320, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 66052, 16909320, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 16909320, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 4328785936, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 1161999622361579520, 1108169199648, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 1161999622361579520, 2323998145211531264, 283691315109952, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 1161999622361579520, 2323998145211531264, 4647714815446351872, 72624976668147840, 145249953336295424, 290499906672525312, 580999813328273408, 1161999622361579520, 2323998145211531264, 4647714815446351872, 9223372036854775808}
)