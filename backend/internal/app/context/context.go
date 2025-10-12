package context

import (
	. "github.com/it-ankka/battleline/internal/app/game"
)

type AppContext struct {
	// Fs       fs.FS
	GameManager *GameManager
}

func NewAppContext() *AppContext {
	return &AppContext{
		// Fs:       filesystem,
		GameManager: NewGameManager(),
	}
}
