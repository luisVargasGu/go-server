package hub

import (
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"user/server/types"
)

type Hub struct {
	Channels map[int]*types.Channel
}

var (
	HubInstance *Hub
	once        sync.Once
)

// TODO: move to a client type definition file
func NewClient(conn *websocket.Conn, id string, username string) *types.Client {
	return &types.Client{Conn: conn, Send: make(chan []byte), ID: id, Username: username}
}

type Handler struct {
	store        types.HubStore
	channelStore types.ChannelStore
	roomStore    types.RoomStore
	userStore    types.UserStore
}

func NewHandler(store types.HubStore, channelStore types.ChannelStore, roomStore types.RoomStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, channelStore: channelStore, roomStore: roomStore, userStore: userStore}
}

func (h *Handler) HubInitialize() *Hub {
	once.Do(func() {
		channels, err := h.channelStore.GetAllChannels()
		if err != nil {
			log.Println("Error getting all channels during Hub initialize")
			return
		}
		HubInstance = &Hub{
			Channels: make(map[int]*types.Channel),
		}

		for _, channel := range channels {
			rooms, err := h.roomStore.GetRoomsInChannel(channel.ID)
			if err != nil {
				log.Println("Error getting all channels during Hub initialize")
				return
			}
			channel.Rooms = make(map[int]*types.Room)
			for _, room := range rooms {
				room.Clients = make(map[*types.Client]bool)
				room.Register = make(chan *types.Client)
				room.Unregister = make(chan *types.Client)
				room.Broadcast = make(chan []byte)
				channel.Rooms[room.ID] = room
				go room.Run()
			}
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
