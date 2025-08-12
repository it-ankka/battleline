package game

type Lane struct {
	cards   Deck
	claimed int
}

type GameLanes [9][2]Lane
