package main

import (
	"fmt"

	. "github.com/it-ankka/battleline/game"
)

func main() {
	g := NewGameState()
	fmt.Printf("Player 1 Hand: %s\n", g.PlayerHands[0].SortByRank().String())
	fmt.Printf("Player 2 Hand: %s\n", g.PlayerHands[1].SortByRank().String())
	fmt.Printf("\nDrawDeck: %s\n", g.DrawDeck.SortByRank().String())
}
