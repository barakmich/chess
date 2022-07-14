package chess

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

// A Outcome is the result of a game.
type Outcome string

const (
	// NoOutcome indicates that a game is in progress or ended without a result.
	NoOutcome Outcome = "*"
	// WhiteWon indicates that white won the game.
	WhiteWon Outcome = "1-0"
	// BlackWon indicates that black won the game.
	BlackWon Outcome = "0-1"
	// Draw indicates that game was a draw.
	Draw Outcome = "1/2-1/2"
)

// String implements the fmt.Stringer interface
func (o Outcome) String() string {
	return string(o)
}

// A Method is the method that generated the outcome.
type Method uint8

const (
	// NoMethod indicates that an outcome hasn't occurred or that the method can't be determined.
	NoMethod Method = iota
	// Checkmate indicates that the game was won checkmate.
	Checkmate
	// Resignation indicates that the game was won by resignation.
	Resignation
	// DrawOffer indicates that the game was drawn by a draw offer.
	DrawOffer
	// Stalemate indicates that the game was drawn by stalemate.
	Stalemate
	// ThreefoldRepetition indicates that the game was drawn when the game
	// state was repeated three times and a player requested a draw.
	ThreefoldRepetition
	// FivefoldRepetition indicates that the game was automatically drawn
	// by the game state being repeated five times.
	FivefoldRepetition
	// FiftyMoveRule indicates that the game was drawn by the half
	// move clock being one hundred or greater when a player requested a draw.
	FiftyMoveRule
	// SeventyFiveMoveRule indicates that the game was automatically drawn
	// when the half move clock was one hundred and fifty or greater.
	SeventyFiveMoveRule
	// InsufficientMaterial indicates that the game was automatically drawn
	// because there was insufficient material for checkmate.
	InsufficientMaterial
)

// TagPair represents metadata in a key value pairing used in the PGN format.
type TagPair struct {
	Key   string
	Value string
}

// A Game represents a single chess game.
type Game struct {
	Notation             Notation
	tagPairs             map[string]string
	moves                []*Move
	positions            []*Position
	pos                  *Position
	outcome              Outcome
	method               Method
	ignoreAutomaticDraws bool
}

// NewGameFromPGN takes a reader and returns a function that creates
// the game to reflect the PGN data.  The PGN can use any
// move notation supported by this package.
// An error is returned if there is a problem parsing the PGN data.
func NewGameFromPGN(r io.Reader) (*Game, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	game, err := decodePGN(string(b))
	if err != nil {
		return nil, err
	}
	return game, nil
}

// FEN takes a string and returns a function that creates
// the game to reflect the FEN data.  Since FEN doesn't encode
// prior moves, the move list will be empty.
// An error is returned if there is a problem parsing the FEN data.
func NewGameFromFEN(fen string) (*Game, error) {
	g := NewGame()
	pos, err := decodeFEN(fen)
	if err != nil {
		return nil, err
	}
	pos.inCheck = isInCheck(pos)
	g.pos = pos
	g.positions = []*Position{pos}
	g.updatePosition()
	return g, nil
}

// NewGame defaults to returning a game in the standard
// opening position.  Options can be given to configure
// the game's initial state.
func NewGame() *Game {
	pos := StartingPosition()
	game := &Game{
		Notation:  SANNotation,
		moves:     []*Move{},
		pos:       pos,
		positions: []*Position{pos},
		outcome:   NoOutcome,
		method:    NoMethod,
	}
	return game
}

// Move updates the game with the given move.  An error is returned
// if the move is invalid or the game has already been completed.
func (g *Game) Move(m *Move) error {
	valid := moveSlice(g.ValidMoves()).find(m)
	if valid == nil {
		return fmt.Errorf("chess: invalid move %s", m)
	}
	g.moves = append(g.moves, valid)
	g.pos = g.pos.Update(valid)
	g.positions = append(g.positions, g.pos)
	g.updatePosition()
	return nil
}

// MoveStr decodes the given string, trying the obvious notations
// as though it were a PGN, and calls the Move function.
// To parse with a
// An error is returned if
// the move can't be decoded or the move is invalid.
func (g *Game) MoveStr(s string) error {
	m, err := g.pos.DecodeMove(s)
	if err != nil {
		return err
	}
	return g.Move(m)
}

// ValidMoves returns a list of valid moves in the
// current position.
func (g *Game) ValidMoves() []*Move {
	return g.pos.ValidMoves()
}

// Positions returns the position history of the game.
func (g *Game) Positions() []*Position {
	return append([]*Position(nil), g.positions...)
}

// Moves returns the move history of the game.
func (g *Game) Moves() []*Move {
	return append([]*Move(nil), g.moves...)
}

// TagPairs returns the game's tag pairs.
func (g *Game) TagPairs() []*TagPair {
	if g.tagPairs == nil {
		return nil
	}
	var out []*TagPair
	for k, v := range g.tagPairs {
		out = append(out, &TagPair{Key: k, Value: v})
	}
	return out
}

// Position returns the game's current position.
func (g *Game) Position() *Position {
	return g.pos
}

// Outcome returns the game outcome.
func (g *Game) Outcome() Outcome {
	return g.outcome
}

// Method returns the method in which the outcome occurred.
func (g *Game) Method() Method {
	return g.method
}

// FEN returns the FEN notation of the current position.
func (g *Game) FEN() string {
	return g.pos.String()
}

// String implements the fmt.Stringer interface and returns
// the game's PGN.
func (g *Game) String() string {
	return encodePGN(g)
}

