package message

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"user/server/services/auth"
	"user/server/services/hub"
	"user/server/services/utils"
	"user/server/types"
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
	r.HandleFunc("/messages/{messageID}/seen", utils.CorsHandler(
		auth.WithJWTAuth(
			h.MarkMessageAsSeenHandler,
			h.userStore,
		),
	)).Methods("PUT", "OPTIONS")
}

func (h *Handler) ChattingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	channelID, err := strconv.Atoi(vars["channelID"])
	if err != nil {
		http.Error(w, "Invalid channel", http.StatusBadRequest)
		log.Println("Invalid channel:", err)
		return
	}

	roomID, err := strconv.Atoi(vars["roomID"])
	if err != nil {
		http.Error(w, "Invalid room", http.StatusBadRequest)
		log.Println("Invalid room:", err)
		return
	}

	channel := hub.HubInstance.GetChannel(channelID)
	if channel == nil {
		http.Error(w, "Channel not found", http.StatusNotFound)
		log.Println("Channel not found")
		return
	}

	room := channel.GetRoom(roomID)
	if room == nil {
		http.Error(w, "Room not found", http.StatusNotFound)
		log.Println("Room not found")
		return
	}

	ws, err := utils.UpgradeToWebSocket(w, r)
	if err != nil {
		http.Error(w, "Could not open WebSocket connection", http.StatusBadRequest)
		log.Println("Could not open WebSocket connection: ", err)
		return
	}

	user := auth.GetUserFromContext(r.Context())

	client := hub.NewClient(ws, user)
	if room.Bus == nil {
		http.Error(w, "Could not connect to the room", http.StatusBadRequest)
		log.Println("Could not connect to the room: ", err)
		return
	}

	room.Bus.Publish(types.Event{
		Type:    types.EventRegister,
		Payload: client,
	})

	go client.ReadMessages(room, h.store)

	go client.WriteMessages()
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

func (h *Handler) MarkMessageAsSeenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := auth.GetUserFromContext(r.Context())
	messageID, err := strconv.Atoi(vars["messageID"])
	if err != nil {
		log.Println("Invalid message ID", err)
		http.Error(w, "Invalid message ID", http.StatusBadRequest)
		return
	}

	err = h.store.MarkMessageAsSeen(user.ID, messageID)
	if err != nil {
		log.Println("Error marking message as seen:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
