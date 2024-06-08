package types

import "sync"


type Hub struct {
	mu       sync.RWMutex
	Channels map[int]*Channel
}

// TODO: Get rid of channel iteration (it's one hub per channel)
func (h *Hub) Run() {
	for _, channel := range h.Channels {
		for _, room := range channel.Rooms {
			for {
				select {
				case client := <-room.Register:
					room.Clients[client] = true
				case client := <-room.Unregister:
					if _, ok := room.Clients[client]; ok {
						delete(room.Clients, client)
						close(client.Send)
					}
				case message := <-room.Broadcast:
					for client := range room.Clients {
						select {
						case client.Send <- message:
						default:
							delete(room.Clients, client)
							close(client.Send)
						}
					}
				}
			}
		}
	}
}

func (h *Hub) GetChannel(channelID int) *Channel {
	h.mu.RLock()
	defer h.mu.RUnlock()
	channel, ok := h.Channels[channelID]
	if !ok {
		return nil
	}
	return channel
}

func (h *Hub) AddChannel(channel *Channel) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Channels[channel.ID] = channel
}

func (h *Hub) RemoveChannel(channelID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.Channels, channelID)
}

func (h *Hub) GetRoom(channelID, roomID int) *Room {
	channel := h.GetChannel(channelID)
	h.mu.RLock()
	defer h.mu.Unlock()
	room, ok := channel.Rooms[roomID]
	if !ok {
		return nil
	}
	return room
}

func (h *Hub) AddRoom(channelID, roomID int, room *Room) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if channel, ok := h.Channels[channelID]; ok {
		channel.Rooms[roomID] = room
	}
}

func (h *Hub) RemoveRoom(channelID, roomID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if channel, ok := h.Channels[channelID]; ok {
		delete(channel.Rooms, roomID)
	}
}
