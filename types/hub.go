package types

import "sync"

type Hub struct {
	mu       sync.RWMutex
	Channels map[int]*Channel
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

// TODO: I have a suspicion I need to mutex the channels and rooms too
func (h *Hub) GetRoom(channelID, roomID int) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	if channel, ok := h.Channels[channelID]; ok {
		if room, ok := channel.Rooms[roomID]; ok {
			return room
		}
	}
	return nil
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
