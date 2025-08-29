package web

type CreateChatRoomRequest struct {
	RoomID      int64  `json:"roomId"`
	RoomName    string `json:"roomName"`
	Description string `json:"description"`
}
