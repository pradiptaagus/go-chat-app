package service

// Maintain the set of active clients and broadcast another message to the active clients.
type Hub interface {
	// Watch channel value changes to decide the action to be taken.
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
	// When a value emitted by unregister channel, it will remove matched client if exist.
	unregister chan Client
}

func NewHubImpl() *HubImpl {
	return &HubImpl{
		clients:    make(map[Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan Client),
		unregister: make(chan Client),
	}
}

// Observe channel changes and decide actions will be taken.
func (hub *HubImpl) Run() {
	for {
		select {
		// Register connected client
		case client := <-hub.register:
			hub.clients[client] = true

		// Unregister client
		case client := <-hub.unregister:
			if ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				client.FinalizeSend()
			}

		// Broadcast message to all clients
		case msg := <-hub.broadcast:
			for client := range hub.clients {
				client.Send(msg)
			}
		}
	}
}

func (hub *HubImpl) Register(client Client) {
	hub.register <- client
}

func (hub *HubImpl) Unregister(client Client) {
	hub.unregister <- client
}

func (hub *HubImpl) Broadcast(msg []byte) {
	hub.broadcast <- msg
}
