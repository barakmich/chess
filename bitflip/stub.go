// Code generated by command: go run bitflipavo.go -out bitflip.s -stubs stub.go. DO NOT EDIT.

package bitflip

// Flips the bytes in x, MSB->LSB and vice-versa
func Reverse64AVX(x uint64) uint64