package service

import "log"

// Maintain the set of active clients and broadcast another message to the active clients.
type Hub interface {
	// Observe channel value changes to decide the action to be taken.
	Run()

	// Register client by add the client into the channel
	Register(client Client)

	// Unregister client by delete client in channel if exist
	Unregister(Client Client)

	// Broadcast message to other connected clients
	Broadcast(msg []byte)
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

func (hub *HubImpl) Run() {
	for {
		select {
		// Register connected client
		case client := <-hub.register:
			hub.clients[client] = true
			log.Printf("A new user is connected. Total user %d\n\n", len(hub.clients))

		// Unregister client
		case client := <-hub.unregister:
			if ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				client.Destroy()
				log.Printf("A user is disconnected. Total user %d\n\n", len(hub.clients))
			}

		// Broadcast message to all clients
		case msg := <-hub.broadcast:
			for client := range hub.clients {
				client.Send(msg)
				log.Printf("Broadcasted message: %s\n\n", string(msg))
			}
		}
	}
}

func (hub *HubImpl) Register(client Client) {
	hub.register <- client
	log.Print("Connecting a user\n")
}

func (hub *HubImpl) Unregister(client Client) {
	hub.unregister <- client
	log.Print("Disconnecting a user\n")
}

func (hub *HubImpl) Broadcast(msg []byte) {
	hub.broadcast <- msg
	log.Print("A user broadcasting a message\n")
}
