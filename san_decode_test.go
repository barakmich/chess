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
			N:        SANNotation,
			Pos:      unsafeFEN("r2qk1nr/pp3ppp/2n1p3/1B1pPb2/1b1P4/2N1B3/PP2NPPP/R2QK2R b KQkq - 3 9"),
			Text:     "Ne7",
			MoveText: "g8e7-0",
		},
		{
			N:    SANNotation,
			Pos:  unsafeFEN("r2q2kr/1p3pbp/p1npbnp1/3Np3/4P3/PN2BB2/1PP2PPP/R2Q2KR b - - 5 12"),
			Text: "Rc8",
		},
		{
			N:        SANNotation,
			Pos:      unsafeFEN("r3k2r/p1p1npbp/1pn1p1p1/4P3/4PBP1/5N2/PPP4P/R2K1B1R b kq - 2 12"),
			Text:     "O-O-O+",
			MoveText: "e8c8-18",
		},
	}
)

func TestSANParseFailures(t *testing.T) {
	for _, p := range validParseTests {
		m, err := parseSAN(p.Text, p.Pos)
		if err != nil {
			t.Fatalf("starting from board\n%s\n expected move notation %s to be valid:\n %s", p.Pos.board.Draw(), p.Text, err)
		}
		if p.MoveText != "" && p.MoveText != m.StringWithTags() {
			t.Fatalf("staring from board\n%s\nexpected move %s to be `%s` but got `%s`", p.Pos.board.Draw(), p.Text, p.MoveText, m.StringWithTags())
		}
	}
}
