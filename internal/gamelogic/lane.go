package gamelogic

const MaxCardsPerSide = 3
const (
	NotClaimed = iota
	ClaimedByPlayerOne
	ClaimedByPlayerTwo
)

type Lane struct {
	Cards     [2]Deck `json:"cards"`
	Claimed   int     `json:"claimed"`
	Claimable bool    `json:"claimable"`
}

type GameLanes [9]Lane

func (gameState *GameState) UpdateClaimableLanes(playerIdx int) {
	opponentIdx := 1 - playerIdx

	unplayedCards := gameState.TroopDeck.Copy()
	unplayedCards = append(unplayedCards, gameState.PlayerHands[0]...)
	unplayedCards = append(unplayedCards, gameState.PlayerHands[1]...)

	for i := range gameState.Lanes {
		lane := &gameState.Lanes[i]

		if lane.Claimed != NotClaimed {
			lane.Claimable = false
			continue
		}

		playerCards := lane.Cards[playerIdx]
		opponentCards := lane.Cards[opponentIdx]

		playerSideComplete := len(playerCards) >= MaxCardsPerSide
		if !playerSideComplete {
			lane.Claimable = false
			continue
		}

		opponentSideComplete := len(opponentCards) >= MaxCardsPerSide
		if opponentSideComplete {
			lane.Claimable = playerCards.GetTotalValue() > opponentCards.GetTotalValue()
			continue
		}

		lane.Claimable = opponentCards.GetBestPossibleTotalValue(0, unplayedCards) < playerCards.GetTotalValue()
	}
}

func (gameState *GameState) PlayerCanClaimLane(playerIdx int, laneIdx int) bool {
	if (laneIdx < 0 && laneIdx > len(gameState.Lanes)-1) || gameState.Lanes[laneIdx].Claimed != NotClaimed {
		return false
	}
	return gameState.Lanes[laneIdx].Claimable
}

func (lanes *GameLanes) PlayerCanPlaceInLane(playerIdx int, laneIdx int) bool {
	if (laneIdx < 0 && laneIdx > len(lanes)-1) || lanes[laneIdx].Claimed != NotClaimed {
		return false
	}
	cardsCount := len(lanes[laneIdx].Cards[playerIdx])
	return cardsCount < MaxCardsPerSide
}
