package game

type Formation int

const (
	FormationFray Formation = iota
	FormationSkirmish
	FormationColumn
	FormationSquare
	FormationWedge
)

func (d Deck) IsAllSameSuit() bool {
	suit := d[0].Suit
	for _, c := range d {
		if suit != c.Suit {
			return false
		}
		suit = c.Suit
	}
	return true
}

// TODO
// func (d Deck) IsRun() bool {
// 	sortedDeck := d.SortByRank()

// 	val := sortedDeck[0].Value
// 	for _, c := range d {
// 		if (suit != c.suit) {
// 			return false
// 		}
// 		suit = c.suit
// 	}
// 	return true
// }

func (d *Deck) GetFormation() Formation {
	return FormationFray
}
