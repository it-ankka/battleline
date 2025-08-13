package game

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"
)

type SessionStatus int

const (
	SessionCreated SessionStatus = iota
	SessionInProgress
	SessionCompleted
)

type PlayerInfo struct {
	ID       string
	Key      string
	Nickname string `json:"nickname"`
	Ready    bool   `json:"ready"`
}

type GameSession struct {
	ID        string         `json:"id"`
	Players   [2]*PlayerInfo `json:"players"`
	Status    SessionStatus  `json:"status"`
	CreatedAt time.Time      `json:"createdAt"`
}

func generateID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(bytes)
	return strings.TrimRight(id[:length], "="), nil
}

func generatePlayerKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func NewPlayerInfo() (*PlayerInfo, error) {
	playerId, err := generateID(16)
	if err != nil {
		return nil, errors.New("Unable to generate IDs.")
	}

	playerKey, err := generatePlayerKey()
	if err != nil {
		return nil, errors.New("Unable to generate Player Key.")
	}

	p := &PlayerInfo{
		ID:       playerId,
		Key:      playerKey,
		Nickname: "Player",
		Ready:    false,
	}

	return p, nil
}

func NewGameSession() (*GameSession, error) {
	id, err := generateID(16)
	if err != nil {
		return nil, errors.New("Unable to generate game IDs.")
	}

	s := &GameSession{
		ID:        id,
		Status:    SessionCreated,
		CreatedAt: time.Now().UTC(),
	}

	p, err := NewPlayerInfo()

	if err != nil {
		return nil, errors.New("Unable to initialize players.")
	}

	s.Players[0] = p

	return s, nil
}

func (gs *GameSession) ConnectPlayer() (*PlayerInfo, error) {
	// If a second player has not joined
	if gs.Players[1] == nil {
		p, err := NewPlayerInfo()
		if err != nil {
			return nil, errors.New("Unable to connect player.")
		}
		gs.Players[1] = p
		return p, nil
	}
	return nil, errors.New("Game is full.")
}
