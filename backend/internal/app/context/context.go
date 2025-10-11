package context

import (
	"log"

	. "github.com/it-ankka/battleline/internal/app/game"
)

type AppContext struct {
	// Fs       fs.FS
	Logger      *log.Logger
	GameManager *GameManager
}

func NewAppContext(logger *log.Logger) *AppContext {
	return &AppContext{
		// Fs:       filesystem,
		Logger:      logger,
		GameManager: NewGameManager(),
	}
}

func Run() {

}
