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
