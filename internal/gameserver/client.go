package gameserver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type SessionClient struct {
	ID         string
	Key        string
	Connection *websocket.Conn
	Nickname   string `json:"nickname"`
	send       chan SessionMessage
}

// Possible client actions:
// "move": The client makes a move in the game
// "chat" The client sends a message in the game chat
// "updateinfo": The client changes their nickname
// "close": The client closes the connection
type ClientMessage struct {
	ActionType string `json:"type"`
	Data       any    `json:"data"`
}

func generateID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(bytes)
	return strings.TrimRight(id[:length], "="), nil
}

func generateClientKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func NewClient() (*SessionClient, error) {
	clientId, err := generateID(16)
	if err != nil {
		return nil, errors.New("Unable to generate IDs.")
	}

	clientKey, err := generateClientKey()
	if err != nil {
		return nil, errors.New("Unable to generate client Key.")
	}

	client := &SessionClient{
		ID:       clientId,
		Key:      clientKey,
		Nickname: "Client",
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
	ctx := client.Connection.CloseRead(context.Background())

	go client.ListenToSession(ctx)

	for {
		var m ClientMessage
		err := wsjson.Read(ctx, client.Connection, &m)
		if err != nil {
			wsjson.Write(ctx, client.Connection, SessionMessage{
				Error: struct{ message string }{message: "Invalid message format"},
			})
		} else {
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
