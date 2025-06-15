package service

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client interface {
	ReadMessage()
	WriteMessage()
	Send(message string)
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type ClientImpl struct {
	Hub         Hub
	Conn        *websocket.Conn
	outgoingMsg chan []byte
}

func (client *ClientImpl) ReadMessage() {
	func() {
		client.Hub.Unregister(client)
	}()
}

func (Client *ClientImpl) WriteMessage() {

}

func (client *ClientImpl) Send(message string) {

}
