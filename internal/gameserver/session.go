package gameserver

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/it-ankka/battleline/internal/gamelogic"
	"github.com/it-ankka/battleline/internal/gameutils"
)

type SessionStatus string

const (
	SessionStatusCreated    SessionStatus = "created"
	SessionStatusReady      SessionStatus = "ready"
	SessionStatusInProgress SessionStatus = "in-progress"
	SessionStatusEnded      SessionStatus = "ended"
)

type GameSession struct {
	ID        string            `json:"id"`
	Clients   [2]*SessionClient `json:"clients"`
	Status    SessionStatus     `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
	ChatLog   []*ChatMessage    `json:"chatLog"`

	GameState *gamelogic.GameState

	// Channels for communication
	messages chan ClientMessage
	done     chan struct{}

	mu sync.RWMutex
}

type GameSessionSnapshot struct {
	ID        string            `json:"id"`
	Status    SessionStatus     `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
	Clients   [2]*SessionClient `json:"clients"`
	ChatLog   []*ChatMessage    `json:"chatLog"`
}

func NewGameSession() (*GameSession, error) {
	id, err := gameutils.GenerateID(16)
	if err != nil {
		return nil, errors.New("Unable to generate game IDs.")
	}

	game := &GameSession{
		ID:        id,
		Status:    SessionStatusCreated,
		CreatedAt: time.Now().UTC(),
		GameState: nil,
		ChatLog:   []*ChatMessage{},
		messages:  make(chan ClientMessage),
		done:      make(chan struct{}),
	}

	client, err := NewClient(0)

	if err != nil {
		return nil, errors.New("Unable to initialize clients.")
	}

	game.Clients[0] = client

	return game, nil
}

func (game *GameSession) Snapshot() *GameSessionSnapshot {
	return &GameSessionSnapshot{
		ID:        game.ID,
		Status:    game.Status,
		CreatedAt: game.CreatedAt,
		Clients:   game.Clients,
		ChatLog:   game.ChatLog,
	}
}

func (game *GameSession) AddClient() (*SessionClient, error) {
	game.mu.Lock()
	defer game.mu.Unlock()

	// If a second client has not joined and game is not started
	if game.Clients[1] == nil && game.Status != SessionStatusEnded {
		client, err := NewClient(1)
		if err != nil {
			return nil, errors.New("Unable to add client to session.")
		}
		game.Clients[1] = client
		return client, nil
	}
	return nil, errors.New("Game is full.")
}

func (game *GameSession) GetClient(clientId string, clientKey string) (*SessionClient, error) {
	for _, p := range game.Clients {
		if p != nil && p.ID == clientId && p.Key == clientKey {
			return p, nil
		}
	}
	return nil, errors.New("Client not found.")
}

func (game *GameSession) IsListening() bool {
	return game.Status != SessionStatusCreated && game.Status != SessionStatusEnded
}

func (game *GameSession) IsReadyToStart() bool {
	if game.Status != SessionStatusReady {
		return false
	}
	for _, client := range game.Clients {
		if client == nil || client.Connection == nil || client.Ready == false {
			return false
		}
	}
	return true
}

func (game *GameSession) StartGame() {
	game.mu.Lock()
	defer game.mu.Unlock()
	game.GameState = gamelogic.NewGameState()
	game.Status = SessionStatusInProgress
	game.Broadcast(SessionMessageSessionStart)
}

func (game *GameSession) Listen() {
	slog.Info("Game listening", slog.String("gameId", game.ID))
	func() {
		game.mu.Lock()
		defer game.mu.Unlock()
		game.Status = SessionStatusReady
	}()
	for {
		// Wait for client messages
		select {
		case message := <-game.messages:
			game.ProcessClientMessage(message)
		case <-game.done:
			slog.Info("Game closed", slog.String("gameId", game.ID))
		}
	}
}
