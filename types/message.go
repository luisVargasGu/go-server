package types

import (
	"github.com/pion/webrtc/v3"
	"time"
)

// TODO: move ID, name and Avater to a separate Struct
// TODO: move some fields to a different type so it's more performant
// If they are always together
type Message struct {
	ID              int                        `json:"id,omitempty"`
	Type            string                     `json:"type"`
	TrackID         string                     `json:"track_id,omitempty"`
	RoomID          int                        `json:"room_id"`
	SenderID        int                        `json:"sender_id"`
	SenderName      string                     `json:"sender_name,omitempty"`
	SenderAvatar    string                     `json:"sender_avatar,omitempty"`
	Content         string                     `json:"content,omitempty"`
	SeenBy          []SeenByUser               `json:"seen_by,omitempty"`
	Timestamp       time.Time                  `json:"timestamp,omitempty"`
	IsRead          bool                       `json:"is_read"`
	IsVideoEnabled  *bool                      `json:"isVideoEnabled,omitempty"`
	IsScreenEnabled *bool                      `json:"isScreenEnabled,omitempty"`
	IsMicEnabled    *bool                      `json:"isMicEnabled,omitempty"`
	Offer           *webrtc.SessionDescription `json:"offer,omitempty"`
	Answer          *webrtc.SessionDescription `json:"answer,omitempty"`
	Candidate       *webrtc.ICECandidateInit   `json:"candidate,omitempty"`
}

type MessageStore interface {
	GetMessagesInRoom(roomID int) ([]*Message, error)
	CreateMessage(message *Message) error
	MarkMessageAsSeen(user int, messageID int) error
}

type MessagesResponse struct {
	Messages []*Message `json:"messages"`
}
