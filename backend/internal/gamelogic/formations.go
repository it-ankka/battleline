package gamelogic

type Formation int

const (
	FormationNone Formation = iota
	FormationFray
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
	}
	return true
}

func (d Deck) IsAllSameValue() bool {
	val := d[0].Value
	for _, c := range d {
		if val != c.Value {
			return false
		}
	}
	return true
}

func (d Deck) IsStraight() bool {
	sortedDeck := d.SortByRank()

	val := sortedDeck[0].Value
	for _, c := range sortedDeck[1:] {
		if val != c.Value-1 {
			return false
		}
		val = c.Value
	}
	return true
}

func (d Deck) GetFormation() Formation {
	sameSuit, sameValue, straight := d.IsAllSameSuit(), d.IsAllSameValue(), d.IsStraight()
	if len(d) < 3 {
		return FormationNone
	} else if sameSuit && straight {
		return FormationWedge
	} else if sameValue {
		return FormationSquare
	} else if sameSuit {
		return FormationColumn
	} else if straight {
		return FormationSkirmish
	}
	return FormationFray
}
