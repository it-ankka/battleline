package gamelogic

const MaxCardsPerSide = 3
const (
	NotClaimed = iota
	ClaimedByPlayerOne
	ClaimedByPlayerTwo
)

type Lane struct {
	Cards   [2]Deck `json:"cards"`
	Claimed int     `json:"claimed"`
}

func (lane Lane) IsComplete() bool {
	return len(lane.Cards[0]) >= MaxCardsPerSide &&
		len(lane.Cards[1]) >= MaxCardsPerSide
}

type GameLanes [9]Lane

// TODO figure out which lanes the opposing player can
// no longer win and make them claimable
func (lanes *GameLanes) GetClaimableLanes(playerIdx int) []int {
	opponentIdx := 1
	if playerIdx != 0 {
		opponentIdx = 0
		playerIdx = 1
	}
	claimableLanes := []int{}
	for i, lane := range lanes {
		if lane.IsComplete() &&
			lane.Cards[playerIdx].GetFormation() > lane.Cards[opponentIdx].GetFormation() {
			claimableLanes = append(claimableLanes, i)
		}
	}
	return claimableLanes
}

func (lanes *GameLanes) PlayerCanPlaceInLane(playerIdx int, laneIdx int) bool {
	if laneIdx < 0 && laneIdx > len(lanes)-1 {
		return false
	}
	cardsCount := len(lanes[laneIdx].Cards[playerIdx])
	return cardsCount < MaxCardsPerSide
}
