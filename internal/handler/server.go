package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pradiptaagus/go-chat-app/internal/service"
	"github.com/pradiptaagus/go-chat-app/pkg/util"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Allow all connections
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handles websokcet requests from the peer
func WsHandler(w http.ResponseWriter, r *http.Request, hub service.Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	util.PanicIfError(err)

	// Create new client
	client := service.NewClient(conn, hub)

	// Register client into Hub service
	client.Register()

	// Allow client to read and write message.
	go client.ReadMessage()
	go client.WriteMessage()
}
