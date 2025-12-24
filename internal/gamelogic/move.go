package gamelogic

type MoveAction string

const (
	PlacementAction MoveAction = "placement"
	DrawAction      MoveAction = "draw"
	ClaimAction     MoveAction = "claim"
)

type MoveData struct {
	Action      MoveAction `json:"action"`
	Card        *Card      `json:"card"`
	Lane        *int       `json:"lane"`
	TacticsDeck *bool      `json:"tacticsDeck"`
}

func (gameState *GameState) IsValidPlayerMove(playerIdx int, move *MoveData) bool {
	switch move.Action {
	case PlacementAction:
		hasRequiredData := move.Card != nil && move.Lane != nil
		return gameState.TurnPhase == PlacementPhase &&
			hasRequiredData &&
			gameState.PlayerHands[playerIdx].FindCardIdx(*move.Card) != -1 &&
			gameState.Lanes.PlayerCanPlaceInLane(playerIdx, *move.Lane)
	case ClaimAction:
		return gameState.TurnPhase == ClaimPhase && move.Lane != nil && gameState.PlayerCanClaimLane(playerIdx, *move.Lane)
	case DrawAction:
		return gameState.TurnPhase == DrawPhase && move.TacticsDeck != nil
	default:
		return false
	}
}

func (gameState *GameState) ExecutePlayerMove(playerIdx int, move *MoveData) {
	switch move.Action {
	case PlacementAction:
		cardIdx := gameState.PlayerHands[playerIdx].FindCardIdx(*move.Card)

		gameState.Lanes[*move.Lane].Cards[playerIdx] = append(gameState.Lanes[*move.Lane].Cards[playerIdx], *move.Card)
		gameState.PlayerHands[playerIdx] = gameState.PlayerHands[playerIdx].RemoveAt(cardIdx)
		gameState.UpdateClaimableLanes(playerIdx)
		claimableLanes := []int{}
		for i, lane := range gameState.Lanes {
			if lane.Claimable {
				claimableLanes = append(claimableLanes, i)
			}
		}
		if len(claimableLanes) < 1 {
			gameState.TurnPhase += 1
		}

	case ClaimAction:
		if playerIdx == 0 {
			gameState.Lanes[*move.Lane].Claimed = ClaimedByPlayerOne
		} else {
			gameState.Lanes[*move.Lane].Claimed = ClaimedByPlayerTwo
		}

	case DrawAction:
		newTroopDeck, card := gameState.TroopDeck.Pop()
		gameState.TroopDeck = newTroopDeck
		gameState.PlayerHands[playerIdx] = append(gameState.PlayerHands[playerIdx], card)

	default:
		return
	}

	gameState.TurnPhase += 1
	if gameState.TurnPhase > DrawPhase {
		if playerIdx == 0 {
			gameState.ActivePlayer = 1
		} else {
			gameState.ActivePlayer = 0
		}
		gameState.TurnPhase = PlacementPhase
	}
}
