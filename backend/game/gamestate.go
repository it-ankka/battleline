package game

import "math/rand/v2"

type GameState struct {
	DrawDeck     Deck      `json:"drawDeck"`
	ActivePlayer int       `json:"activePlayer"`
	Lanes        GameLanes `json:"lanes"`
	PlayerHands  [2]Deck   `json:"players"`
}

type PrivateGameState struct {
	ActivePlayer     int       `json:"activePlayer"`
	DrawDeckSize     int       `json:"drawDeckSize"`
	OpponentHandSize int       `json:"opponentHandSize"`
	PlayerHand       Deck      `json:"playerState"`
	Lanes            GameLanes `json:"lanes"`
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
