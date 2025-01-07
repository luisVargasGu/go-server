package types

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	Avatar    []byte    `json:"avatar"`
}

type UserInfo struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	Avatar          string      `json:"avatar,omitempty"`
	IsMicEnabled    bool        `json:"isMicEnabled"`
	IsVideoEnabled  bool        `json:"isVideoEnabled"`
	IsScreenEnabled bool        `json:"isScreenEnabled"`
	Tracks          []TrackInfo `json:"tracks"`
}

type SeenByUser struct {
	Avatar   string `json:"avatar"`
	Username string `json:"username"`
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

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Avatar   string `json:"avatar"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
	Avatar  string `json:"avatar"`
}
