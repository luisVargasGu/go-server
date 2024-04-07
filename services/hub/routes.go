package hub

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Channel struct {
	ID    int
	Name  string
	Rooms []*Room
}

type Room struct {
	ID         int
	Name       string
	Clients    map[*Client]bool
	Messages   []*Message
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan []byte
}

type Message struct {
	ID      int
	RoomID  int
	Sender  string
	Content string
	Time    time.Time
}

type Hub struct {
	Channels map[int]*Channel
}

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
	ID   string
}

var (
	HubInstance *Hub
	once        sync.Once
)

// Create a new WebSocket client
func NewClient(conn *websocket.Conn, id string) *Client {
	return &Client{conn, make(chan []byte), id}
}

func HubInitialize() *Hub {
	once.Do(func() {
		HubInstance = &Hub{
			Channels: make(map[int]*Channel),
		}
		go HubInstance.Run()
	})
	return HubInstance
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

// WriteMessages writes messages from the send channel to the client connection
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
