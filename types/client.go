package types

import (
	"log"
	"user/server/services/utils"
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
		room.Unregister <- c
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseNoStatusReceived,
			) {
				log.Printf("WebSocket closed: %v", err)
			} else {
				log.Printf("Error reading from WebSocket: %v", err)
			}
			return
		}

		var msg Message
		err = utils.Unmarshal(message, &msg)
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
