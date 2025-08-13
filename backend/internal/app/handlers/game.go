package handlers

import (
	"net/http"
	"time"

	. "github.com/it-ankka/battleline/internal/app/context"
)

func addPlayerCookies(r *http.Request, playerId string, playerKey string) {
	r.AddCookie(&http.Cookie{
		Name:     "battlelinePlayerId",
		Value:    playerId,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(5 * time.Hour),
		HttpOnly: true,
		Secure:   false,
	})

	r.AddCookie(&http.Cookie{
		Name:     "battlelinePlayerKey",
		Value:    playerKey,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(5 * time.Hour),
		HttpOnly: true,
		Secure:   false,
	})
}

func JoinGameHandler(a *AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameId := r.PathValue("gameId")
		if len(gameId) < 1 {
			http.Error(w, "Game not found", 404)
			return
		}

		game, exists := a.Store.GetGame(gameId)
		if !exists {
			http.Error(w, "Game not found with ID "+gameId, 404)
			return
		}
		playerInfo, err := game.ConnectPlayer()

		a.Logger.Printf("PLAYER %s JOINED GAME %s", playerInfo.ID, game.ID)
		if err != nil {
			http.Error(w, "Unable to connect to game", 500)
			return
		}
		addPlayerCookies(r, playerInfo.ID, playerInfo.Key)

	}
}

func CreateGameHandler(a *AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		game, err := a.Store.CreateGame()
		if err != nil {
			a.Logger.Print(err.Error())
			http.Error(w, "Failed to create game", 500)
		}
		a.Logger.Printf("NEW GAME CREATED WITH ID %s", game.ID)
		addPlayerCookies(r, game.Players[0].ID, game.Players[0].Key)

	}
}
