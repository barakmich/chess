package bitflip

//go:generate go run ./asm/bitflipavo.go -out bitflip_amd64.s -stubs bitflip_amd64.go
//go:generate go run ./attacks/calcAttacks.go -out attacks_amd64.s -stubs attacks_amd64.go
