package types

import (
	"github.com/pion/webrtc/v3"
	"time"
)

type Message struct {
	ID           int                        `json:"id,omitempty"`
	Type         string                     `json:"type"`
	RoomID       int                        `json:"room_id"`
	SenderID     int                        `json:"sender_id"`
	SenderName   string                     `json:"sender_name,omitempty"`
	SenderAvatar string                     `json:"sender_avatar,omitempty"`
	Content      string                     `json:"content,omitempty"`
	SeenBy       []SeenByUser               `json:"seen_by,omitempty"`
	Timestamp    time.Time                  `json:"timestamp,omitempty"`
	IsRead       bool                       `json:"is_read"`
	Offer        *webrtc.SessionDescription `json:"offer,omitempty"`
	Answer       *webrtc.SessionDescription `json:"answer,omitempty"`
	Candidate    *webrtc.ICECandidateInit   `json:"candidate,omitempty"`
}

type MessageStore interface {
	GetMessagesInRoom(roomID int) ([]*Message, error)
	CreateMessage(message *Message) error
	MarkMessageAsSeen(user int, messageID int) error
}

type MessagesResponse struct {
	Messages []*Message `json:"messages"`
}
