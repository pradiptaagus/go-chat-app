package service

import (
	"context"
	"log"
	"sync"
)

type ChatServer interface {
	GetOrCreateRoom(ctx context.Context, roomId string) ChatRoom
	DeleteChatRoom(ctx context.Context, roomId string)
}

type chatServerImpl struct {
	rooms map[string]ChatRoom
	mu    sync.Mutex
}

func (chatServer *chatServerImpl) GetOrCreateRoom(ctx context.Context, roomId string) ChatRoom {
	chatServer.mu.Lock()
	defer chatServer.mu.Unlock()

	if room, exists := chatServer.rooms[roomId]; exists {
		log.Printf("Chat room '%s' already exists, returning existing room.\n", roomId)
		return room
	}

	room := NewChatRoomImpl()
	chatServer.rooms[roomId] = room

	go room.Run(ctx)

	return room
}

func (chatServer *chatServerImpl) DeleteChatRoom(ctx context.Context, roomId string) {
	chatServer.mu.Lock()
	defer chatServer.mu.Unlock()

	if room, exists := chatServer.rooms[roomId]; exists {
		room.Stop(ctx)

		select {
		case <-ctx.Done():
			log.Printf("Context cancelled while stopping room '%s'.\n", roomId)
		case <-room.IsStopped(ctx):
			delete(chatServer.rooms, roomId)
			log.Printf("Chat room '%s' deleted successfully.\n", roomId)
		}
	} else {
		log.Printf("Chat room '%s' does not exist.\n", roomId)
	}
}

func NewChatServerImpl() ChatServer {
	return &chatServerImpl{
		rooms: make(map[string]ChatRoom),
	}
}
