package gameserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket/wsjson"
	. "github.com/it-ankka/battleline/internal/gamestate"
)

type SessionStatus int

const (
	SessionCreated SessionStatus = iota
	SessionStarted
	SessionEnded
)

type SessionMessage struct {
	GameState PrivateGameState `json:"state"`
	Error     any              `json:"error"`
}

type GameSession struct {
	ID        string            `json:"id"`
	Clients   [2]*SessionClient `json:"clients"`
	Status    SessionStatus     `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
	GameState GameState

	// Channels for communication
	messages chan ClientMessage
	done     chan struct{}

	mu sync.RWMutex
}

func NewGameSession() (*GameSession, error) {
	id, err := generateID(16)
	if err != nil {
		return nil, errors.New("Unable to generate game IDs.")
	}

	game := &GameSession{
		ID:        id,
		Status:    SessionCreated,
		CreatedAt: time.Now().UTC(),
		GameState: NewGameState(),
		messages:  make(chan ClientMessage),
		done:      make(chan struct{}),
	}

	client, err := NewClient()

	if err != nil {
		return nil, errors.New("Unable to initialize clients.")
	}

	game.Clients[0] = client

	return game, nil
}

func (game *GameSession) AddClient() (*SessionClient, error) {
	game.mu.Lock()
	defer game.mu.Unlock()

	// If a second client has not joined and game is not started
	if game.Clients[1] == nil && game.Status != SessionEnded {
		client, err := NewClient()
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

func (game *GameSession) UpdateClients() {
	for i, client := range game.Clients {
		if client != nil && client.Connection != nil {
			privateGameState := game.GameState.GetPrivateGameState(i)
			wsjson.Write(context.Background(), client.Connection, SessionMessage{
				GameState: *privateGameState,
			})
		} else if client != nil && client.Connection != nil {
			slog.Error("Could not connect to client", slog.String("clientId", client.ID))
			// TODO: Unable to connect to client
		}
	}
}

func (game *GameSession) IsStarted() bool {
	return game.Status != SessionCreated
}

func (game *GameSession) StartUpdateTick(d time.Duration) {
	for {
		time.Sleep(d)
		game.UpdateClients()
	}
}

func (game *GameSession) Listen() {
	fmt.Printf("GAME %s LISTENING", game.ID)
	func() {
		game.mu.Lock()
		defer game.mu.Unlock()
		game.Status = SessionStarted
	}()
	for {
		// Wait for game update
		select {

		case message := <-game.messages:
			if message.ActionType == "update" {
				fmt.Printf("UPDATE MESSAGE RECEIVED FROM CLIENT")
				game.UpdateClients()
			}
		case <-game.done:
			fmt.Printf("GAME %s CLOSED", game.ID)
		}
	}
}
