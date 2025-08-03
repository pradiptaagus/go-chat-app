package service

import (
	"bytes"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client interface {
	// ReadMessage reads messages from the websocket connection.
	// The application runs ReadMessage for each connected client in the hub.
	// The application ensures that there is at most one ReadMessage on a connected client in the hub.
	ReadMessage()

	// WriteMessage write messages from the hub to the websocket connection.
	// The application runs WriteMessage for each connection.
	// The application ensures that there is at most one WriteMessage on a connection.
	WriteMessage()

	// Send message to all connected clients
	Send(message []byte)

	// Register client to the hub
	Register()

	// Destroy all goroutines related to the current user
	Destroy()
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

var (
	newLine = []byte{'\n'}
	space   = []byte{' '}
)

type ClientImpl struct {
	hub         Hub
	conn        *websocket.Conn
	outgoingMsg chan []byte
}

func (client *ClientImpl) ReadMessage() {
	// Unregister user when disconnected.
	// Typically invoked when got error [websocket.IsUnexpectedCloseError].
	defer func() {
		client.hub.Unregister(client)
		client.conn.Close()
	}()

	// Set read limit for the message text
	client.conn.SetReadLimit(maxMessageSize)

	// Set read deadline for the first connection created
	client.conn.SetReadDeadline(time.Now().Add(pongWait))

	// If the server doesn't get a Pong within `pongWait` time, it will assume the connection is dead and close it.
	// to handle this, set SetPongHandler callback that is triggered when a Pong message is received.
	// It's typically used to keep connections alive by resetting read deadlines after heartbeats.
	client.conn.SetPongHandler(func(appData string) error {
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Read message
		_, msg, err := client.conn.ReadMessage()

		log.Printf("ReadMessage: %v\n", err)

		if err != nil {

			// If unexpected close error occured, it will break the loop for current client.
			// [websocket.IsUnexpectedCloseError] indicates user is disconnected.
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v\n\n", err)
			}
			break
		}

		log.Printf("Unformatted incoming message: %v\n", string(msg))

		// Trim and replace unwanted characters of the message
		// It will replace `newLine` character by `space` character.
		msg = bytes.TrimSpace(bytes.ReplaceAll(msg, newLine, space))

		log.Printf("Formatted incoming message %v\n\n", string(msg))

		// Broadcast message to the clients
		client.hub.Broadcast(msg)
	}
}

func (client *ClientImpl) WriteMessage() {
	// Create ticker
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		client.conn.SetWriteDeadline(time.Now().Add(writeWait))
		select {
		// Read outgoing message
		case msg, ok := <-client.outgoingMsg:
			log.Printf("WriteMessage: %s, status: %t\n\n", msg, ok)
			if !ok {
				// Write close message when failed to get outgoingMsg channel
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write outgoing message from outgoindMsg channel
			err := client.conn.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Printf("WriteMessage error: %v", err)
				return
			}

		// Run loop every tick time
		case <-ticker.C:
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WriteMessage ticker error: %v", err)
				return
			}
		}
	}
}

func (client *ClientImpl) Send(msg []byte) {
	client.outgoingMsg <- msg
}

func (client *ClientImpl) Register() {
	client.hub.Register(client)
}

func (client *ClientImpl) Destroy() {
	close(client.outgoingMsg)
}

// return a new [Client]
func NewClient(conn *websocket.Conn, hub Hub) Client {
	return &ClientImpl{
		hub:         hub,
		conn:        conn,
		outgoingMsg: make(chan []byte),
	}
}
