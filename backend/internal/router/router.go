package router

import (
	"net/http"

	. "github.com/it-ankka/battleline/internal/gameserver"
	. "github.com/it-ankka/battleline/internal/handlers"
)

func NewAppRouter(s *GameServer) *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/", http.FileServer(http.Dir("./web/static")))
	router.HandleFunc("/ws/{gameId}", ConnectHandler(s))
	router.HandleFunc("POST /game", CreateGameHandler(s))
	router.HandleFunc("POST /game/{gameId}", JoinGameHandler(s))
	return router
}
