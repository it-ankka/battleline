package cards

import (
	"math/rand/v2"
	"sort"
)

type Deck []Card

func (deck Deck) Shuffle() Deck {
	d := make([]Card, len(deck))
	copy(d, deck)

	var cur = len(d) - 1
	var t int

	for cur > 0 {
		t = rand.IntN(cur)
		temp := d[t]
		d[t] = d[cur]
		d[cur] = temp
		cur -= 1
	}
	return d
}

func (deck Deck) SortBySuit() Deck {
	d := make([]Card, len(deck))
	copy(d, deck)

	sort.Slice(d, func(i, j int) bool {
		return (d[i].SuitSortingValue() < d[j].SuitSortingValue())
	})
	return d
}

func (deck Deck) SortByRank() Deck {
	d := make([]Card, len(deck))
	copy(d, deck)

	sort.Slice(d, func(i, j int) bool {
		return (d[i].Value < d[j].Value)
	})
	return d
}
