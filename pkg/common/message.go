package common

// MessageType defines the type of message sent by the client.
// It can be "join", "leave", or "message".
type MessageType string

const (
	MessageTypeJoin  MessageType = "join"
	MessageTypeLeave MessageType = "leave"
	MessageTypeText  MessageType = "message"
)

type Message struct {
	ID       int64       `json:"id,omitempty"`
	Type     MessageType `json:"type"`
	RoomID   int64       `json:"roomId"`
	ClientID int64       `json:"clientId"`
	Content  string      `json:"content"`
}
