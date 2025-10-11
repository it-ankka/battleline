package game

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type SessionStatus int

const (
	SessionCreated SessionStatus = iota
	SessionStarted
	SessionEnded
)

type SessionUpdateType int

type SessionPlayer struct {
	ID         string
	Key        string
	Connection *websocket.Conn
	Nickname   string `json:"nickname"`
	send       chan SessionMessage
}

type PlayerMessage struct {
	MessageType string `json:"type"`
	Data        any    `json:"data"`
}

type SessionMessage struct {
	GameState PrivateGameState `json:"state"`
	Error     any              `json:"error"`
}

type GameSession struct {
	ID        string            `json:"id"`
	Players   [2]*SessionPlayer `json:"players"`
	Status    SessionStatus     `json:"status"`
	CreatedAt time.Time         `json:"createdAt"`
	GameState GameState

	// Channels for communication
	messages chan PlayerMessage
	done     chan struct{}

	mu sync.RWMutex
}

// --- MISC ---

func generateID(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	id := base64.URLEncoding.EncodeToString(bytes)
	return strings.TrimRight(id[:length], "="), nil
}

func generatePlayerKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// --- PLAYER ---

func NewPlayer() (*SessionPlayer, error) {
	playerId, err := generateID(16)
	if err != nil {
		return nil, errors.New("Unable to generate IDs.")
	}

	playerKey, err := generatePlayerKey()
	if err != nil {
		return nil, errors.New("Unable to generate Player Key.")
	}

	p := &SessionPlayer{
		ID:       playerId,
		Key:      playerKey,
		Nickname: "Player",
		send:     make(chan SessionMessage),
	}

	return p, nil
}

func (p *SessionPlayer) HandleConnection(c *websocket.Conn, game *GameSession) {
	p.Connection = c
	ctx := p.Connection.CloseRead(context.Background())

	go func() {
		for {
			// Wait for updates from session
			message := <-p.send
			err := wsjson.Write(ctx, p.Connection, message)
			if err != nil {
				// TODO Print error
			}
		}
	}()

	for {
		var m PlayerMessage
		err := wsjson.Read(ctx, p.Connection, &m)
		if err != nil {
			wsjson.Write(ctx, p.Connection, SessionMessage{
				Error: struct{ message string }{message: "Invalid message format"},
			})
		} else {
			// Send message to gameSession for processing
			select {
			case game.messages <- m:
			default:
				wsjson.Write(ctx, p.Connection, SessionMessage{
					Error: struct{ message string }{message: "Session not responding"},
				})
			}
		}
	}
}

func NewGameSession() (*GameSession, error) {
	id, err := generateID(16)
	if err != nil {
		return nil, errors.New("Unable to generate game IDs.")
	}

	s := &GameSession{
		ID:        id,
		Status:    SessionCreated,
		CreatedAt: time.Now().UTC(),
		GameState: NewGameState(),
		messages:  make(chan PlayerMessage),
		done:      make(chan struct{}),
	}

	p, err := NewPlayer()

	if err != nil {
		return nil, errors.New("Unable to initialize players.")
	}

	s.Players[0] = p

	return s, nil
}

func (game *GameSession) AddPlayer() (*SessionPlayer, error) {
	game.mu.Lock()
	defer game.mu.Unlock()

	// If a second player has not joined and game is not started
	if game.Players[1] == nil && game.Status == SessionCreated {
		p, err := NewPlayer()
		if err != nil {
			return nil, errors.New("Unable to add player to session.")
		}
		game.Players[1] = p
		return p, nil
	}
	return nil, errors.New("Game is full.")
}

func (game *GameSession) GetPlayer(playerId string, playerKey string) (*SessionPlayer, error) {
	for _, p := range game.Players {
		if p.ID == playerId && p.Key == playerKey {
			return p, nil
		}
	}
	return nil, errors.New("Player not authorized to connect to game.")
}

func (game *GameSession) UpdatePlayers() {
	for i, p := range game.Players {
		if p != nil && p.Connection != nil {
			privateGameState := game.GameState.GetPrivateGameState(i)
			wsjson.Write(context.Background(), p.Connection, SessionMessage{
				GameState: *privateGameState,
			})
		} else {
			// TODO: Unable to connect to player
		}
	}
}

func (game *GameSession) IsStarted() bool {
	return game.Status != SessionCreated
}

func (game *GameSession) StartUpdateTick(d time.Duration) {
	for {
		time.Sleep(d)
		game.UpdatePlayers()
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
			if message.MessageType == "update" {
				fmt.Printf("UPDATE MESSAGE RECEIVED FROM PLAYER")
				game.UpdatePlayers()
			}
		case <-game.done:
			fmt.Printf("GAME %s CLOSED", game.ID)
		}
	}
}
