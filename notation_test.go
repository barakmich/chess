package chess

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type validNotationTest struct {
	Pos1        *Position
	Pos2        *Position
	AlgText     string
	LongAlgText string
	UCIText     string
	Description string
}

func TestValidDecoding(t *testing.T) {
	f, err := os.Open("fixtures/valid_notation_tests.json")
	if err != nil {
		t.Fatal(err)
		return
	}

	validTests := []validNotationTest{}
	if err := json.NewDecoder(f).Decode(&validTests); err != nil {
		t.Fatal(err)
		return
	}

	for _, test := range validTests {
		for i, n := range []Notation{SANNotation, LongAlgebraicNotation, UCINotation} {
			var moveText string
			switch i {
			case 0:
				moveText = test.AlgText
			case 1:
				moveText = test.LongAlgText
			case 2:
				moveText = test.UCIText
			}
			m, err := test.Pos1.DecodeMove(moveText, n)
			if err != nil {
				movesStrList := []string{}
				for _, m := range test.Pos1.ValidMoves() {
					s := test.Pos1.EncodeMove(m, n)
					movesStrList = append(movesStrList, s)
				}
				t.Fatalf("starting from board \n%s\n expected move to be valid error - %s %s\n", test.Pos1.board.Draw(), err, strings.Join(movesStrList, ","))
			}
			postPos := test.Pos1.Update(m)
			if test.Pos2.String() != postPos.String() {
				t.Fatalf("starting from board \n%s\n after move %s\n (%s, %d) expected board to be %s\n%s\n but was %s\n%s\n",
					test.Pos1.board.Draw(), m.String(), moveText, i, test.Pos2.String(),
					test.Pos2.board.Draw(), postPos.String(), postPos.board.Draw())
			}
		}

	}
}

type notationDecodeTest struct {
	N       Notation
	Pos     *Position
	Text    string
	PostPos *Position
}

var (
	invalidDecodeTests = []notationDecodeTest{
		{
			// opening for white
			N:    SANNotation,
			Pos:  unsafeFEN("rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"),
			Text: "e5",
		},
		{
			// http://en.lichess.org/W91M4jms#14
			N:    SANNotation,
			Pos:  unsafeFEN("rn1qkb1r/pp3ppp/2p1pn2/3p4/2PP4/2NQPN2/PP3PPP/R1B1K2R b KQkq - 0 7"),
			Text: "Nd7",
		},
		{
			// http://en.lichess.org/W91M4jms#17
			N:       SANNotation,
			Pos:     unsafeFEN("r2qk2r/pp1n1ppp/2pbpn2/3p4/2PP4/1PNQPN2/P4PPP/R1B1K2R w KQkq - 1 9"),
			Text:    "O-O-O-O",
			PostPos: unsafeFEN("r2qk2r/pp1n1ppp/2pbpn2/3p4/2PP4/1PNQPN2/P4PPP/R1B2RK1 b kq - 0 9"),
		},
		{
			// http://en.lichess.org/W91M4jms#23
			N:    SANNotation,
			Pos:  unsafeFEN("3r1rk1/pp1nqppp/2pbpn2/3p4/2PP4/1PNQPN2/PB3PPP/3RR1K1 b - - 5 12"),
			Text: "dx4",
		},
		{
			// should not assume pawn for unknown piece type "n"
			N:    SANNotation,
			Pos:  unsafeFEN("rnbqkbnr/pppp1ppp/8/4p3/4P3/8/PPPP1PPP/RNBQKBNR w KQkq e6 0 2"),
			Text: "nf3",
		},
		{
			// disambiguation should not allow for this since it is not a capture
			N:    SANNotation,
			Pos:  unsafeFEN("rnbqkbnr/ppp1pppp/8/3p4/3P4/8/PPP1PPPP/RNBQKBNR w KQkq - 0 2"),
			Text: "bf4",
		},
	}
)

func TestInvalidDecoding(t *testing.T) {
	for _, test := range invalidDecodeTests {
		if _, err := test.Pos.DecodeMove(test.Text, test.N); err == nil {
			t.Fatalf("starting from board\n%s\n expected move notation %s to be invalid", test.Pos.board.Draw(), test.Text)
		}
	}
}

func BenchmarkValidAlgebraicDecoding(b *testing.B) {
	f, err := os.Open("fixtures/valid_notation_tests.json")
	if err != nil {
		b.Fatal(err)
		return
	}

	validTests := []validNotationTest{}
	if err := json.NewDecoder(f).Decode(&validTests); err != nil {
		b.Fatal(err)
		return
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		for _, test := range validTests {
			_, err := test.Pos1.DecodeMove(test.AlgText, SANNotation)
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}
