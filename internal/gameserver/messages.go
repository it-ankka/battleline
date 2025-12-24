package gameserver

import (
	"context"
	"log/slog"
	"time"

	"github.com/coder/websocket/wsjson"
	"github.com/it-ankka/battleline/internal/gamelogic"
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

type SessionError struct {
	Message string `json:"message"`
}

type SessionMessage struct {
	MessageType SessionMessageType          `json:"type"`
	Timestamp   time.Time                   `json:"timestamp"`
	ClientIdx   int                         `json:"clientIdx"`
	GameState   *gamelogic.PrivateGameState `json:"state"`
	SessionInfo *GameSessionSnapshot        `json:"session"`
	Error       *SessionError               `json:"error"`
}

type ClientMessageData struct {
	Move  *gamelogic.MoveData `json:"move"`
	Chat  *string             `json:"chat"`
	Ready *bool               `json:"ready"`
}

type ClientMessage struct {
	Client      *SessionClient
	MessageType ClientMessageType  `json:"type"`
	Data        *ClientMessageData `json:"data"`
}

func (game *GameSession) IsValidMessage(m ClientMessage) bool {
	switch m.MessageType {
	case ClientMessageSetReady:
		return m.Data != nil && m.Data.Ready != nil && game.Status != SessionStatusInProgress && game.Status != SessionStatusEnded
	case ClientMessageChat:
		return m.Data != nil && m.Data.Chat != nil && len(*m.Data.Chat) > 0
	case ClientMessageMove:
		return m.Data != nil && m.Data.Move != nil &&
			game.Status == SessionStatusInProgress &&
			game.GameState.ActivePlayer == m.Client.Index &&
			game.GameState.IsValidPlayerMove(m.Client.Index, m.Data.Move)
	default:
		return false
	}
}

func (client *SessionClient) SendSessionMessage(
	messageType SessionMessageType,
	game *GameSession,
	error *SessionError,
) {
	if client.Connection == nil {
		slog.Error("Could not connect to client", slog.String("clientId", client.ID))
		return
	}

	message := SessionMessage{MessageType: messageType, Timestamp: time.Now(), ClientIdx: client.Index}
	if game != nil {
		message.SessionInfo = game.Snapshot()
	}

	if game != nil && game.GameState != nil {
		message.GameState = game.GameState.GetPrivateGameState(client.Index)
	}

	if error != nil {
		message.Error = error
	}

	wsjson.Write(context.Background(), client.Connection, message)
}

func (game *GameSession) Broadcast(messageType SessionMessageType) {
	for _, client := range game.Clients {
		if client != nil {
			client.SendSessionMessage(messageType, game, nil)
		}
	}
}

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

func (game *GameSession) HandleClientMoveMessage(m ClientMessage) {
	game.mu.Lock()
	defer game.mu.Unlock()

	game.GameState.ExecutePlayerMove(m.Client.Index, m.Data.Move)
	game.Broadcast(SessionMessageClientMove)
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

	if !game.IsValidMessage(m) {
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
