package handler

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pradiptaagus/go-chat-app/internal/service"
	"github.com/pradiptaagus/go-chat-app/pkg/util"
)

var upgrader = websocket.Upgrader{
	// Set size of internal read buffer I/O
	ReadBufferSize: 1024,

	// Set size of internal write buffer I/O
	WriteBufferSize: 1024,

	// Set allowed origins for the websocket connection.
	// For now, allow all connections from any origins
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handles websokcet requests from the peer
func WsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, hub service.Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	util.PanicIfError(err)

	// Create new client
	client := service.NewClient(conn, hub)

	// Register client into Hub service
	client.Register(ctx)

	// Allow client to read and write message.
	go client.ReadMessage(ctx)
	go client.WriteMessage(ctx)
}
