package types

import "sync"

type Channel struct {
	mu    sync.RWMutex
	ID    int           `json:"id"`
	Name  string        `json:"name"`
	Rooms map[int]*Room `json:"-"`
	Avatar    []byte    `json:"avatar"`
}

type ChannelStore interface {
	GetAllChannels() ([]*Channel, error)
	GetChannelsForUser(userID int) ([]*Channel, error)
	CreateChannel(channel *Channel, user *User) error
	DeleteChannel(channelID int) error
}

type ChannelResponse struct {
	Channels []*Channel `json:"channels"`
}

func (c *Channel) GetRoom(roomID int) *Room {
	c.mu.RLock()
	defer c.mu.RUnlock()
	room, ok := c.Rooms[roomID]
	if !ok {
		return nil
	}
	return room
}
