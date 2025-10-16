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
	ClientMessageSetReady ClientMessageType = "set_ready"
	ClientMessageMove     ClientMessageType = "move"
	ClientMessageChat     ClientMessageType = "chat"
	ClientMessageClose    ClientMessageType = "close"

	SessionMessagePing  SessionMessageType = "ping"
	SessionMessageSync  SessionMessageType = "sync"
	SessionMessageError SessionMessageType = "error"
	SessionMessageClose SessionMessageType = "close"

	SessionMessageSessionStart SessionMessageType = "session_start"
	SessionMessageSessionEnd   SessionMessageType = "session_end"

	SessionMessageClientReady      SessionMessageType = "client_ready"
	SessionMessageClientUnready    SessionMessageType = "client_unready"
	SessionMessageClientMove       SessionMessageType = "client_move"
	SessionMessageClientChat       SessionMessageType = "client_chat"
	SessionMessageClientConnect    SessionMessageType = "client_connect"
	SessionMessageClientDisconnect SessionMessageType = "client_disconnect"
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
	case ClientMessageSetReady:
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
		var privateGameState *gamestate.PrivateGameState = nil
		if game.GameState != nil {
			privateGameState = game.GameState.GetPrivateGameState(client.Index)
		}
		wsjson.Write(context.Background(), client.Connection, SessionMessage{
			MessageType: messageType,
			GameState:   privateGameState,
			SessionInfo: game.Snapshot(),
		})
	}
}

// TODO
func (game *GameSession) HandleClientSetReadyMessage(m ClientMessage) {

	game.mu.Lock()
	m.Client.Ready = *m.Data.Ready
	game.mu.Unlock()

	if game.IsReadyToStart() {
		game.StartGame()
		return
	}
	if *m.Data.Ready {
		game.Broadcast(SessionMessageClientReady)
	} else {
		game.Broadcast(SessionMessageClientUnready)
	}
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
	case ClientMessageSetReady:
		game.HandleClientSetReadyMessage(m)
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
