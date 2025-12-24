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
	if len(deck)-1 < 0 {
		return deck, Card{}
	}
	c := deck[len(deck)-1]

	d := make(Deck, len(deck)-1)
	copy(d, deck[:len(deck)-1])

	return d, c
}

func (deck Deck) FindCardIdx(card Card) int {
	for i, c := range deck {
		if c.Suit == card.Suit && c.Value == card.Value {
			return i
		}
	}
	return -1
}

func (deck Deck) RemoveAt(idx int) Deck {
	if idx < 0 || idx >= len(deck) {
		return deck
	}

	d := make(Deck, 0, len(deck)-1)
	d = append(d, deck[:idx]...)
	d = append(d, deck[idx+1:]...)

	return d
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

func CreateTroopDeck() Deck {
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

// Returns a integer where 100s are the formation value and the sum of the cards is the 10s and 1s
func (deck Deck) GetBestPossibleTotalValue(bestFormationValue int, possibleCards Deck) int {
	if len(deck) == MaxCardsPerSide || len(possibleCards) == 0 {
		return max(deck.GetTotalValue(), bestFormationValue)
	}

	remainingPossible, nextCard := possibleCards.Pop()
	formationWithCard := append(deck, nextCard).GetBestPossibleTotalValue(bestFormationValue, remainingPossible)
	if formationWithCard > bestFormationValue {
		bestFormationValue = formationWithCard
	}

	formationWithoutCard := deck.GetBestPossibleTotalValue(bestFormationValue, remainingPossible)
	if formationWithoutCard > bestFormationValue {
		bestFormationValue = formationWithoutCard
	}

	return bestFormationValue
}
