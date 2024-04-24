package message

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"user/server/services/auth"
	"user/server/services/hub"
	"user/server/services/utils"
	"user/server/types"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Handler struct {
	store     types.MessageStore
	userStore types.UserStore
}

func NewHandler(store types.MessageStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/channels/{channelID}/room/{roomID}/messages", auth.WithJWTAuth(h.ChatAndCreateMessageHandler, h.userStore))
}

func (h *Handler) ChatAndCreateMessageHandler(w http.ResponseWriter, r *http.Request) {
	if websocket.IsWebSocketUpgrade(r) {
		h.ChattingHandler(w, r)
		return
	}

	h.CreateMessage(w, r)
}

func (h *Handler) ChattingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("ChattingHandler")
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channelID"])
	roomID, err := strconv.Atoi(vars["roomID"])
	if err != nil {
		http.Error(w, "Invalid channel or room", http.StatusBadRequest)
		log.Println("Invalid channel or room:", err)
		return
	}

	channel := hub.HubInstance.Channels[channelID]
	room := channel.Rooms[roomID]

	log.Println("room: ", room)
	log.Println("channel: ", channel)
	// Upgrade this connection to a WebSocket
	ws, err := utils.UpgradeToWebSocket(w, r)
	if err != nil {
		http.Error(w, "Could not open WebSocket connection", http.StatusBadRequest)
		log.Println("Could not open WebSocket connection: ", err)
	}

	user := r.Context().Value(auth.UserKey).(*types.User)

	// Create a new client
	client := hub.NewClient(ws, fmt.Sprint(user.ID), fmt.Sprint(user.Username))
	room.Register <- client

	go client.ReadMessages(room)

	client.WriteMessages()
}

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channelID"])
	roomID, err := strconv.Atoi(vars["roomID"])
	if err != nil {
		http.Error(w, "Invalid channel or room", http.StatusBadRequest)
		log.Println("Invalid channel or room:", err)
		return
	}

	channel := hub.HubInstance.Channels[channelID]
	room := channel.Rooms[roomID]

	user := r.Context().Value(auth.UserKey).(*types.User)

	var message types.Message
	err = utils.ParseJSON(r, &message)
	if err != nil {
		http.Error(w, "Invalid message", http.StatusBadRequest)
		log.Println("Invalid message:", err)
		return
	}

	message.SenderID = user.Username
	message.RoomID = room.ID
	room.Broadcast <- []byte(message.Content)
}
