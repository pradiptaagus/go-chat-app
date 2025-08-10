package service

import (
	"context"
	"log"
)

// Maintain the set of active clients and broadcast another message to the active clients.
type Hub interface {
	// Observe channel value changes to decide the action to be taken.
	Run(ctx context.Context)

	// Register client by add the client into the channel
	Register(ctx context.Context, client Client)

	// Unregister client by delete client in channel if exist
	Unregister(ctx context.Context, Client Client)

	// Broadcast message to other connected clients
	Broadcast(ctx context.Context, msg []byte)
}

type HubImpl struct {
	// A channel to store connected clients.
	clients map[Client]bool

	// A channel to store message to the connected clients.
	// When channel get new message, it will be broadcasted to all connected clients.
	broadcast chan []byte

	// A channel to store a new client.
	// The client in register channel will be moved to clients channel.
	register chan Client

	// A channel to store a client that will be removed.
	// When a value is emitted by unregister channel, it will remove matched client if exist.
	unregister chan Client
}

// return a new [Hub]
func NewHubImpl() Hub {
	return &HubImpl{
		clients:    make(map[Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan Client),
		unregister: make(chan Client),
	}
}

func (hub *HubImpl) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Hub Run cancelled")
			return

		// Register connected client
		case client := <-hub.register:
			hub.clients[client] = true
			log.Printf("A new user is connected. Total user %d\n", len(hub.clients))

		// Unregister client
		case client := <-hub.unregister:
			if ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				client.Destroy(ctx)
				log.Printf("A user is disconnected. Total user %d\n", len(hub.clients))
			}

		// Broadcast message to all clients
		case msg := <-hub.broadcast:
			for client := range hub.clients {
				select {
				case <-ctx.Done():
					log.Println("Broadcast cancelled due to context done")
					return
				default:
					client.Send(ctx, msg)
					log.Printf("Broadcasted message: %s\n", string(msg))
				}

			}
		}
	}
}

func (hub *HubImpl) Register(ctx context.Context, client Client) {
	select {
	case <-ctx.Done():
		log.Printf("Registering client was cancelled: %v\n", client)
		return
	case hub.register <- client:
		log.Printf("Successfully registering client: %v\n", client)
	}
}

func (hub *HubImpl) Unregister(ctx context.Context, client Client) {
	select {
	case <-ctx.Done():
		log.Printf("Unregistering client was cancelled: %v\n", client)
		return
	case hub.unregister <- client:
		log.Printf("Successfully unregistering client: %v\n", client)
	}
}

func (hub *HubImpl) Broadcast(ctx context.Context, msg []byte) {
	select {
	case <-ctx.Done():
		log.Printf("Broadcasting message was cancelled: %s\n", msg)
		return
	case hub.broadcast <- msg:
		log.Printf("Broadcasting message: %s\n", msg)
	}
}
