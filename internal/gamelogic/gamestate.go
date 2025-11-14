package gamelogic

import (
	"math/rand/v2"
)

type TurnPhase int

const (
	PlacementPhase TurnPhase = iota
	ClaimPhase
	DrawPhase
)

var phaseNames = map[TurnPhase]string{
	PlacementPhase: "placement",
	ClaimPhase:     "claim",
	DrawPhase:      "draw",
}

func (tp TurnPhase) String() string {
	return phaseNames[tp]
}

type GameState struct {
	ActivePlayer int
	TurnPhase    TurnPhase
	DrawDeck     Deck
	Lanes        GameLanes
	PlayerHands  [2]Deck
}

type PrivateGameState struct {
	ActivePlayer     int       `json:"activePlayer"`
	TurnPhase        string    `json:"turnPhase"`
	Lanes            GameLanes `json:"lanes"`
	PlayerHand       Deck      `json:"playerState"`
	DrawDeckSize     int       `json:"drawDeckSize"`
	OpponentHandSize int       `json:"opponentHandSize"`
}

func NewGameState() *GameState {
	gs := &GameState{}

	gs.ActivePlayer = rand.IntN(2)
	gs.DrawDeck = CreateStartingDeck()
	gs.DrawDeck = gs.DrawDeck.Shuffle()

	for range 7 {
		for i := range gs.PlayerHands {
			if len(gs.DrawDeck) > 0 {
				newDeck, c := gs.DrawDeck.Pop()
				gs.DrawDeck = newDeck
				gs.PlayerHands[i] = append(gs.PlayerHands[i], c)
			}
		}
	}

	return gs
}

func (gs *GameState) GetPrivateGameState(playerIdx int) *PrivateGameState {
	var opponentIdx = 0
	if playerIdx == 0 {
		opponentIdx = 1
	} else {
		playerIdx = 1
	}

	return &PrivateGameState{
		ActivePlayer:     gs.ActivePlayer,
		Lanes:            gs.Lanes,
		PlayerHand:       gs.PlayerHands[playerIdx],
		DrawDeckSize:     len(gs.DrawDeck),
		OpponentHandSize: len(gs.PlayerHands[opponentIdx]),
	}
}