// MarshalText implements the encoding.TextMarshaler interface and
// encodes the game's PGN.
func (g *Game) MarshalText() (text []byte, err error) {
	return []byte(encodePGN(g)), nil
}

// UnmarshalText implements the encoding.TextUnarshaler interface and
// assumes the data is in the PGN format.
func (g *Game) UnmarshalText(text []byte) error {
	game, err := decodePGN(string(text))
	if err != nil {
		return err
	}
	g.mergeInto(game)
	return nil
}

// Draw attempts to draw the game by the given method.  If the
// method is valid, then the game is updated to a draw by that
// method.  If the method isn't valid then an error is returned.
func (g *Game) Draw(method Method) error {
	switch method {
	case ThreefoldRepetition:
		if g.numOfRepetitions() < 3 {
			return errors.New("chess: draw by ThreefoldRepetition requires at least three repetitions of the current board state")
		}
	case FiftyMoveRule:
		if g.pos.halfMoveClock < 100 {
			return fmt.Errorf("chess: draw by FiftyMoveRule requires the half move clock to be at 100 or greater but is %d", g.pos.halfMoveClock)
		}
	case DrawOffer:
	default:
		return fmt.Errorf("chess: unsupported draw method %s", method.String())
	}
	g.outcome = Draw
	g.method = method
	return nil
}

// Resign resigns the game for the given color.  If the game has
// already been completed then the game is not updated.
func (g *Game) Resign(color Color) {
	if g.outcome != NoOutcome || color == NoColor {
		return
	}
	if color == White {
		g.outcome = BlackWon
	} else {
		g.outcome = WhiteWon
	}
	g.method = Resignation
}

// EligibleDraws returns valid inputs for the Draw() method.
func (g *Game) EligibleDraws() []Method {
	draws := []Method{DrawOffer}
	if g.numOfRepetitions() >= 3 {
		draws = append(draws, ThreefoldRepetition)
	}
	if g.pos.halfMoveClock >= 100 {
		draws = append(draws, FiftyMoveRule)
	}
	return draws
}

// AddTagPair adds or updates a tag pair with the given key and
// value and returns true if the value is overwritten.
func (g *Game) AddTagPair(k, v string) bool {
	if g.tagPairs == nil {
		g.tagPairs = make(map[string]string)
	}
	_, ok := g.tagPairs[k]
	g.tagPairs[k] = v
	return ok
}

// GetTagPair returns the tag pair for the given key or nil
// if it is not present.
func (g *Game) GetTagPair(k string) *TagPair {
	if g.tagPairs == nil {
		return nil
	}
	v, ok := g.tagPairs[k]
	if !ok {
		return nil
	}
	return &TagPair{Key: k, Value: v}
}

// RemoveTagPair removes the tag pair for the given key and
// returns true if a tag pair was removed.
func (g *Game) RemoveTagPair(k string) bool {
	if g.tagPairs == nil {
		return false
	}
	_, ok := g.tagPairs[k]
	delete(g.tagPairs, k)
	return ok
}

// MoveHistory is a move's result from Game's MoveHistory method.
// It contains the move itself, any comments, and the pre and post
// positions.
type MoveHistory struct {
	PrePosition  *Position
	PostPosition *Position
	Move         *Move
}

// MoveHistory returns the moves in order along with the pre and post
// positions and any comments.
func (g *Game) MoveHistory() []*MoveHistory {
	h := []*MoveHistory{}
	for i, p := range g.positions {
		if i == 0 {
			continue
		}
		m := g.moves[i-1]
		mh := &MoveHistory{
			PrePosition:  g.positions[i-1],
			PostPosition: p,
			Move:         m,
		}
		h = append(h, mh)
	}
	return h
}

func (g *Game) updatePosition() {
	method := g.pos.Status()
	if method == Stalemate {
		g.method = Stalemate
		g.outcome = Draw
	} else if method == Checkmate {
		g.method = Checkmate
		g.outcome = WhiteWon
		if g.pos.Turn() == White {
			g.outcome = BlackWon
		}
	}
	if g.outcome != NoOutcome {
		return
	}

	// five fold rep creates automatic draw
	if !g.ignoreAutomaticDraws && g.numOfRepetitions() >= 5 {
		g.outcome = Draw
		g.method = FivefoldRepetition
	}

	// 75 move rule creates automatic draw
	if !g.ignoreAutomaticDraws && g.pos.halfMoveClock >= 150 && g.method != Checkmate {
		g.outcome = Draw
		g.method = SeventyFiveMoveRule
	}

	// insufficient material creates automatic draw
	if !g.ignoreAutomaticDraws && !g.pos.board.hasSufficientMaterial() {
		g.outcome = Draw
		g.method = InsufficientMaterial
	}
}

func (g *Game) mergeInto(other *Game) {
	g.Notation = other.Notation
	g.tagPairs = other.tagPairs
	g.moves = other.moves
	g.positions = other.positions
	g.pos = other.pos
	g.outcome = other.outcome
	g.method = other.method
	g.ignoreAutomaticDraws = other.ignoreAutomaticDraws
}

func (g *Game) Clone() *Game {
	var newTags map[string]string

	if g.tagPairs != nil {
		newTags := make(map[string]string)
		for k, v := range g.tagPairs {
			newTags[k] = v
		}
	}

	return &Game{
		tagPairs:  newTags,
		Notation:  g.Notation,
		moves:     g.Moves(),
		positions: g.Positions(),
		pos:       g.pos,
		outcome:   g.outcome,
		method:    g.method,
	}
}

func (g *Game) numOfRepetitions() int {
	count := 0
	for _, pos := range g.Positions() {
		if g.pos.samePosition(pos) {
			count++
		}
	}
	return count
}
