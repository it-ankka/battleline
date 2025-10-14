package gamelogic

import (
	"math/rand/v2"
	"sort"
)

type Deck []Card

func (deck Deck) String() string {
	s := ""
	for i, c := range deck {
		if i > 0 {
			s = s + ", "
		}
		s = s + c.String()
	}
	return s
}

func (deck Deck) Copy() Deck {
	d := make([]Card, len(deck))
	copy(d, deck)
	return d
}

func (deck Deck) Pop() (Deck, Card) {
	c := deck[len(deck)-1]

	d := make(Deck, len(deck)-1)
	copy(d, deck[:len(deck)-1])

	return d, c
}

func (deck Deck) Shuffle() Deck {
	d := deck.Copy()

	for i := len(d) - 1; i > 0; i-- {
		j := rand.IntN(i)
		d[j], d[i] = d[i], d[j]
	}
	return d
}

func (deck Deck) SortBySuit() Deck {
	d := deck.Copy()

	sort.Slice(d, func(i, j int) bool {
		return (d[i].SuitSortingValue() < d[j].SuitSortingValue())
	})
	return d
}

func (deck Deck) SortByRank() Deck {
	d := deck.Copy()

	sort.Slice(d, func(i, j int) bool {
		return (d[i].Value < d[j].Value)
	})
	return d
}

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
