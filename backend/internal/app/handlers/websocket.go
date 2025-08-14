package handlers

import (
	"context"
	"net/http"
	"time"

	. "github.com/it-ankka/battleline/internal/app/context"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// TODO Check user id and key and process moves and send board status updates
func WebsocketHandler(a *AppContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := websocket.Accept(w, r, nil)
		if err != nil {
			a.Logger.Printf("WebSocket Error %s", err.Error())
			http.Error(w, "Unable to create WebSocket connection.", 500)
		}
		defer c.CloseNow()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()
		var v any
		err = wsjson.Read(ctx, c, &v)
		if err != nil {

		}

		a.Logger.Printf("received: %v", v)

		c.Close(websocket.StatusNormalClosure, "")
	}
}
