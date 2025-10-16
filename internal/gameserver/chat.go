package gameserver

import "time"

type ChatMessage struct {
	Timestamp time.Time `json:"timestamp"`
	ClientId  string    `json:"clientId"`
	Nickname  string    `json:"nickname"`
	Content   string    `json:"content"`
}

func (game *GameSession) AddChatMessage(chatMessage *ChatMessage) {
	game.mu.Lock()
	defer game.mu.Unlock()
	game.ChatLog = append(game.ChatLog, chatMessage)
}
