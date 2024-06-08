package types

type Room struct {
    ID         int            `json:"id"`
    Name       string         `json:"name"`
    ChannelID  int            `json:"channel_id"`
    Clients    map[*Client]bool `json:"-"`
    Messages   []*Message     `json:"-"`
    Register   chan *Client   `json:"-"`
    Unregister chan *Client   `json:"-"`
    Broadcast  chan []byte    `json:"-"`
}

type RoomResponse struct {
	Rooms []*Room `json:"rooms"`
}

type RoomStore interface {
	GetRoomsInChannel(channelID int) ([]*Room, error)
	CreateRoom(room *Room) error
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
