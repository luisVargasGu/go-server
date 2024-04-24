package types

import (
	"log"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	Send     chan []byte
	Username string
	ID       string
}

func (c *Client) ReadMessages(room *Room) {
	defer func() {
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("Error reading from WebSocket:", err)
			break
		}
		// Process the received message (e.g., save to database)
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
