package types

import (
	"encoding/json"
	"log"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	Send     chan []byte
	Username string
	ID       string
}

func (c *Client) ReadMessages(room *Room, store MessageStore) {
	defer func() {
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("Error reading from WebSocket:", err)
			break
		}

		var msg Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("Error unmarshalling JSON message", err)
			return
		}

		store.CreateMessage(msg)
		room.Broadcast <- message
	}
}

func (c *Client) WriteMessages() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				return
			}
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error writing to WebSocket:", err)
				return
			}
		}
	}
}
