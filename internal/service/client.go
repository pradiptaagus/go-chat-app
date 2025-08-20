package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Client interface {
	// ReadPump reads messages from the websocket connections.
	// The application runs ReadPump for each connected client in the chat room.
	// The application ensures that there is at most one ReadPump on a connected client in the chat room.
	ReadPump(ctx context.Context)

	// WritePump write messages for all clients in the same rooms.
	// The application runs WritePump for each connection.
	// The application ensures that there is at most one WritePump on a connection.
	WritePump(ctx context.Context)

	// Send message to all connected clients
	Send(ctx context.Context, message string)

	// Register client to the chat room
	register(ctx context.Context, roomId string)

	// Handle client to leave the chat room
	leave(ctx context.Context, roomId string)

	// Destroy all goroutines related to the current user
	Destroy(ctx context.Context)
}

// MessageType defines the type of message sent by the client.
// It can be "join", "leave", or "message".
type MessageType string

const (
	MessageTypeJoin  MessageType = "join"
	MessageTypeLeave MessageType = "leave"
	MessageTypeText  MessageType = "message"
)

type Message struct {
	Type    MessageType `json:"type"`
	RoomId  string      `json:"roomId"`
	Content string      `json:"content"`
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
	newLine string = "\n"
	space   string = " "
)

type clientImpl struct {
	chatServer  ChatServer
	chatRoom    map[string]ChatRoom
	conn        *websocket.Conn
	outgoingMsg chan string
}

func (client *clientImpl) ReadPump(ctx context.Context) {
	// Unregister user when disconnected.
	defer func() {
		if client.chatRoom != nil {
			for _, room := range client.chatRoom {
				room.Unregister(ctx, client)
			}
		}
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
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("ReadMessage Error: %v\n", err)

			// If unexpected close error occured, it will break the loop for current client.
			// [websocket.IsUnexpectedCloseError] indicates user is disconnected.
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v\n", err)
			}
			return
		}

		err = client.handleReadPumpMessage(ctx, string(msg))
		if err != nil {
			log.Printf("Error handling message: %v", err)
			return
		}
	}
}

// WritePump writes messages to the clients with same rooms.
// It reads messages from the outgoingMsg channel and writes them to the connection.
// It also sends ping messages to the peer to keep the connection alive.
// If the peer doesn't respond to the ping message within `pongWait` time, it will close the connection.
// The application will close the connection if it doesn't receive a Pong message.
func (client *clientImpl) WritePump(ctx context.Context) {
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
			err := client.conn.WriteMessage(websocket.TextMessage, []byte(msg))
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

func (client *clientImpl) Send(ctx context.Context, msg string) {
	select {
	case <-ctx.Done():
		log.Printf("Sending message was cancelled: %s\n", msg)
		return
	case client.outgoingMsg <- msg:
		log.Printf("Message sent: %s\n", msg)
	}
}

func (client *clientImpl) register(ctx context.Context, roomId string) {
	select {
	case <-ctx.Done():
		log.Printf("Registering client was cancelled: %v\n", client)
		return
	default:
		_, ok := client.chatRoom[roomId]
		if !ok {
			log.Printf("Chat room %s not found for client %v\n", roomId, client)

			// If the chat room is not found, create a new one.
			client.chatRoom[roomId] = client.chatServer.GetOrCreateRoom(ctx, roomId)
		}
		log.Printf("Registering client %v to chat room %s\n", client, roomId)
		client.chatRoom[roomId].Register(ctx, client)
		log.Printf("Successfully registered client: %v\n", client)
	}
}

func (client *clientImpl) leave(ctx context.Context, roomId string) {
	select {
	case <-ctx.Done():
		log.Printf("Leaving chat room was cancelled: %v\n", client)
		return
	default:
		room, ok := client.chatRoom[roomId]
		if !ok {
			log.Printf("Chat room %s not found for client %v\n", roomId, client)
			return
		}
		log.Printf("Unregistering client %v from chat room %s\n", client, roomId)
		room.Unregister(ctx, client)
		delete(client.chatRoom, roomId)
		log.Printf("Successfully unregistered client: %v from chat room %s\n", client, roomId)
	}
}

func (client *clientImpl) Destroy(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Printf("Destroying client was cancelled: %v\n", client)
		return
	default:
		log.Printf("Destroying client: %v\n", client)
		close(client.outgoingMsg)
	}
}

// Handle read pump message.
// The message is a string object ans should be unmarshalled into a Message struct.
// It will handle different types of messages like join, leave, or text messages.
func (client *clientImpl) handleReadPumpMessage(ctx context.Context, msg string) error {
	// Unmarshal the message into a Message struct, because it's a stringified JSON object.
	var message Message
	log.Printf("Unformatted incoming message: %v\n", msg)
	if err := json.Unmarshal([]byte(msg), &message); err != nil {
		log.Printf("Error unmarshalling message: %v", err)
		return err
	}
	log.Printf("Unmarshalled message: %v\n", message)

	content := message.Content
	log.Printf("Unformatted incoming message: %v\n", content)

	// Trim and replace unwanted characters of the message
	// It will replace `newLine` character by `space` character.
	msg = strings.TrimSpace(content)
	msg = strings.ReplaceAll(msg, newLine, space)

	log.Printf("Formatted incoming message %v\n", string(msg))

	switch message.Type {
	case MessageTypeJoin:
		client.register(ctx, message.RoomId)
	case MessageTypeText:
		if room, ok := client.chatRoom[message.RoomId]; ok {
			log.Printf("Sending message to room %s: %s\n", message.RoomId, msg)
			room.Broadcast(ctx, msg)
		} else {
			log.Printf("Chat room %s not found for client %v\n", message.RoomId, client)
			return fmt.Errorf("chat room %s not found", message.RoomId)
		}
	case MessageTypeLeave:
		client.leave(ctx, message.RoomId)
	}

	return nil
}

// return a new [Client]
func NewClient(conn *websocket.Conn, chatServer ChatServer) Client {
	return &clientImpl{
		chatServer:  chatServer,
		chatRoom:    make(map[string]ChatRoom),
		conn:        conn,
		outgoingMsg: make(chan string),
	}
}
