package service

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/pradiptaagus/go-chat-app/internal/model/domain"
	"github.com/pradiptaagus/go-chat-app/pkg/common"
	"github.com/pradiptaagus/go-chat-app/pkg/util"
)

// Maintain the set of active clients and broadcast another message to the active clients.
type ChatRoomService interface {
	// Observe channel value changes to decide the action to be taken.
	Run(ctx context.Context)

	// Register client by add the client into the channel
	Register(ctx context.Context, client *clientServiceImpl)

	// Unregister client by delete client in channel if exist
	Unregister(ctx context.Context, client *clientServiceImpl)

	// Broadcast message to other connected clients
	Broadcast(ctx context.Context, msg *common.Message)

	// Stop the chat room gracefully.
	// It will close all channels and stop the chat room.
	Stop(ctx context.Context)

	// It is used to notify the chat server that the chat room is stopped.
	// The channel is used to wait for the chat room to be stopped before deleting it from the chat server.
	IsStopped(ctx context.Context) <-chan bool

	GetChatRoom(ctx context.Context) *domain.ChatRoom
}

type chatRoomServiceImpl struct {
	*domain.ChatRoom

	// A channel to store connected clients.
	clients map[int64]*clientServiceImpl

	// A channel to store message to the connected clients.
	// When channel get new message, it will be broadcasted to all connected clients.
	broadcast chan *common.Message

	// A channel to store a new
	// The client in register channel will be moved to clients channel.
	register chan *clientServiceImpl

	// A channel to store a client that will be removed.
	// When a value is emitted by unregister channel, it will remove matched client if exist.
	unregister chan *clientServiceImpl

	// A channel to stop the chat room.
	// When a value is emitted by stop channel, it will close all channels and stop the chat room.
	// The channel is used to stop the chat room gracefully.
	// It will wait until all goroutines are stopped before returning.
	// This is useful to avoid memory leaks and ensure that all resources are released properly.
	stop chan bool

	// A channel to indicate that the chat room is stopped.
	// It is used to notify the chat server that the chat room is stopped.
	// The channel is used to wait for the chat room to be stopped before deleting it from the chat server.
	isStopped chan bool

	// Mutex to protect concurrent access to the clients map.
	mu sync.Mutex
}

// return a new [ChatRoom]
func NewChatRoomService(chatRoom *domain.ChatRoom) ChatRoomService {
	return &chatRoomServiceImpl{
		ChatRoom:   chatRoom,
		clients:    make(map[int64]*clientServiceImpl),
		broadcast:  make(chan *common.Message, 2),
		register:   make(chan *clientServiceImpl),
		unregister: make(chan *clientServiceImpl),
	}
}

func (chatRoom *chatRoomServiceImpl) Run(ctx context.Context) {
	defer func() {
		log.Println("ChatRoom Run stopped")
	}()

	for {
		select {
		case <-ctx.Done():
			log.Println("ChatRoom Run cancelled")
			return

		// Register connected client
		case client := <-chatRoom.register:
			chatRoom.mu.Lock()
			chatRoom.clients[client.ID] = client
			log.Printf("A new user is connected. Total user %d\n", len(chatRoom.clients))
			chatRoom.mu.Unlock()
			go chatRoom.Broadcast(ctx, &common.Message{
				ID:       util.GenerateRandomID(),
				Type:     common.MessageTypeText,
				Content:  "A new user has joined the chat",
				ClientID: client.ID,
				RoomID:   chatRoom.ID,
			})

		// Unregister client
		case client := <-chatRoom.unregister:
			chatRoom.mu.Lock()
			if _, ok := chatRoom.clients[client.ID]; ok {
				delete(chatRoom.clients, client.ID)
				client.destroy(ctx)
				log.Printf("A user is disconnected. Total user %d\n", len(chatRoom.clients))
				go chatRoom.Broadcast(ctx, &common.Message{
					ID:       util.GenerateRandomID(),
					Type:     common.MessageTypeText,
					Content:  "A user has left the chat",
					ClientID: client.ID,
					RoomID:   chatRoom.ID,
				})
			}
			chatRoom.mu.Unlock()

		// Broadcast message to all clients
		case msg := <-chatRoom.broadcast:
			chatRoom.mu.Lock()
			for _, client := range chatRoom.clients {
				go func(_client ClientService, msg *common.Message) {
					client.send(ctx, msg)
					byteMsg, err := json.Marshal(msg)
					if err != nil {
						log.Printf("Error marshalling message: %v", err)
						return
					}
					log.Printf("Broadcasted message to client %v, message: %s\n", _client, string(byteMsg))
				}(client, msg)
			}
			chatRoom.mu.Unlock()

		case <-chatRoom.stop:
			chatRoom.mu.Lock()
			close(chatRoom.broadcast)
			close(chatRoom.register)
			close(chatRoom.unregister)
			chatRoom.isStopped <- true
			chatRoom.mu.Unlock()
			return
		}
	}
}

func (chatRoom *chatRoomServiceImpl) Register(ctx context.Context, client *clientServiceImpl) {
	select {
	case <-ctx.Done():
		log.Printf("Registering client was cancelled: %v\n", client)
		return
	case chatRoom.register <- client:
		log.Printf("Successfully registering client: %v\n", client)
	}
}

func (chatRoom *chatRoomServiceImpl) Unregister(ctx context.Context, client *clientServiceImpl) {
	select {
	case <-ctx.Done():
		log.Printf("Unregistering client was cancelled: %v\n", client)
		return
	case chatRoom.unregister <- client:
		log.Printf("Successfully unregistering client: %v\n", client)
	}
}

func (chatRoom *chatRoomServiceImpl) Broadcast(ctx context.Context, msg *common.Message) {
	select {
	case <-ctx.Done():
		log.Printf("Broadcasting message was cancelled: %v\n", msg)
		return
	case chatRoom.broadcast <- msg:
		log.Printf("Broadcasting message: %v\n", msg)
	}
}

func (chatRoom *chatRoomServiceImpl) Stop(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Println("Stoping chat room was cancelled")
		return
	default:
		chatRoom.stop <- true
		log.Println("Stoping chat room")
	}
}

func (chatRoom *chatRoomServiceImpl) IsStopped(ctx context.Context) <-chan bool {
	select {
	case <-ctx.Done():
		chatRoom.isStopped <- false
		return chatRoom.isStopped
	default:
		return chatRoom.isStopped
	}
}

func (chatRoom *chatRoomServiceImpl) GetChatRoom(ctx context.Context) *domain.ChatRoom {
	select {
	case <-ctx.Done():
		log.Println("GetChatRoom was cancelled")
		return nil
	default:
		return chatRoom.ChatRoom
	}
}
