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
	case DrawAction:
		return true
	case ClaimAction:
		return true
	default:
		return false
	}
}
