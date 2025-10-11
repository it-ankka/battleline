package router

import (
	"net/http"

	. "github.com/it-ankka/battleline/internal/app/context"
	. "github.com/it-ankka/battleline/internal/app/handlers"
)

func NewAppRouter(a *AppContext) *http.ServeMux {
	router := http.NewServeMux()
	router.Handle("/", http.FileServer(http.Dir("./web/static")))
	router.HandleFunc("/ws/{gameId}", ConnectHandler(a))
	router.HandleFunc("POST /game", CreateGameHandler(a))
	router.HandleFunc("POST /game/{gameId}", JoinGameHandler(a))
	return router
}
