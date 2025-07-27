package handler

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pradiptaagus/go-chat-app/internal/service"
	"github.com/pradiptaagus/go-chat-app/utils"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Allow all connections
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WsHandler(w http.ResponseWriter, r *http.Request, hub service.Hub) {
	conn, err := upgrader.Upgrade(w, r, nil)
	utils.PanicIfError(err)
	// defer conn.Close()

	// Create new client
	client := service.NewClient(conn, hub)

	// Register client into Hub service
	client.Register()

	// Run ReadMessage and WriteMessage goroutine for current connection
	go client.ReadMessage()
	go client.WriteMessage()

	// for {
	// 	msgType, msg, err := conn.ReadMessage()
	// 	if err != nil {
	// 		fmt.Printf("read: %s\n", err)
	// 		break
	// 	}

	// 	err = conn.WriteMessage(msgType, msg)
	// 	if err != nil {
	// 		fmt.Printf("write: %s\n", err)
	// 		break
	// 	}
	// }
}
