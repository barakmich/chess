package chess

import (
	"compress/bzip2"
	"context"
	"os"
	"strings"
	"testing"
)

func runParallelBigScanner(t testing.TB) {
	f, err := os.Open("fixtures/lichess_head_50k_2022_06.pgn.bz2")
	if err != nil {
		t.Fatal(err)
		return
	}
	defer f.Close()
	bz := bzip2.NewReader(f)
	if b, ok := t.(*testing.B); ok {
		b.StartTimer()
	}

	scan := NewParallelScanner(bz)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gamesChan := make(chan *Game)
	go scan.Begin(ctx, gamesChan)

	whiteWins := 0
	blackWins := 0
	total := 0
	for game := range gamesChan {
		total += 1
		if total%500 == 0 {
			t.Logf("total: %d", total)
		}
		pair := game.GetTagPair("Site")
		if pair == nil {
			t.Fatal("No Site tag in PGN")
		}
		if !strings.HasPrefix(pair.Value, "https://lichess") {
			t.Fatal("Site tag not from lichess")
		}
		switch game.Outcome() {
		case WhiteWon:
			whiteWins += 1
		case BlackWon:
			blackWins += 1
		}
	}
	if whiteWins != 1214 {
		t.Errorf("Apparent White wins doesn't match: got %d expected %d", whiteWins, 1214)
	}
	if blackWins != 1189 {
		t.Errorf("Apparent Black wins doesn't match: got %d expected %d", blackWins, 1189)
	}
}

func TestParallelBigScanner(t *testing.T) {
	runParallelBigScanner(t)
}

func BenchmarkParallelBigScanner(b *testing.B) {
	for n := 0; n < b.N; n++ {
		b.StopTimer()
		runParallelBigScanner(b)
	}
}
