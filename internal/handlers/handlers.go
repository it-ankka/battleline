package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/coder/websocket"
	. "github.com/it-ankka/battleline/internal/gameserver"
)

const (
	ClientIdCookieName  = "battlelineClientId"
	ClientKeyCookieName = "battlelineClientKey"
)

func addClientCookies(w http.ResponseWriter, clientId string, clientKey string) {
	http.SetCookie(w, &http.Cookie{
		Name:     ClientIdCookieName,
		Value:    clientId,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(5 * time.Hour),
		HttpOnly: true,
		Secure:   false,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     ClientKeyCookieName,
		Value:    clientKey,
		Path:     "/",
		Domain:   "localhost",
		Expires:  time.Now().Add(5 * time.Hour),
		HttpOnly: true,
		Secure:   false,
	})
}

func getClientCookies(r *http.Request) (clientId string, clientKey string) {
	clientIdCookie, err := r.Cookie(ClientIdCookieName)
	if err == nil {
		clientId = clientIdCookie.Value
	}
	clientKeyCookie, err := r.Cookie(ClientKeyCookieName)
	if err == nil {
		clientKey = clientKeyCookie.Value
	}
	return clientId, clientKey
}

func JoinGameHandler(s *GameServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameId := r.PathValue("gameId")
		game, exists := s.GameManager.GetGame(gameId)
		if !exists {
			http.Error(w, "Unable to join game with ID: "+gameId, 400)
			return
		}
		clientId, clientKey := getClientCookies(r)
		client, _ := game.GetClient(clientId, clientKey)
		if client != nil {
			slog.Info("Client rejoined game", slog.String("clientId", clientId))
			w.WriteHeader(http.StatusNoContent)
			return
		}

		clientInfo, err := game.AddClient()

		if err != nil {
			slog.Error("Failed to add client to game", slog.Any("error", err.Error()))
			http.Error(w, "Unable to join game with ID: "+gameId, 400)
			return
		}
		slog.Info("Client joined game", slog.String("GameId", game.ID), slog.String("clientId", clientInfo.ID))
		addClientCookies(w, clientInfo.ID, clientInfo.Key)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(game)
	}
}

func CreateGameHandler(a *GameServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		game, err := a.GameManager.CreateGame()
		if err != nil {
			slog.Error("Game creation failed", slog.Any("error", err.Error()))
			http.Error(w, "Failed to create game", 500)
		}
		slog.Info("Game Created", slog.String("gameId", game.ID))
		addClientCookies(w, game.Clients[0].ID, game.Clients[0].Key)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(game.GetInfo())
	}
}

func ConnectHandler(s *GameServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		clientId, clientKey := getClientCookies(r)
		gameId := r.PathValue("gameId")
		// Upgrade to websockets
		c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			CompressionMode: websocket.CompressionContextTakeover,
		})
		if err != nil {
			slog.Error("Websocket Error", slog.Any("error", err.Error()), slog.String("clientId", clientId), slog.String("gameId", gameId))
			return
		}

		game, exists := s.GameManager.GetGame(gameId)
		if !exists {
			slog.Error("Game not found", slog.String("gameId", gameId))
			c.Close(websocket.StatusPolicyViolation, "Unable to join session with ID: "+gameId)
			return
		}

		client, err := game.GetClient(clientId, clientKey)
		if err != nil {
			slog.Error("Client connection not authorized", slog.String("clientId", clientId))
			c.Close(websocket.StatusPolicyViolation, "Unable to join session with ID: "+gameId)
			return
		}

		defer c.Close(websocket.StatusNormalClosure, "connection closed")

		if !game.IsStarted() {
			go game.Listen()
		}

		client.HandleConnection(c, game)

	}
}
