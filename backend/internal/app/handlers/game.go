package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	. "github.com/it-ankka/battleline/internal/app/context"
)

const (
	PlayerIdCookieName  = "battlelinePlayerId"
	PlayerKeyCookieName = "battlelinePlayerKey"
)

func addPlayerCookies(w http.ResponseWriter, playerId string, playerKey string) {
	http.SetCookie(w, &http.Cookie{
		Name:     PlayerIdCookieName,
		Value:    playerId,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(5 * time.Hour),
		HttpOnly: true,
		Secure:   false,
	})

	http.SetCookie(w, &http.Cookie{
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
		game, exists := a.GameManager.GetGame(gameId)
		if !exists {
			http.Error(w, "Game not found with ID "+gameId, 404)
			return
		}
		playerInfo, err := game.AddPlayer()

		if err != nil {
			http.Error(w, "Unable to connect to game", 500)
			return
		}
		slog.Info("Player joined game", slog.String("GameId", game.ID), slog.String("playerId", playerInfo.ID))
		addPlayerCookies(w, playerInfo.ID, playerInfo.Key)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(game)
	}
}

// TODO Return some actually useful data
func CreateGameHandler(a *AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		game, err := a.GameManager.CreateGame()
		if err != nil {
			slog.Error("Game creation failed", slog.Any("error", err.Error()))
			http.Error(w, "Failed to create game", 500)
		}
		slog.Info("Game Created", slog.String("gameId", game.ID))
		addPlayerCookies(w, game.Players[0].ID, game.Players[0].Key)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(game)
	}
}

// TODO Check user id and key and process moves and send board status updates
func ConnectHandler(a *AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameId := r.PathValue("gameId")
		game, exists := a.GameManager.GetGame(gameId)
		if !exists {
			http.Error(w, "Game not found with ID "+gameId, 404)
			return
		}

		playerId, playerKey := getPlayerCookies(r)
		player, err := game.GetPlayer(playerId, playerKey)
		if err != nil {
			slog.Error("Player connection not authorized", slog.String("playerId", playerId))
			http.Error(w, "Player not authorized to connect to game", http.StatusUnauthorized)
			return
		}

		// Upgrade to websockets
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			slog.Error("Websocket Error", slog.Any("error", err.Error()), slog.String("playerId", playerId), slog.String("gameId", gameId))
			http.Error(w, "Unable to create WebSocket connection.", 500)
			return
		}
		defer c.CloseNow()

		if !game.IsStarted() {
			go game.StartUpdateTick(time.Second * 1)
			go game.Listen()
		}

		player.HandleConnection(c, game)

	}
}
