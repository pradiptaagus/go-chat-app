package handler

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pradiptaagus/go-chat-app/utils"
)

var upgrader = websocket.Upgrader{}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	utils.PanicIfError(err)
	defer conn.Close()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("read: %s\n", err)
			break
		}

		err = conn.WriteMessage(msgType, msg)
		if err != nil {
			fmt.Printf("write: %s\n", err)
			break
		}
	}
}
