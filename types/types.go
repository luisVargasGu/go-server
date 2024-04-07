package types

import (
	"github.com/gorilla/websocket"
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
	Rooms []*Room
}

type Room struct {
	ID         int
	Name       string
	ChannelID  int
	Clients    map[*Client]bool
	Messages   []*Message
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

type Message struct {
	ID        int
	RoomID    int
	SenderID  string
	Content   string
	Timestamp time.Time
	IsRead    bool
}

type Client struct {
	Conn     *websocket.Conn
	Send     chan []byte
	Username string
	ID       string
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

type RoomStore interface {
	GetRoomsInChannel(channelID int) ([]*Room, error)
	CreateRoom(room CreateRoomPayload) error
	DeleteRoom(roomID int) error
}

type ChannelStore interface {
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

type CreateRoomPayload struct {
	Name      string `json:"name" validate:"required"`
	ChannelID int    `json:"channelID" validate:"required"`
}

type CreateMessagePayload struct {
	Content  string `json:"content" validate:"required"`
	RoomID   int    `json:"roomID" validate:"required"`
	SenderID string `json:"senderID" validate:"required"`
}

type CreateChannelPayload struct {
	Name string `json:"name" validate:"required"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  int    `json:"userID"`
}
