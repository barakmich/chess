package bitflip

//go:generate go run ./asm/bitflipavo.go -out bitflip.s -stubs stub.go
//go:generate go run ./attacks/calcAttacks.go -out attacks.s -stubs attackstub.go
