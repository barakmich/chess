package main

import (
	"math/bits"

	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

//const cm0 = 0x5555555555555555 // 01010101 ...
//const cm1 = 0x3333333333333333 // 00110011 ...
//const cm2 = 0x0f0f0f0f0f0f0f0f // 00001111 ...

const shufConstA = 0x0001020304050607
const shufConstB = 0x08090a0b0c0d0e0f

var cm0 = []uint64{0x0f0f0f0f0f0f0f0f, 0x0f0f0f0f0f0f0f0f}

var cm1 = []uint64{0x0109050d030b070f, 0x0008040c020a060e}   // Big Endian
var cm1le = []uint64{0x0f070b030d050901, 0x0e060a020c040800} // Little Endian

var cm2 = []uint64{0x109050d030b070f0, 0x008040c020a060e0}   // Big Endian
var cm2le = []uint64{0xf070b030d0509010, 0xe060a020c0408000} // Little Endian

func reverseBitsInYMMBytes(data reg.VecVirtual, m [3]reg.VecVirtual) {

	roundout := [3]reg.VecVirtual{YMM(), YMM(), data}
	roundin := [3]reg.VecVirtual{data, roundout[0], roundout[1]}
	tmp := GP64()

	shift := reg.X14
	for i := 0; i < 3; i++ {
		cp := YMM()
		MOVQ(U64(1<<i), tmp)
		MOVQ(tmp, shift)
		andTmpA, andTmpB, andTmpC, andTmpD := reg.Y10, reg.Y11, reg.Y12, reg.Y13
		VMOVDQA(roundin[i], cp)
		VPSRLQ(shift, roundin[i], andTmpA)
		VPAND(m[i], cp, andTmpB)
		VPAND(m[i], andTmpA, andTmpC)
		VPSLLQ(shift, andTmpB, andTmpD)
		VPOR(andTmpC, andTmpD, roundout[i])
	}
}

func reverseYMMBytes(data reg.VecVirtual, byteRevMask reg.VecVirtual) reg.VecVirtual {
	out := YMM()
	VPSHUFB(byteRevMask, data, out) // AVX2
	return out
}

func main() {
	bytes := GLOBL("bytes", RODATA|NOPTR)
	DATA(0, U64(cm0[0]))
	DATA(8, U64(cm0[1]))
	DATA(16, U64(cm0[0]))
	DATA(24, U64(cm0[1]))
	DATA(32, U64(cm1le[1]))
	DATA(40, U64(cm1le[0]))
	DATA(48, U64(cm1le[1]))
	DATA(56, U64(cm1le[0]))
	DATA(64, U64(cm2le[1]))
	DATA(72, U64(cm2le[0]))
	DATA(80, U64(cm2le[1]))
	DATA(88, U64(cm2le[0]))

	TEXT("Reverse64AVX", NOSPLIT, "func(x uint64) uint64")
	Doc("Flips the bytes in x, MSB->LSB and vice-versa")
	x := Load(Param("x"), GP64())
	out := GP64()
	data := YMM()
	bits.Reverse64(0x1)
	MOVQ(x, data.AsX())
	bytesPtr := Mem{Base: GP64()}
	LEAQ(bytes, bytesPtr.Base)

	// TODO(load from data)
	shuf := YMM()
	sa, sb := GP64(), GP64()
	MOVQ(U64(shufConstA), sa)
	MOVQ(U64(shufConstB), sb)
	MOVQ(sb, shuf.AsX())
	MOVLHPS(shuf.AsX(), shuf.AsX())
	MOVQ(sa, shuf.AsX())

	m0, m1, m2 := YMM(), YMM(), YMM()

	//a := GP64()
	//MOVQ(U64(cm0[0]), a)
	//MOVQ(a, m0.AsX())
	//VPBROADCASTQ(m0.AsX(), m0)

	//b := GP64()
	//c := GP64()
	//tmpA := YMM()
	//MOVQ(U64(cm1[1]), b)
	//MOVQ(U64(cm1[0]), c)
	//MOVQ(b, tmpA.AsX())
	//MOVQ(c, m1.AsX())
	//VMOVLHPS(m1.AsX(), tmpA.AsX(), m1.AsX())
	//VINSERTI128(U8(0x1), m1.AsX(), m1, m1)

	VMOVDQA(bytesPtr.Offset(0), m0)
	VMOVDQA(bytesPtr.Offset(32), m1)
	VMOVDQA(bytesPtr.Offset(64), m2)
	tmp := YMM()
	VPAND(m0, data, tmp)
	VPANDN(data, m0, data)
	VPSRLD(U8(0x4), data, data)
	VPSHUFB(tmp, m2, tmp)
	VPSHUFB(data, m1, data)
	VPOR(data, tmp, data)

	//m, byteRevMask := prepMVars()

	outvec := reverseYMMBytes(data, shuf)

	MOVQ(outvec.AsX(), out)
	Store(out, ReturnIndex(0))
	RET()
	Generate()
}
