package handlers

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

// Define our WebSocket endpoint
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	HubInstance *Hub
	once        sync.Once
)

type Client struct {
	conn *websocket.Conn
	send chan []byte
	id   string
}

// Hub manages the clients
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

func ChattingHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate the user here
	user, err := AuthenticateRequest(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("Unauthorized:", err)
		return
	}
	// Upgrade this connection to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open WebSocket connection", http.StatusBadRequest)
		log.Println("Could not open WebSocket connection: ", err)
	}

	// Make sure we close the connection when the function returns
	defer ws.Close()

	log.Println("User ID: ", user.ID)

	// Create a new client
	client := NewClient(ws, user.ID)
	HubInstance.register <- client

	go client.ReadMessages()

	client.WriteMessages()
}

// Create a new WebSocket client
func NewClient(conn *websocket.Conn, id string) *Client {
	return &Client{conn, make(chan []byte), id}
}

func HubInitialize() *Hub {
	once.Do(func() {
		HubInstance = &Hub{
			clients:    make(map[*Client]bool),
			broadcast:  make(chan []byte),
			register:   make(chan *Client),
			unregister: make(chan *Client),
		}
		go HubInstance.Run()
	})
	return HubInstance
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (c *Client) ReadMessages() {
	defer func() {
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("Error reading from WebSocket:", err)
			break
		}
		// Process the received message (e.g., save to database)
		HubInstance.broadcast <- message
		// For example, you can pass it to a function to handle saving to PostgreSQL
		SaveMessageToDatabase(message)
	}
}

// WriteMessages writes messages from the send channel to the client connection
func (c *Client) WriteMessages() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				return
			}
			err := c.conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error writing to WebSocket:", err)
				return
			}
		}
	}
}

// SaveMessageToDatabase saves the received message to the PostgreSQL database
func SaveMessageToDatabase(message []byte) {
	// Implement your logic to save the message to the PostgreSQL database
	log.Println("Message received: ", string(message))
}
