package types

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"createdAt"`
}

type Channel struct {
	ID    int
	Name  string
	Rooms map[int]*Room
}

type Message struct {
	ID        int
	RoomID    int
	SenderID  string
	Content   string
	Timestamp time.Time
	IsRead    bool
}

type MessageStore interface {
	GetMessagesInRoom(roomID int) ([]Message, error)
	CreateMessage(message CreateMessagePayload) error
}

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(User) (int, error)
}

// TODO: move to hub type definition file
// Do we need this? It's empty
type HubStore interface {
}

type ChannelStore interface {
	GetAllChannels() ([]*Channel, error)
	GetChannelsForUser(userID int) ([]*Channel, error)
	CreateChannel(channel CreateChannelPayload) error
	DeleteChannel(channelID int) error
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type CreateMessagePayload struct {
	Content  string `json:"content" validate:"required"`
	RoomID   int    `json:"roomID" validate:"required"`
	SenderID string `json:"senderID" validate:"required"`
}

type ChannelResponse struct {
	Channels []*Channel `json:"channels"`
}

type CreateChannelPayload struct {
	Name string `json:"name" validate:"required"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  int    `json:"userID"`
}
