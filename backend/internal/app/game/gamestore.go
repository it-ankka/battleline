package game

import (
	"errors"
	"sync"
)

type GameStore struct {
	mu    sync.RWMutex
	games map[string]*GameSession
}

func NewGameStore() *GameStore {
	return &GameStore{
		games: make(map[string]*GameSession),
	}
}

func (gs *GameStore) CreateGame() (*GameSession, error) {
	gs.mu.Lock()
	defer gs.mu.Unlock()

	game, err := NewGameSession()
	if err != nil {
		return nil, errors.New("Failed to create game: " + err.Error())
	}

	gs.games[game.ID] = game
	return game, nil
}

func (gs *GameStore) GetGame(gameID string) (*GameSession, bool) {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	game, exists := gs.games[gameID]

	return game, exists
}
