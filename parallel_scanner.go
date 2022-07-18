package chess

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"
)

type ParallelScanner struct {
	scanr *bufio.Scanner
	err   error
}

// NewParallelScanner returns a new scanner that decodes PGN in parallel.
func NewParallelScanner(r io.Reader) *ParallelScanner {
	scanr := bufio.NewScanner(r)
	return &ParallelScanner{scanr: scanr}
}

func (s *ParallelScanner) Begin(ctx context.Context, output chan *Game) error {
	if s.err == io.EOF {
		return s.err
	}
	s.err = nil
	var sb strings.Builder
	state := notInPGN
	var wg sync.WaitGroup
	work := make(chan string)
	for i := 0; i < runtime.NumCPU(); i++ {
		go parseGameWorker(i, work, output, &wg)
		wg.Add(1)
	}
OUTER:
	for {
		select {
		case <-ctx.Done():
			break OUTER
		default:
			scan := s.scanr.Scan()
			if !scan {
				s.err = s.scanr.Err()
				// err is nil if io.EOF
				if s.err == nil {
					s.err = io.EOF
				}
				break OUTER
			}
			line := strings.TrimSpace(s.scanr.Text())
			isTagPair := strings.HasPrefix(line, "[")
			isMoveSeq := strings.HasPrefix(line, "1. ")
			switch state {
			case notInPGN:
				if !isTagPair {
					break
				}
				state = inTagPairs
				sb.WriteString(line + "\n")
			case inTagPairs:
				if isMoveSeq {
					state = inMoves
				}
				sb.WriteString(line + "\n")
			case inMoves:
				if line == "" {
					work <- sb.String()
					sb.Reset()
					state = notInPGN
				}
				sb.WriteString(line + "\n")
			}
		}
	}
	close(work)
	wg.Wait()
	close(output)
	return ctx.Err()
}

// Err returns an error encountered during scanning.
// Typically this will be a PGN parsing error or an
// io.EOF.
func (s *ParallelScanner) Err() error {
	return s.err
}

func parseGameWorker(i int, work chan string, out chan *Game, wg *sync.WaitGroup) {
	for {
		s, ok := <-work
		if !ok {
			break
		}
		game, err := decodePGN(s, false)
		if err != nil {
			fmt.Println(i, "err:", err)
		}
		if game != nil {
			out <- game
		}
	}
	wg.Done()
}
