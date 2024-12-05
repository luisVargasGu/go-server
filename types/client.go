package types

import (
	"log"
	"sync"
	"user/server/services/utils"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type TrackInfo struct {
	Track *webrtc.TrackLocalStaticRTP `json:"-"`
	ID    string                      `json:"id"`
	Kind  string                      `json:"type"`
}

type ClientInfo struct {
	Connected   bool
	MediaTracks map[string]*TrackInfo
}

type Client struct {
	mu                  sync.RWMutex
	WebsocketConnection *websocket.Conn
	PeerConnection      *webrtc.PeerConnection
	Send                chan []byte
	Username            string
	ID                  string
	Avatar              []byte
	MicEnabled          bool
	VideoEnabled        bool
	ScreenEnabled       bool
}

func (c *Client) ReadMessages(room *Room, store MessageStore) {
	defer func() {
		c.WebsocketConnection.Close()
		room.Bus.Publish(Event{
			Type:    EventUnregister,
			Payload: c,
		})
	}()
	for {
		_, message, err := c.WebsocketConnection.ReadMessage()
		if err != nil {
			log.Printf("Error reading WebSocket message: %v", err)
			return
		}
		c.handleMessage(message, room, store)
	}
}

func (c *Client) handleMessage(message []byte, room *Room, store MessageStore) {
	var msg Message
	err := utils.Unmarshal(message, &msg)
	if err != nil {
		log.Println("Error unmarshalling JSON message", err)
		return
	}
	if msg.Type == "chat-message" {
		if err := store.CreateMessage(&msg); err != nil {
			log.Println("Error creating message:", err)
			return
		}

		broadcastMessage := utils.Marshal(msg)
		if broadcastMessage == nil {
			log.Println("Error marshalling message for broadcast:", err)
			return
		}

		room.Bus.Publish(Event{
			Type:    EventBroadcast,
			Payload: broadcastMessage,
		})
	} else {
		room.Bus.Publish(Event{
			Type:    EventBroadcast,
			Payload: message,
		})
	}
}

func (c *Client) WriteMessages() {
	defer func() {
		c.WebsocketConnection.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				log.Println("Error writing to nil Socket")
				return
			}
			err := c.WebsocketConnection.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error writing to WebSocket:", err)
				return
			}
		}
	}
}

func (r *Room) GetClientByID(clientID string) *Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for client := range r.Clients {
		if client.ID == clientID {
			return client
		}
	}
	return nil
}
