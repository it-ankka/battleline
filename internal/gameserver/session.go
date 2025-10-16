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

type SessionStatus int
type SessionMessageType int

const (
	SessionCreated SessionStatus = iota
	SessionReady
	SessionStarted
	SessionEnded
)

const (
	SessionMessageTick SessionMessageType = iota
	SessionMessageMove
	SessionMessageChat
	SessionMessageClose
)

type ChatMessage struct {
	SentAt   time.Time `json:"timestamp"`
	ClientId string    `json:"clientId"`
	Nickname string    `json:"nickname"`
	Content  string    `json:"content"`
}

type SessionMessage struct {
	MessageType string           `json:"type"`
	GameState   PrivateGameState `json:"state"`
	SessionInfo GameSessionInfo  `json:"session"`
	Error       any              `json:"error"`
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

type GameSessionInfo struct {
	ID        string            `json:"id"`
	Clients   [2]*SessionClient `json:"clients"`
	Status    SessionStatus     `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
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

func (game *GameSession) GetInfo() GameSessionInfo {
	return GameSessionInfo{
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

func (messageType SessionMessageType) ToString() string {
	switch messageType {
	case SessionMessageTick:
		return "tick"
	case SessionMessageMove:
		return "move"
	case SessionMessageChat:
		return "chat"
	case SessionMessageClose:
		return "close"
	default:
		return ""
	}
}

func (game *GameSession) Broadcast(messageType SessionMessageType) {
	for i, client := range game.Clients {
		if client != nil && client.Connection != nil {
			privateGameState := game.GameState.GetPrivateGameState(i)
			message := SessionMessage{
				MessageType: messageType.ToString(),
				GameState:   privateGameState,
				SessionInfo: game.GetInfo(),
			}

			wsjson.Write(context.Background(), client.Connection, message)
		} else if client != nil && client.Connection != nil {
			slog.Error("Could not connect to client", slog.String("clientId", client.ID))
		}
	}
}

// TODO
func (game *GameSession) HandleMove(client *SessionClient, data *ClientMessageData) {
}

func (game *GameSession) HandleChatMessage(client *SessionClient, data *ClientMessageData) {
	game.mu.Lock()
	defer game.mu.Unlock()
	chatMessage := &ChatMessage{
		SentAt:   time.Now(),
		ClientId: client.ID,
		Content:  *data.Chat,
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

	switch m.GetType() {
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
