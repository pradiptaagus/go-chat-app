package service

import (
	"context"
	"log"
	"sync"
)

// Maintain the set of active clients and broadcast another message to the active clients.
type ChatRoom interface {
	// Observe channel value changes to decide the action to be taken.
	Run(ctx context.Context)

	// Register client by add the client into the channel
	Register(ctx context.Context, client Client)

	// Unregister client by delete client in channel if exist
	Unregister(ctx context.Context, Client Client)

	// Broadcast message to other connected clients
	Broadcast(ctx context.Context, msg string)

	// Stop the chat room gracefully.
	// It will close all channels and stop the chat room.
	Stop(ctx context.Context)

	// It is used to notify the chat server that the chat room is stopped.
	// The channel is used to wait for the chat room to be stopped before deleting it from the chat server.
	IsStopped(ctx context.Context) <-chan bool
}

type ChatRoomImpl struct {
	// A channel to store connected clients.
	clients map[Client]bool

	// A channel to store message to the connected clients.
	// When channel get new message, it will be broadcasted to all connected clients.
	broadcast chan string

	// A channel to store a new client.
	// The client in register channel will be moved to clients channel.
	register chan Client

	// A channel to store a client that will be removed.
	// When a value is emitted by unregister channel, it will remove matched client if exist.
	unregister chan Client

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
func NewChatRoomImpl() ChatRoom {
	return &ChatRoomImpl{
		clients:    make(map[Client]bool),
		broadcast:  make(chan string, 2),
		register:   make(chan Client),
		unregister: make(chan Client),
	}
}

func (chatRoom *ChatRoomImpl) Run(ctx context.Context) {
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
			chatRoom.clients[client] = true
			log.Printf("A new user is connected. Total user %d\n", len(chatRoom.clients))
			chatRoom.mu.Unlock()
			go chatRoom.Broadcast(ctx, "A new user has joined the chat")

		// Unregister client
		case client := <-chatRoom.unregister:
			chatRoom.mu.Lock()
			if ok := chatRoom.clients[client]; ok {
				delete(chatRoom.clients, client)
				client.Destroy(ctx)
				log.Printf("A user is disconnected. Total user %d\n", len(chatRoom.clients))
				go chatRoom.Broadcast(ctx, "A user has left the chat")
			}
			chatRoom.mu.Unlock()

		// Broadcast message to all clients
		case msg := <-chatRoom.broadcast:
			chatRoom.mu.Lock()
			for client := range chatRoom.clients {
				go func(c Client, msg string) {
					client.Send(ctx, msg)
					log.Printf("Broadcasted message to client %v, message: %s\n", client, string(msg))
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

func (chatRoom *ChatRoomImpl) Register(ctx context.Context, client Client) {
	select {
	case <-ctx.Done():
		log.Printf("Registering client was cancelled: %v\n", client)
		return
	case chatRoom.register <- client:
		log.Printf("Successfully registering client: %v\n", client)
	}
}

func (chatRoom *ChatRoomImpl) Unregister(ctx context.Context, client Client) {
	select {
	case <-ctx.Done():
		log.Printf("Unregistering client was cancelled: %v\n", client)
		return
	case chatRoom.unregister <- client:
		log.Printf("Successfully unregistering client: %v\n", client)
	}
}

func (chatRoom *ChatRoomImpl) Broadcast(ctx context.Context, msg string) {
	select {
	case <-ctx.Done():
		log.Printf("Broadcasting message was cancelled: %s\n", msg)
		return
	case chatRoom.broadcast <- msg:
		log.Printf("Broadcasting message: %s\n", msg)
	}
}

func (chatRoom *ChatRoomImpl) Stop(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Println("Stoping chat room was cancelled")
		return
	default:
		chatRoom.stop <- true
		log.Println("Stoping chat room")
	}
}

func (chatRoom *ChatRoomImpl) IsStopped(ctx context.Context) <-chan bool {
	select {
	case <-ctx.Done():
		chatRoom.isStopped <- false
		return chatRoom.isStopped
	default:
		return chatRoom.isStopped
	}
}
