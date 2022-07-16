package main

import (
	"math/bits"

	. "github.com/mmcloughlin/avo/build"
	"github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/operand"
	"github.com/mmcloughlin/avo/reg"
)

const cm0 = 0x5555555555555555 // 01010101 ...
const cm1 = 0x3333333333333333 // 00110011 ...
const cm2 = 0x0f0f0f0f0f0f0f0f // 00001111 ...

const shufConstA = 0x0001020304050607
const shufConstB = 0x08090a0b0c0d0e0f

func prepMVars() ([3]reg.VecVirtual, reg.VecVirtual) {
	m0, m1, m2 := YMM(), YMM(), YMM()
	shuf := YMM()
	a, b, c, sa, sb := GP64(), GP64(), GP64(), GP64(), GP64()
	MOVQ(U64(cm0), a)
	MOVQ(U64(cm1), b)
	MOVQ(U64(cm2), c)
	MOVQ(U64(shufConstA), sa)
	MOVQ(U64(shufConstB), sb)
	MOVQ(a, m0.AsX())
	MOVQ(b, m1.AsX())
	MOVQ(c, m2.AsX())
	MOVQ(sb, shuf.AsX())
	MOVLHPS(m0.AsX(), m0.AsX())
	MOVLHPS(m1.AsX(), m1.AsX())
	MOVLHPS(m2.AsX(), m2.AsX())
	MOVLHPS(shuf.AsX(), shuf.AsX())
	MOVQ(sa, shuf.AsX())
	return [3]reg.VecVirtual{m0, m1, m2}, shuf
}

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
	TEXT("Reverse64AVX", NOSPLIT, "func(x uint64) uint64")
	Doc("Flips the bytes in x, MSB->LSB and vice-versa")
	x := Load(Param("x"), GP64())
	out := GP64()
	data := YMM()
	bits.Reverse64(0x1)
	PINSRQ(operand.U8(0x0), x, data.AsX())

	m, byteRevMask := prepMVars()

	reverseBitsInYMMBytes(data, m)
	outvec := reverseYMMBytes(data, byteRevMask)

	MOVQ(outvec.AsX(), out)
	Store(out, ReturnIndex(0))
	RET()
	Generate()
}
