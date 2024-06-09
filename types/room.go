package types

import "sync"

type Room struct {
	mu         sync.RWMutex
	ID         int              `json:"id"`
	Name       string           `json:"name"`
	ChannelID  int              `json:"channel_id"`
	Clients    map[*Client]bool `json:"-"`
	Messages   []*Message       `json:"-"`
	Register   chan *Client     `json:"-"`
	Unregister chan *Client     `json:"-"`
	Broadcast  chan []byte      `json:"-"`
}

type RoomResponse struct {
	Rooms []*Room `json:"rooms"`
}

type RoomStore interface {
	GetRoomsInChannel(channelID int) ([]*Room, error)
	CreateRoom(room *Room) error
	DeleteRoom(roomID int) (*Room, error)
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.mu.Lock()
			r.Clients[client] = true
			r.mu.Unlock()

		case client := <-r.Unregister:
			r.mu.Lock()
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send)
			}
			r.mu.Unlock()
		case message := <-r.Broadcast:
			r.mu.RLock()
			for client := range r.Clients {
				select {
				case client.Send <- message:
				default:
					r.mu.RUnlock()
					r.mu.Lock()
					delete(r.Clients, client)
					close(client.Send)
					r.mu.Unlock()
					r.mu.RLock()
				}
			}
			r.mu.RUnlock()
		}
	}
}
