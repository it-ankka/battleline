package gameserver

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/coder/websocket/wsjson"
	. "github.com/it-ankka/battleline/internal/gamestate"
	"github.com/it-ankka/battleline/internal/gameutils"
)

type SessionStatus string
type SessionMessageType string

const (
	SessionCreated SessionStatus = "created"
	SessionReady   SessionStatus = "ready"
	SessionStarted SessionStatus = "started"
	SessionEnded   SessionStatus = "ended"
)

const (
	SessionMessageTick  SessionMessageType = "tick"
	SessionMessageMove  SessionMessageType = "move"
	SessionMessageChat  SessionMessageType = "chat"
	SessionMessageError SessionMessageType = "close"
	SessionMessageClose SessionMessageType = "close"
)

type ChatMessage struct {
	Timestamp time.Time `json:"timestamp"`
	ClientId  string    `json:"clientId"`
	Nickname  string    `json:"nickname"`
	Content   string    `json:"content"`
}

type SessionMessage struct {
	MessageType SessionMessageType   `json:"type"`
	Timestamp   time.Time            `json:"timestamp"`
	GameState   *PrivateGameState    `json:"state"`
	SessionInfo *GameSessionSnapshot `json:"session"`
	Error       any                  `json:"error"`
}

type GameSession struct {
	ID        string            `json:"id"`
	Clients   [2]*SessionClient `json:"clients"`
	Status    SessionStatus     `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
	ChatLog   []*ChatMessage    `json:"chatLog"`

	GameState *GameState

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
		Status:    SessionCreated,
		CreatedAt: time.Now().UTC(),
		GameState: NewGameState(),
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
	if game.Clients[1] == nil && game.Status != SessionEnded {
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

func (game *GameSession) Broadcast(messageType SessionMessageType) {
	for _, client := range game.Clients {
		if client == nil || client.Connection == nil {
			slog.Error("Could not connect to client", slog.String("clientId", client.ID))
			continue
		}
		wsjson.Write(context.Background(), client.Connection, SessionMessage{
			MessageType: messageType,
			GameState:   game.GameState.GetPrivateGameState(client.Index),
			SessionInfo: game.Snapshot(),
		})
	}
}

// TODO
func (game *GameSession) HandleMove(client *SessionClient, data *ClientMessageData) {
}

func (game *GameSession) HandleChatMessage(client *SessionClient, data *ClientMessageData) {
	game.mu.Lock()
	defer game.mu.Unlock()
	chatMessage := &ChatMessage{
		Timestamp: time.Now(),
		ClientId:  client.ID,
		Nickname:  client.Nickname,
		Content:   *data.Chat,
	}
	game.ChatLog = append(game.ChatLog, chatMessage)
	game.Broadcast(SessionMessageChat)
}

// TODO
func (game *GameSession) HandleUpdateClientInfo(client *SessionClient, data *ClientMessageData) {
}

// TODO
func (game *GameSession) HandleClientClose(client *SessionClient) {
}

func (game *GameSession) ProcessClientMessage(m ClientMessage) {

	slog.Info("ClientMessage received", slog.Any("clientMessage", m))

	client, err := game.GetClient(m.ClientId, m.ClientKey)
	if err != nil {
		slog.Error("Unable to process client message.", slog.String("clientId", m.ClientId), slog.Any("error", err.Error()))
		return
	}

	if !m.IsValid() {
		slog.Error("Unable to process client message.", slog.String("clientId", m.ClientId))
		return
	}

	switch m.MessageType {
	case ClientMessageMove:
		game.HandleMove(client, m.Data)
	case ClientMessageChat:
		game.HandleChatMessage(client, m.Data)
	case ClientMessageUpdateInfo:
		game.HandleUpdateClientInfo(client, m.Data)
	case ClientMessageClose:
		game.HandleClientClose(client)
	default:
		slog.Error("Unable to process client message.", slog.String("clientId", m.ClientId))
		return
	}
}

func (game *GameSession) IsStarted() bool {
	return game.Status != SessionCreated
}

func (game *GameSession) StartUpdateTick(d time.Duration) {
	for {
		time.Sleep(d)
		game.Broadcast(SessionMessageTick)
	}
}

func (game *GameSession) Listen() {
	slog.Info("Game listening", slog.String("gameId", game.ID))
	func() {
		game.mu.Lock()
		defer game.mu.Unlock()
		game.Status = SessionReady
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
