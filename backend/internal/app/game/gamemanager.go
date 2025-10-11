package game

import (
	"errors"
	"sync"
)

type GameManager struct {
	mu    sync.RWMutex
	games map[string]*GameSession
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*GameSession),
	}
}

func (gm *GameManager) CreateGame() (*GameSession, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	game, err := NewGameSession()
	if err != nil {
		return nil, errors.New("Failed to create game: " + err.Error())
	}

	gm.games[game.ID] = game
	return game, nil
}

func (gm *GameManager) GetGame(gameID string) (*GameSession, bool) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	game, exists := gm.games[gameID]

	return game, exists
}
