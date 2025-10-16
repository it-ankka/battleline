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

	// Get initial sync
	game.Broadcast(SessionMessageSync)

	for {
		var m ClientMessage
		err := wsjson.Read(ctx, client.Connection, &m)
		if err != nil {
			wsjson.Write(ctx, client.Connection, SessionMessage{
				Error: struct{ message string }{message: "Invalid message format"},
			})
			continue
		}

		// Include client info
		m.Client = client

		select {
		case game.messages <- m:
		default:
			wsjson.Write(ctx, client.Connection, SessionMessage{
				Error: struct{ message string }{message: "Session not responding"},
			})
		}
	}
}
