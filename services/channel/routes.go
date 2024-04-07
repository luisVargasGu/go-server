package channel

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"user/server/services/utils"
	"user/server/types"
)

type Handler struct {
	store     types.ChannelStore
	userStore types.UserStore
}

func NewHandler(store types.ChannelStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/channels", h.GetChannelsForUser).Methods("GET")
	r.HandleFunc("/channels", h.CreateChannel).Methods("POST")
	r.HandleFunc("/channels/{channelID}", h.DeleteChannel).Methods("DELETE")
}

func (h *Handler) GetChannelsForUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channelID"])
	if err != nil {
		log.Println("Invalid channel ID")
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	channels, err := h.store.GetChannelsForUser(channelID)
	if err != nil {
		log.Println("Error getting channels")
		http.Error(w, "Error getting channels", http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, channels)
}

func (h *Handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	channel := types.CreateChannelPayload{}
	err := utils.ParseJSON(r, channel)
	if err != nil {
		log.Println("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.store.CreateChannel(channel)
	if err != nil {
		log.Println("Error creating channel")
		http.Error(w, "Error creating channel", http.StatusInternalServerError)
		return
	}

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
		log.Println("Error deleting channel")
		http.Error(w, "Error deleting channel", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

