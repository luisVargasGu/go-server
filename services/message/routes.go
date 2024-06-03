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
)

type Handler struct {
	store     types.MessageStore
	userStore types.UserStore
}

func NewHandler(store types.MessageStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/channels/{channelID}/room/{roomID}/messages", auth.WithJWTAuth(h.ChattingHandler, h.userStore))
	r.HandleFunc("/rooms/{roomID}/messages", auth.WithJWTAuth(
		utils.CorsHandler(h.FetchMessages),
		h.userStore),
	).Methods("GET")
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

	// Upgrade this connection to a WebSocket
	ws, err := utils.UpgradeToWebSocket(w, r)
	if err != nil {
		http.Error(w, "Could not open WebSocket connection", http.StatusBadRequest)
		log.Println("Could not open WebSocket connection: ", err)
	}

	user := auth.GetUserFromContext(r.Context())

	// Create a new client
	client := hub.NewClient(ws, fmt.Sprint(user.ID), fmt.Sprint(user.Username))
	room.Register <- client

	go client.ReadMessages(room, h.store)

	client.WriteMessages()
}

func (h *Handler) FetchMessages(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := strconv.Atoi(vars["roomID"])
	if err != nil {
		http.Error(w, "Invalid room", http.StatusBadRequest)
		log.Println("FetchMessages: Invalid room", err)
	}

	messages, err := h.store.GetMessagesInRoom(roomID)
	if err != nil {
		http.Error(w, "Unable to fetch messages.", http.StatusInternalServerError)
		log.Println("Unable to fetch messages from db.", err)
	}

	response := types.MessagesResponse{Messages: messages}
	utils.SendJSONResponse(w, http.StatusOK, response)
}

