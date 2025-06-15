package service

// Maintain the set of active clients and broadcast another message to the active clients.
type Hub interface {
	Run()
	Register(client Client)
	Unregister(Client Client)
}

type HubImpl struct {
	clients    map[Client]bool
	broadcast  chan []byte
	register   chan Client
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
			}

		// Broadcast message to all client
		case message := <-hub.broadcast:
			for client := range hub.clients {
				client.Send(string(message))

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
