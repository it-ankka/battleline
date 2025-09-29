package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	. "github.com/it-ankka/battleline/internal/app/context"
)

const (
	PlayerIdCookieName  = "battlelinePlayerId"
	PlayerKeyCookieName = "battlelinePlayerKey"
)

func addPlayerCookies(r *http.Request, playerId string, playerKey string) {
	r.AddCookie(&http.Cookie{
		Name:     PlayerIdCookieName,
		Value:    playerId,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(5 * time.Hour),
		HttpOnly: true,
		Secure:   false,
	})

	r.AddCookie(&http.Cookie{
		Name:     PlayerKeyCookieName,
		Value:    playerKey,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(5 * time.Hour),
		HttpOnly: true,
		Secure:   false,
	})
}

func getPlayerCookies(r *http.Request) (playerId string, playerKey string) {
	playerIdCookie, err := r.Cookie(PlayerIdCookieName)
	if err == nil {
		playerId = playerIdCookie.Value
	}
	playerKeyCookie, err := r.Cookie(PlayerKeyCookieName)
	if err == nil {
		playerKey = playerKeyCookie.Value
	}
	return playerId, playerKey
}

// TODO Return some actually useful data
func JoinGameHandler(a *AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameId := r.PathValue("gameId")
		game, exists := a.Store.GetGame(gameId)
		if !exists {
			http.Error(w, "Game not found with ID "+gameId, 404)
			return
		}
		playerInfo, err := game.AddPlayer()

		if err != nil {
			http.Error(w, "Unable to connect to game", 500)
			return
		}
		a.Logger.Printf("PLAYER %s JOINED GAME %s", playerInfo.ID, game.ID)
		addPlayerCookies(r, playerInfo.ID, playerInfo.Key)
	}
}

// TODO Return some actually useful data
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

// PROCESS:
// 1.Player sends request to open connection to game
// 2. Check that player is authorized to connect to game
// 3. Upgrade to websockets
// 4. Add connection to gameSession
// 5. Send game state updates in a seperate thread (goroutine)
// 6. Wait for client updates

// TODO Check user id and key and process moves and send board status updates
func GameWebsocketHandler(a *AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameId := r.PathValue("gameId")
		game, exists := a.Store.GetGame(gameId)
		if !exists {
			http.Error(w, "Game not found with ID "+gameId, 404)
			return
		}

		playerId, playerKey := getPlayerCookies(r)
		err := game.CheckPlayerGameAuthorization(playerId, playerKey)
		if err != nil {
			http.Error(w, "Player not authorized to connect to game", http.StatusUnauthorized)
		}

		// Upgrade to websockets
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			a.Logger.Printf("WebSocket Error %s", err.Error())
			http.Error(w, "Unable to create WebSocket connection.", 500)
		}
		defer c.CloseNow()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		var v any
		// Keep reading data until connection closed
		for true {
			err = wsjson.Read(ctx, c, &v)
			if err != nil {
				a.Logger.Printf("Error reading data: %s", err.Error())
				c.Close(websocket.StatusInternalError, "Error reading data")
			}
			a.Logger.Printf("received: %v", v)
			break
		}
	}
}
