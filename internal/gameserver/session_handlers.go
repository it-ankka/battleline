package gameserver

import (
	"log/slog"
	"time"
)

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
