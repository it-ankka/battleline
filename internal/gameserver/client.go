package gameserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/it-ankka/battleline/internal/gameutils"
)

type SessionClient struct {
	Key        string
	Connection *websocket.Conn
	send       chan SessionMessage

	ID        string `json:"playerId"`
	Index     int    `json:"playerIndex"`
	Nickname  string `json:"nickname"`
	Connected bool   `json:"connected"`
	Ready     bool   `json:"ready"`
}

type ClientMessageType string

const (
	ClientMessageMove       ClientMessageType = "move"
	ClientMessageChat       ClientMessageType = "chat"
	ClientMessageUpdateInfo ClientMessageType = "updateInfo"
	ClientMessageClose      ClientMessageType = "close"
)

type ClientMessageData struct {
	Move *any    `json:"move"` //TODO
	Chat *string `json:"chat"`
}

type ClientMessage struct {
	ClientId    string
	ClientKey   string
	MessageType ClientMessageType  `json:"type"`
	Data        *ClientMessageData `json:"data"`
}

// Checks client message vali
func (m ClientMessage) IsValid() bool {
	switch m.MessageType {
	case ClientMessageChat:
		return m.Data != nil && m.Data.Chat != nil && len(*m.Data.Chat) > 0
	default:
		return false
	}
}

func NewClient(index int) (*SessionClient, error) {
	clientId, err := gameutils.GenerateID(16)
	if err != nil {
		return nil, errors.New("Unable to generate IDs.")
	}

	clientKey, err := gameutils.GenerateKey(32)
	if err != nil {
		return nil, errors.New("Unable to generate client Key.")
	}

	client := &SessionClient{
		ID:       clientId,
		Index:    index,
		Key:      clientKey,
		Nickname: fmt.Sprintf("Player %d", index+1),
		send:     make(chan SessionMessage),
	}

	return client, nil
}

func (client *SessionClient) ListenToSession(ctx context.Context) {
	for {
		message := <-client.send
		err := wsjson.Write(ctx, client.Connection, message)
		if err != nil {
			slog.Error("Failed to send message to client", slog.String("client", client.ID), slog.Any("error", err.Error()))
		}
	}
}

func (client *SessionClient) HandleConnection(c *websocket.Conn, game *GameSession) {
	client.Connection = c
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go client.ListenToSession(ctx)

	game.Broadcast(SessionMessageTick)

	for {
		var m ClientMessage
		err := wsjson.Read(ctx, client.Connection, &m)
		if err != nil {
			wsjson.Write(ctx, client.Connection, SessionMessage{
				Error: struct{ message string }{message: "Invalid message format"},
			})
		} else {
			// Add client information to the message
			m.ClientId = client.ID
			m.ClientKey = client.Key

			select {
			case game.messages <- m:
			default:
				wsjson.Write(ctx, client.Connection, SessionMessage{
					Error: struct{ message string }{message: "Session not responding"},
				})
			}
		}
	}
}
