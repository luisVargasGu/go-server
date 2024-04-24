package types

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

type CreateRoomPayload struct {
	Name      string `json:"name" validate:"required"`
	ChannelID int    `json:"channelID" validate:"required"`
}

type RoomStore interface {
	GetRoomsInChannel(channelID int) ([]*Room, error)
	CreateRoom(room CreateRoomPayload) error
	DeleteRoom(roomID int) error
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Register:
			r.Clients[client] = true

		case client := <-r.Unregister:
			if _, ok := r.Clients[client]; ok {
				delete(r.Clients, client)
				close(client.Send)
			}
		case message := <-r.Broadcast:
			for client := range r.Clients {
				select {
				case client.Send <- message:
				default:
					delete(r.Clients, client)
					close(client.Send)
				}
			}
		}
	}
}
