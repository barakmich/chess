package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

const shufConstA = 0x0001020304050607

const shufConstAle = 0x0706050403020100
const shufConstB = 0x08090a0b0c0d0e0f
const shufConstBle = 0x0f0e0d0c0b0a0908

var cm0 = []uint64{0x0f0f0f0f0f0f0f0f, 0x0f0f0f0f0f0f0f0f}

var cm1 = []uint64{0x0109050d030b070f, 0x0008040c020a060e}   // Big Endian
var cm1le = []uint64{0x0f070b030d050901, 0x0e060a020c040800} // Little Endian

var cm2 = []uint64{0x109050d030b070f0, 0x008040c020a060e0}   // Big Endian
var cm2le = []uint64{0xf070b030d0509010, 0xe060a020c0408000} // Little Endian

func reverse64(data reg.VecVirtual, rev [3]reg.VecVirtual, shuf reg.VecVirtual) {
	reverseBits(data, rev)
	VPSHUFB(shuf, data, data)
}

func reverseBits(data reg.VecVirtual, rev [3]reg.VecVirtual) {
	tmp := XMM()
	VPAND(rev[0], data, tmp)
	VPANDN(data, rev[0], data)
	VPSRLD(U8(0x4), data, data)
	VPSHUFB(tmp, rev[2], tmp)
	VPSHUFB(data, rev[1], data)
	VPOR(data, tmp, data)
}

func main() {
	bytes := GLOBL("bytes", RODATA|NOPTR)
	DATA(0, U64(cm0[0]))
	DATA(8, U64(cm0[1]))
	DATA(16, U64(cm1le[1]))
	DATA(24, U64(cm1le[0]))
	DATA(32, U64(cm2le[1]))
	DATA(40, U64(cm2le[0]))
	DATA(48, U64(shufConstA))
	DATA(56, U64(shufConstB))

	// Rank, Diag, File, AntiDiag -- here's why: the lanes match
	// Returns Ortho, Diag
	TEXT("CalcAttacks", NOSPLIT, "func(occupied uint64, location uint64, angles [4]uint64) (uint64, uint64)")
	occ := Load(Param("occupied"), GP64())
	pos := Load(Param("location"), GP64())
	rank := Load(Param("angles").Index(0), GP64())
	file := Load(Param("angles").Index(1), GP64())
	diag := Load(Param("angles").Index(2), GP64())
	antidiag := Load(Param("angles").Index(3), GP64())
	bytesPtr := Mem{Base: GP64()}
	LEAQ(bytes, bytesPtr.Base)
	shuf := XMM()
	rev := [3]reg.VecVirtual{XMM(), XMM(), XMM()}
	maskLeft, maskRight := XMM(), XMM()
	Comment("Load Constants")
	MOVAPD(bytesPtr.Offset(0), rev[0])
	MOVAPD(bytesPtr.Offset(16), rev[1])
	MOVAPD(bytesPtr.Offset(32), rev[2])
	MOVAPD(bytesPtr.Offset(48), shuf)
	Comment("Load Masks")
	MOVQ(diag, maskLeft)
	MOVQ(antidiag, maskRight)
	tmpl, tmpr := XMM(), XMM()
	MOVQ(rank, tmpl)
	MOVQ(file, tmpr)
	SHUFPD(U8(0), tmpl, maskLeft)
	SHUFPD(U8(0), tmpr, maskRight)
	dataL, dataR := XMM(), XMM()
	nonrevL, nonrevR := XMM(), XMM()
	posX := XMM()
	posShift := XMM()
	Comment("Prep position vars")
	MOVQ(pos, posX)
	MOVDDUP(posX, posX)
	MOVAPD(posX, posShift)
	PSLLQ(U8(1), posShift)
	Comment("Prep data vars")
	MOVQ(occ, dataL)
	MOVDDUP(dataL, dataL)
	MOVAPD(dataL, dataR)
	PAND(maskLeft, dataL)
	PAND(maskRight, dataR)
	Comment("Subtract first half")
	VPSUBQ(posShift, dataL, nonrevL)
	VPSUBQ(posShift, dataL, nonrevR)
	Comment("Reverse pos")
	reverse64(posX, rev, shuf)
	Comment("Shift pos")
	PSLLQ(U8(1), posX)
	Comment("Reverse dataL")
	reverse64(dataL, rev, shuf)
	Comment("Reverse dataR")
	reverse64(dataR, rev, shuf)
	Comment("Subtract second half")
	VPSUBQ(posX, dataL, dataL)
	VPSUBQ(posX, dataR, dataR)
	Comment("Unreverse dataL")
	reverse64(dataL, rev, shuf)
	Comment("Unreverse dataR")
	reverse64(dataR, rev, shuf)
	Comment("Finish")
	PXOR(nonrevL, dataL)
	PXOR(nonrevR, dataR)
	PAND(maskLeft, dataL)
	PAND(maskRight, dataR)
	out := XMM()
	PXOR(out, out)
	POR(dataL, out)
	POR(dataR, out)
	Comment("Extract")
	outOrtho, outDiag := GP64(), GP64()
	PEXTRQ(U8(1), out, outOrtho)
	MOVQ(out, outDiag)
	Store(outOrtho, ReturnIndex(0))
	Store(outDiag, ReturnIndex(1))
	RET()
	Generate()
	//PSHUFB(shuf, out)
}
