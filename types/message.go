package types

import "time"

type Message struct {
	ID           int          `json:"id"`
	RoomID       int          `json:"room_id"`
	SenderID     int          `json:"sender_id"`
	SenderName   string       `json:"sender_name"`
	SenderAvatar string       `json:"sender_avatar"`
	Content      string       `json:"content"`
	SeenBy       []SeenByUser `json:"seen_by"`
	Timestamp    time.Time    `json:"timestamp"`
	IsRead       bool         `json:"is_read"`
}

type MessageStore interface {
	GetMessagesInRoom(roomID int) ([]*Message, error)
	CreateMessage(message Message) error
	MarkMessageAsSeen(user int, messageID int) error
}

type MessagesResponse struct {
	Messages []*Message `json:"messages"`
}
