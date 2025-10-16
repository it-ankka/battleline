package gameserver

import (
	"context"
	"log/slog"
	"time"

	"github.com/coder/websocket/wsjson"
	gamestate "github.com/it-ankka/battleline/internal/gamestate"
)

type SessionMessageType string
type ClientMessageType string

const (
	ClientMessageReady ClientMessageType = "ready"
	ClientMessageMove  ClientMessageType = "move"
	ClientMessageChat  ClientMessageType = "chat"
	ClientMessageClose ClientMessageType = "close"

	SessionMessagePing  SessionMessageType = "ping"
	SessionMessageSync  SessionMessageType = "sync"
	SessionMessageError SessionMessageType = "error"
	SessionMessageClose SessionMessageType = "close"

	SessionMessageClientReady      SessionMessageType = "client_ready"
	SessionMessageClientMove       SessionMessageType = "client_move"
	SessionMessageClientChat       SessionMessageType = "chat"
	SessionMessageClientConnect    SessionMessageType = "client_connect"
	SessionMessageClientDisconnect SessionMessageType = "client_connect"
)

type SessionMessage struct {
	MessageType SessionMessageType          `json:"type"`
	Timestamp   time.Time                   `json:"timestamp"`
	GameState   *gamestate.PrivateGameState `json:"state"`
	SessionInfo *GameSessionSnapshot        `json:"session"`
	Error       any                         `json:"error"`
}

type ClientMessageData struct {
	Move  *any    `json:"move"` //TODO
	Chat  *string `json:"chat"`
	Ready *bool   `json:"ready"`
}

type ClientMessage struct {
	Client      *SessionClient
	MessageType ClientMessageType  `json:"type"`
	Data        *ClientMessageData `json:"data"`
}

// Checks client message vali
func (m ClientMessage) IsValid() bool {
	switch m.MessageType {
	case ClientMessageReady:
		return m.Data != nil && m.Data.Ready != nil
	case ClientMessageChat:
		return m.Data != nil && m.Data.Chat != nil && len(*m.Data.Chat) > 0
	default:
		return false
	}
}

func (game *GameSession) Broadcast(messageType SessionMessageType) {
	for _, client := range game.Clients {
		if client == nil {
			continue
		}
		if client.Connection == nil {
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
func (game *GameSession) HandleClientReadyMessage(m ClientMessage) {
	game.mu.Lock()
	defer game.mu.Unlock()
	m.Client.Ready = *m.Data.Ready
	game.Broadcast(SessionMessageClientReady)
}

// TODO
func (game *GameSession) HandleClientMoveMessage(m ClientMessage) {
}

func (game *GameSession) HandleClientChatMessage(m ClientMessage) {
	game.AddChatMessage(&ChatMessage{
		Timestamp: time.Now(),
		ClientId:  m.Client.ID,
		Nickname:  m.Client.Nickname,
		Content:   *m.Data.Chat,
	})
	game.Broadcast(SessionMessageClientChat)
}

// TODO
func (game *GameSession) HandleClientCloseMessage(m ClientMessage) {
}

func (game *GameSession) ProcessClientMessage(m ClientMessage) {

	slog.Info("ClientMessage received", slog.Any("clientMessage", m))

	if !m.IsValid() {
		slog.Error("Unable to process client message.", slog.String("clientId", m.Client.ID))
		return
	}

	switch m.MessageType {
	case ClientMessageReady:
		game.HandleClientReadyMessage(m)
	case ClientMessageMove:
		game.HandleClientMoveMessage(m)
	case ClientMessageChat:
		game.HandleClientChatMessage(m)
	case ClientMessageClose:
		game.HandleClientCloseMessage(m)
	default:
		slog.Error("Unable to process client message.", slog.String("clientId", m.Client.ID))
		return
	}
}
