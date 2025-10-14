package game

const (
	NotClaimed = iota
	ClaimedByPlayerOne
	ClaimedByPlayerTwo
)

type Lane struct {
	Cards   [2]Deck `json:"cards"`
	Claimed int     `json:"claimed"`
}

type GameLanes [9]Lane

// TODO
func (lanes *GameLanes) GetClaimableLanes(playerIdx int) {
	if playerIdx != 0 {
		playerIdx = 1
	}

}
