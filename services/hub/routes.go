package hub

import (
	"fmt"
	"log"
	"sync"
	"user/server/types"

	"github.com/gorilla/websocket"
)

var (
	HubInstance *types.Hub
	once        sync.Once
)

// TODO: move to a client type definition file
func NewClient(conn *websocket.Conn, user *types.User) *types.Client {
	return &types.Client{
		WebsocketConnection: conn,
		Send:                make(chan []byte, 20),
		ID:                  fmt.Sprint(user.ID),
		Username:            user.Username,
		Avatar:              user.Avatar,
		MicEnabled:          false,
		VideoEnabled:        false,
		ScreenEnabled:       false}
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

func (h *Handler) HubInitialize() *types.Hub {
	once.Do(func() {
		channels, err := h.channelStore.GetAllChannels()
		if err != nil {
			log.Println("Error getting all channels during Hub initialize")
			return
		}
		HubInstance = &types.Hub{
			Channels: make(map[int]*types.Channel),
		}

		for _, channel := range channels {
			HubInstance.Channels[channel.ID] = channel
			rooms, err := h.roomStore.GetRoomsInChannel(channel.ID)
			if err != nil {
				log.Println("Error getting all channels during Hub initialize")
				return
			}
			channel.Rooms = make(map[int]*types.Room)
			for _, room := range rooms {
				room.Clients = make(map[*types.Client]*types.ClientInfo)
				room.Bus = types.NewEventBus()
				channel.Rooms[room.ID] = room
				go room.Run()
			}
		}
	})
	return HubInstance
}
