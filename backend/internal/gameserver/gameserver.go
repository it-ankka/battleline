package gameserver

import (
	. "github.com/it-ankka/battleline/internal/game"
)

type GameServer struct {
	// Fs       fs.FS
	GameManager *GameManager
}

func NewGameServer() *GameServer {
	return &GameServer{
		// Fs:       filesystem,
		GameManager: NewGameManager(),
	}
}
