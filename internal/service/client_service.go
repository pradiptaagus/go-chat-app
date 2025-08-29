package service

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pradiptaagus/go-chat-app/internal/model/domain"
	"github.com/pradiptaagus/go-chat-app/pkg/common"
)

type ClientService interface {
	// ReadPump reads messages from the websocket connections.
	// The application runs ReadPump for each connected client in the chat room.
	// The application ensures that there is at most one ReadPump on a connected client in the chat room.
	ReadPump(ctx context.Context)

	// WritePump write messages for all clients in the same rooms.
	// The application runs WritePump for each connection.
	// The application ensures that there is at most one WritePump on a connection.
	WritePump(ctx context.Context)

	// Create a new client
	Create(ctx context.Context, w http.ResponseWriter, r *http.Request) (*domain.Client, error)

	// Find all clients
	FindAll(ctx context.Context) ([]*domain.Client, error)

	// Send message to all connected clients
	send(ctx context.Context, message *common.Message)

	// Register client to the chat room
	register(ctx context.Context, roomId int64)

	// Handle client to leave the chat room
	leave(ctx context.Context, roomId int64)

	// Destroy all goroutines related to the current user
	destroy(ctx context.Context)
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

type clientServiceImpl struct {
	*domain.Client
	chatServer  ChatServerService
	chatRoom    map[int64]ChatRoomService
	conn        *websocket.Conn
	outgoingMsg chan *common.Message
}

// return a new [domain.Client]
func NewClientService(conn *websocket.Conn, chatServer ChatServerService, client *domain.Client) ClientService {
	return &clientServiceImpl{
		Client:      client,
		chatServer:  chatServer,
		chatRoom:    make(map[int64]ChatRoomService),
		conn:        conn,
		outgoingMsg: make(chan *common.Message),
	}
}

// func (client *clientServiceImpl) ReadPump(ctx context.Context) {
// 	// Unregister user when disconnected.
// 	defer func() {
// 		if client.chatRoom != nil {
// 			for _, room := range client.chatRoom {
// 				room.Unregister(ctx, client)
// 			}
// 		}
// 		client.conn.Close()
// 	}()

// 	// Set read limit for the message text
// 	client.conn.SetReadLimit(maxMessageSize)

// 	// Set read deadline for the first connection created
// 	client.conn.SetReadDeadline(time.Now().Add(pongWait))

// 	// If the server doesn't get a Pong within `pongWait` time, it will assume the connection is dead and close it.
// 	// To handle this condition, set `SetPongHandler` callback which is triggered when a Pong message is received.
// 	// It's typically used to keep connections alive by resetting read deadlines after a message received.
// 	client.conn.SetPongHandler(func(appData string) error {
// 		log.Println("Pong received, resetting read deadline")
// 		client.conn.SetReadDeadline(time.Now().Add(pongWait))
// 		return nil
// 	})

// 	for {
// 		_, msg, err := client.conn.ReadMessage()
// 		if err != nil {
// 			log.Printf("ReadMessage Error: %v\n", err)

// 			// If unexpected close error occured, it will break the loop for current client.
// 			// [websocket.IsUnexpectedCloseError] indicates user is disconnected.
// 			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
// 				log.Printf("error: %v\n", err)
// 			}
// 			return
// 		}

// 		err = client.handleReadPumpMessage(ctx, string(msg))
// 		if err != nil {
// 			log.Printf("Error handling message: %v", err)
// 			return
// 		}
// 	}
// }

// // WritePump writes messages to the clients with same rooms.
// // It reads messages from the outgoingMsg channel and writes them to the connection.
// // It also sends ping messages to the peer to keep the connection alive.
// // If the peer doesn't respond to the ping message within `pongWait` time, it will close the connection.
// // The application will close the connection if it doesn't receive a Pong message.
// func (client *clientServiceImpl) WritePump(ctx context.Context) {
// 	// Create ticker to send ping message every `pingPeriod` time.
// 	// This will keep the connection alive by sending ping message to the peer.
// 	// If the peer doesn't respond to the ping message within `pongWait` time, it will close the connection.
// 	// The application will close the connection if it doesn't receive a Pong message
// 	// within `pongWait` time.
// 	ticker := time.NewTicker(pingPeriod)

// 	// Close the ticker and connection when the function exits.
// 	defer func() {
// 		ticker.Stop()
// 		client.conn.Close()
// 	}()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			log.Printf("WriteMessage was cancelled: %v\n", client)
// 			return

// 		// Read outgoing message
// 		case msg, ok := <-client.outgoingMsg:
// 			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
// 			log.Printf("WriteMessage: %v, status: %t\n", msg, ok)
// 			if !ok {
// 				// Write close message when failed to get outgoingMsg channel
// 				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
// 				return
// 			}

// 			byteMsg, errMarshal := json.Marshal(msg)
// 			if errMarshal != nil {
// 				log.Printf("Error marshalling message: %v", errMarshal)
// 				return
// 			}

// 			// Write outgoing message from outgoingMsg channel
// 			err := client.conn.WriteMessage(websocket.TextMessage, byteMsg)
// 			if err != nil {
// 				log.Printf("WriteMessage error: %v", err)
// 				return
// 			}

// 		// Run loop every tick time to send ping message.
// 		// This will send a ping message to the peer to keep the connection alive.
// 		case <-ticker.C:
// 			client.conn.SetWriteDeadline(time.Now().Add(writeWait))
// 			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
// 				log.Printf("WriteMessage ticker error: %v", err)
// 				return
// 			}
// 		}
// 	}
// }

// func (client *clientServiceImpl) send(ctx context.Context, msg *common.Message) {
// 	select {
// 	case <-ctx.Done():
// 		log.Printf("Sending message was cancelled: %v\n", msg)
// 		return
// 	case client.outgoingMsg <- msg:
// 		log.Printf("Message sent: %v\n", msg)
// 	}
// }

// func (client *clientServiceImpl) register(ctx context.Context, roomId int64) {
// 	select {
// 	case <-ctx.Done():
// 		log.Printf("Registering client was cancelled: %v\n", client)
// 		return
// 	default:
// 		_, ok := client.chatRoom[roomId]
// 		if !ok {
// 			log.Printf("Chat room %d not found for client %v\n", roomId, client)
// 			return
// 		}
// 		log.Printf("Registering client %v to chat room %d\n", client, roomId)
// 		client.chatRoom[roomId].Register(ctx, client)
// 		log.Printf("Successfully registered client: %v\n", client)
// 	}
// }

// func (client *clientServiceImpl) leave(ctx context.Context, roomId int64) {
// 	select {
// 	case <-ctx.Done():
// 		log.Printf("Leaving chat room was cancelled: %v\n", client)
// 		return
// 	default:
// 		room, ok := client.chatRoom[roomId]
// 		if !ok {
// 			log.Printf("Chat room %d not found for client %v\n", roomId, client)
// 			return
// 		}
// 		log.Printf("Unregistering client %v from chat room %d\n", client, roomId)
// 		room.Unregister(ctx, client)
// 		delete(client.chatRoom, roomId)
// 		log.Printf("Successfully unregistered client: %v from chat room %d\n", client, roomId)
// 	}
// }

// func (client *clientServiceImpl) destroy(ctx context.Context) {
// 	select {
// 	case <-ctx.Done():
// 		log.Printf("Destroying client was cancelled: %v\n", client)
// 		return
// 	default:
// 		log.Printf("Destroying client: %v\n", client)
// 		close(client.outgoingMsg)
// 	}
// }

// // Handle read pump message.
// // The message is a string object ans should be unmarshalled into a Message struct.
// // It will handle different types of messages like join, leave, or text messages.
// func (client *clientServiceImpl) handleReadPumpMessage(ctx context.Context, msg string) error {
// 	// Unmarshal the message into a Message struct, because it's a stringified JSON object.
// 	var message common.Message
// 	log.Printf("Unformatted incoming message: %v\n", msg)
// 	if err := json.Unmarshal([]byte(msg), &message); err != nil {
// 		log.Printf("Error unmarshalling message: %v", err)
// 		return err
// 	}
// 	log.Printf("Unmarshalled message: %v\n", message)

// 	content := message.Content
// 	log.Printf("Unformatted incoming message: %v\n", content)

// 	// Trim and replace unwanted characters of the message
// 	// It will replace `newLine` character by `space` character.
// 	msg = strings.TrimSpace(content)
// 	msg = strings.ReplaceAll(msg, newLine, space)

// 	log.Printf("Formatted incoming message %v\n", string(msg))

// 	switch message.Type {
// 	case common.MessageTypeJoin:
// 		client.register(ctx, message.RoomID)
// 	case common.MessageTypeText:
// 		if room, ok := client.chatRoom[message.RoomID]; ok {
// 			log.Printf("Sending message to room %d: %s\n", message.RoomID, msg)
// 			room.Broadcast(ctx, &common.Message{
// 				ID:       util.GenerateRandomID(),
// 				Content:  msg,
// 				Type:     common.MessageTypeText,
// 				RoomID:   message.RoomID,
// 				ClientID: message.ClientID,
// 			})
// 		} else {
// 			log.Printf("Chat room %d not found for client %v\n", message.RoomID, client)
// 			return fmt.Errorf("chat room %d not found", message.RoomID)
// 		}
// 	case common.MessageTypeLeave:
// 		client.leave(ctx, message.RoomID)
// 	}

// 	return nil
// }
