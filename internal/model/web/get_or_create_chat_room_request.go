package web

type GetOrCreateChatRoomRequest struct {
	RoomID      int64  `json:"roomId"`
	RoomName    string `json:"roomName"`
	Description string `json:"description"`
}
