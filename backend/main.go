package main

import (
	"fmt"
	. "github.com/it-ankka/battleline/cards"
)

func CreateStartingDeck() Deck {
	var deck = Deck{}
	for _, s := range Suits {
		for i := range 10 {
			deck = append(deck, Card{
				Suit:  s,
				Value: i + 1,
			})
		}
	}
	return deck
}

type Lane struct {
	cards   Deck
	claimed int
}

type Lanes [9][2]Lane

type PlayerState struct {
	Hand Deck `json:"hand"`
}

func main() {
	var deck = CreateStartingDeck()
	deck = deck.Shuffle()
	for _, c := range deck {
		fmt.Printf("%s\n", c)
	}
}
