package chess

import "testing"

var opening = []string{"Nf3", "e5", "Nxe5", "Qe7", "Nc4", "d5", "Nbc3", "dxc4"}

func runTestOpeningDecode(t testing.TB) {
	game := NewGame()
	for _, m := range opening {
		move, err := parseSAN(m, game.Position())
		if err != nil {
			t.Fatalf("Can't parse %s: %s", m, err)
		}
		err = game.Move(move)
		if err != nil {
			t.Fatalf("Couldn't move %s: %s", m, err)
		}
	}
}

func TestSANOpeningDecode(t *testing.T) {
	runTestOpeningDecode(t)
}

func BenchmarkSANOpeningDecode(b *testing.B) {
	for n := 0; n < b.N; n++ {
		runTestOpeningDecode(b)
	}
}

func BenchmarkOldOpeningDecode(b *testing.B) {
	for n := 0; n < b.N; n++ {
		game := NewGame()
		for _, m := range opening {
			err := game.MoveStr(m)
			if err != nil {
				b.Fatalf("Couldn't move %s", m)
			}
		}
	}
}

// "Bxc5" for position r2q1rk1/pppn1pp1/5n2/2bP3p/2PP1Pb1/2NQB1P1/PP2B2P/R3K1NR w KQ - 0 11 on move 11
var (
	validParseTests = []notationDecodeTest{
		{
			N:    SANNotation,
			Pos:  unsafeFEN("r2q2kr/1p3pbp/p1npbnp1/3Np3/4P3/PN2BB2/1PP2PPP/R2Q2KR b - - 5 12"),
			Text: "Rc8",
		},
		{
			N:    SANNotation,
			Pos:  unsafeFEN("6k1/3n2p1/4NpP1/p1pr1P2/2P1R2P/P1P5/4K3/8 w - - 1 36"),
			Text: "c4",
		},
	}
)

func TestValidParseFailures(t *testing.T) {
	for _, p := range validParseTests {
		if _, err := parseSAN(p.Text, p.Pos); err != nil {
			t.Fatalf("starting from board\n%s\n expected move notation %s to be valid", p.Pos.board.Draw(), p.Text)
		}
	}
}
