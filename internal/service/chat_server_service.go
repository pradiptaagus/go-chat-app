package service

import (
	"context"
	"errors"
	"log"
	"sync"

	"github.com/pradiptaagus/go-chat-app/internal/model/domain"
	"github.com/pradiptaagus/go-chat-app/internal/model/web"
)

type ChatServerService interface {
	GetChatRoom(ctx context.Context, roomId int64) (ChatRoomService, error)
	CreateChatRoom(ctx context.Context, req *web.CreateChatRoomRequest) (ChatRoomService, error)
	DeleteChatRoom(ctx context.Context, roomId int64) error
}

type chatServerServiceImpl struct {
	rooms map[int64]ChatRoomService
	mu    sync.Mutex
}

func (chatServer *chatServerServiceImpl) GetChatRoom(ctx context.Context, roomId int64) (ChatRoomService, error) {
	chatServer.mu.Lock()
	defer chatServer.mu.Unlock()

	room, exist := chatServer.rooms[roomId]
	if !exist {
		log.Printf("Chat room '%d' does not exist.\n", roomId)
		return nil, errors.New("chat room does not exist")
	}

	return room, nil
}

func (chatServer *chatServerServiceImpl) CreateChatRoom(ctx context.Context, req *web.CreateChatRoomRequest) (ChatRoomService, error) {
	chatServer.mu.Lock()
	defer chatServer.mu.Unlock()

	if _, exist := chatServer.rooms[req.RoomID]; exist {
		log.Printf("Chat room '%d' already exists, cannot create duplicate room.\n", req.RoomID)
		return nil, errors.New("chat room already exists")
	}

	room := NewChatRoomService(&domain.ChatRoom{
		ID:          req.RoomID,
		Name:        req.RoomName,
		Description: req.Description,
	})

	chatServer.rooms[req.RoomID] = room

	return room, nil
}

func (chatServer *chatServerServiceImpl) DeleteChatRoom(ctx context.Context, roomId int64) error {
	chatServer.mu.Lock()
	defer chatServer.mu.Unlock()

	if room, exists := chatServer.rooms[roomId]; exists {
		room.Stop(ctx)

		select {
		case <-ctx.Done():
			log.Printf("Context cancelled while stopping room '%d'.\n", roomId)
		case <-room.IsStopped(ctx):
			delete(chatServer.rooms, roomId)
			log.Printf("Chat room '%d' deleted successfully.\n", roomId)
		}

		return nil
	} else {
		log.Printf("Chat room '%d' does not exist.\n", roomId)
		return errors.New("chat room does not exist")
	}
}

func NewChatServerService() ChatServerService {
	return &chatServerServiceImpl{
		rooms: make(map[int64]ChatRoomService),
	}
}
