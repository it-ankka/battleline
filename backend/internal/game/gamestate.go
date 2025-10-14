package game

import "math/rand/v2"

type GameState struct {
	DrawDeck     Deck
	ActivePlayer int
	Lanes        GameLanes
	PlayerHands  [2]Deck
}

type PrivateGameState struct {
	ActivePlayer     int       `json:"activePlayer"`
	Lanes            GameLanes `json:"lanes"`
	PlayerHand       Deck      `json:"playerState"`
	DrawDeckSize     int       `json:"drawDeckSize"`
	OpponentHandSize int       `json:"opponentHandSize"`
}

func NewGameState() GameState {
	var g GameState

	g.ActivePlayer = rand.IntN(2)
	g.DrawDeck = CreateStartingDeck()
	g.DrawDeck = g.DrawDeck.Shuffle()

	for range 7 {
		for i := range g.PlayerHands {
			if len(g.DrawDeck) > 0 {
				newDeck, c := g.DrawDeck.Pop()
				g.DrawDeck = newDeck
				g.PlayerHands[i] = append(g.PlayerHands[i], c)
			}
		}
	}

	return g
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
