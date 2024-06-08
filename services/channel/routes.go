package channel

import (
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
	store     types.ChannelStore
	userStore types.UserStore
}

func NewHandler(store types.ChannelStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/channels", auth.WithJWTAuth(
		utils.CorsHandler(h.GetChannelsForUser),
		h.userStore),
	).Methods("GET")

	r.HandleFunc("/channels", utils.CorsHandler(
		auth.WithJWTAuth(h.CreateChannel,
			h.userStore),
	)).Methods("POST", "OPTIONS")
	
	r.HandleFunc("/channels/{channelID}", utils.CorsHandler(
		auth.WithJWTAuth(h.DeleteChannel,
			h.userStore),
	)).Methods("DELETE", "OPTIONS")
}

func (h *Handler) GetChannelsForUser(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	channels, err := h.store.GetChannelsForUser(user.ID)
	if err != nil {
		log.Println("Error getting channels")
		http.Error(w, "Error getting channels", http.StatusInternalServerError)
		return
	}

	response := types.ChannelResponse{Channels: channels}
	utils.SendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())

	channel := &types.Channel{ Rooms: make(map[int]*types.Room, 0) }
	err := utils.ParseJSON(r, channel)
	if err != nil {
		log.Println("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.store.CreateChannel(channel, user)
	if err != nil {
		log.Println("Error: ", err)
		http.Error(w, "Error creating channel", http.StatusInternalServerError)
		return
	}

	hub.HubInstance.AddChannel(channel)
	utils.SendJSONResponse(w, http.StatusCreated, channel)
}

func (h *Handler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channelID"])
	if err != nil {
		log.Println("Invalid channel ID")
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	err = h.store.DeleteChannel(channelID)
	if err != nil {
		log.Println("Error: ", err)
		http.Error(w, "Error deleting channel", http.StatusInternalServerError)
		return
	}

	hub.HubInstance.RemoveChannel(channelID)
	w.WriteHeader(http.StatusOK)
}
