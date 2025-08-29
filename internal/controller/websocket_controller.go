package controller

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pradiptaagus/go-chat-app/internal/service"
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

type websocketController struct {
}

// Handles websokcet requests from the peer
func (ws *websocketController) WsHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, chatServerService service.ChatServerService, websocketService service.WebSocketService) {
	conn := websocketService.Must(websocketService.CreateWsConn(ctx, w, r))

	// Create new client
	client := service.NewClientService(conn, chatServerService, nil)

	// Allow client to read and write message.
	go client.ReadPump(ctx)
	go client.WritePump(ctx)
}

func NewWebsocketController() *websocketController {
	return &websocketController{}
}
