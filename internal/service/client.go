package service

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client interface {
	// ReadMessage reads messages from the websocket connection.
	// The application runs ReadMessage for each connected client in the hub.
	// The application ensures that there is at most one ReadMessage on a connected client in the hub.
	ReadMessage(ctx context.Context)

	// WriteMessage write messages from the hub to the websocket connection.
	// The application runs WriteMessage for each connection.
	// The application ensures that there is at most one WriteMessage on a connection.
	WriteMessage(ctx context.Context)

	// Send message to all connected clients
	Send(ctx context.Context, message []byte)

	// Register client to the hub
	Register(ctx context.Context)

	// Destroy all goroutines related to the current user
	Destroy(ctx context.Context)
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	// For now, it would be 1MB. It can be adjusted based on the possible message size.
	// This is used to prevent DoS attack by sending large messages.
	// The application will close the connection if the message size exceeds this limit.
	maxMessageSize = 1024 * 1024
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

func (client *ClientImpl) ReadMessage(ctx context.Context) {
	// Unregister user when disconnected.
	// Typically invoked when got error [websocket.IsUnexpectedCloseError].
	defer func() {
		client.hub.Unregister(ctx, client)
		client.conn.Close()
	}()

	// Set read limit for the message text
	client.conn.SetReadLimit(maxMessageSize)

	// Set read deadline for the first connection created
	client.conn.SetReadDeadline(time.Now().Add(pongWait))

	// If the server doesn't get a Pong within `pongWait` time, it will assume the connection is dead and close it.
	// To handle this condition, set `SetPongHandler` callback which is triggered when a Pong message is received.
	// It's typically used to keep connections alive by resetting read deadlines after a message received.
	client.conn.SetPongHandler(func(appData string) error {
		log.Println("Pong received, resetting read deadline")
		client.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Read message
		_, msg, err := client.conn.ReadMessage()

		if err != nil {
			log.Printf("ReadMessage Error: %v\n", err)

			// If unexpected close error occured, it will break the loop for current client.
			// [websocket.IsUnexpectedCloseError] indicates user is disconnected.
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v\n", err)
			}
			break
		}

		log.Printf("Unformatted incoming message: %v\n", string(msg))

		// Trim and replace unwanted characters of the message
		// It will replace `newLine` character by `space` character.
		msg = bytes.TrimSpace(bytes.ReplaceAll(msg, newLine, space))

		log.Printf("Formatted incoming message %v\n", string(msg))

		// Broadcast message to the clients
		client.hub.Broadcast(ctx, msg)
	}
}

// WriteMessage writes messages to the websocket connection.
// It reads messages from the outgoingMsg channel and writes them to the connection.
// It also sends ping messages to the peer to keep the connection alive.
// If the peer doesn't respond to the ping message within `pongWait` time, it will close the connection.
// The application will close the connection if it doesn't receive a Pong message.
func (client *ClientImpl) WriteMessage(ctx context.Context) {
	// Create ticker to send ping message every `pingPeriod` time.
	// This will keep the connection alive by sending ping message to the peer.
	// If the peer doesn't respond to the ping message within `pongWait` time, it will close the connection.
	// The application will close the connection if it doesn't receive a Pong message
	// within `pongWait` time.
	ticker := time.NewTicker(pingPeriod)

	// Close the ticker and connection when the function exits.
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("WriteMessage was cancelled: %v\n", client)
			return

		// Read outgoing message
		case msg, ok := <-client.outgoingMsg:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			log.Printf("WriteMessage: %s, status: %t\n", msg, ok)
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

		// Run loop every tick time to send ping message.
		// This will send a ping message to the peer to keep the connection alive.
		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WriteMessage ticker error: %v", err)
				return
			}
		}
	}
}

func (client *ClientImpl) Send(ctx context.Context, msg []byte) {
	select {
	case <-ctx.Done():
		log.Printf("Sending message was cancelled: %s\n", msg)
		return
	case client.outgoingMsg <- msg:
		log.Printf("Message sent: %s\n", msg)
	}
}

func (client *ClientImpl) Register(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Printf("Registering client was cancelled: %v\n", client)
		return
	default:
		client.hub.Register(ctx, client)
		log.Printf("Successfully registered client: %v\n", client)
	}
}

func (client *ClientImpl) Destroy(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Printf("Destroying client was cancelled: %v\n", client)
		return
	default:
		log.Printf("Destroying client: %v\n", client)
		close(client.outgoingMsg)
	}
}

// return a new [Client]
func NewClient(conn *websocket.Conn, hub Hub) Client {
	return &ClientImpl{
		hub:         hub,
		conn:        conn,
		outgoingMsg: make(chan []byte),
	}
}
