package room

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
	store     types.RoomStore
	userStore types.UserStore
}

func NewHandler(store types.RoomStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/Channel/{channelID}/rooms", auth.WithJWTAuth(
		utils.CorsHandler(h.GetRoomsInChannel),
		h.userStore),
	).Methods("GET")

	r.HandleFunc("/Room",
		utils.CorsHandler(
			auth.WithJWTAuth(h.CreateRoom,
				h.userStore),
		)).Methods("POST", "OPTIONS")

	r.HandleFunc("/Room/{roomID}",
		utils.CorsHandler(
			auth.WithJWTAuth(h.DeleteRoom,
				h.userStore),
		)).Methods("DELETE", "OPTIONS")
}

func (h *Handler) GetRoomsInChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channelID"])
	if err != nil {
		log.Println("Invalid Channel ID")
		http.Error(w, "Invalid Channel ID", http.StatusBadRequest)
		return
	}

	rooms, err := h.store.GetRoomsInChannel(channelID)
	if err != nil {
		log.Println("Error getting rooms")
		http.Error(w, "Error getting rooms", http.StatusInternalServerError)
		return
	}

	response := types.RoomsResponse{Rooms: rooms}
	utils.SendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	room := &types.Room{}
	err := utils.ParseJSON(r, room)
	if err != nil {
		log.Println("Invalid JSON: ", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	room.Clients = make(map[*types.Client]*types.ClientInfo)
	room.Bus = types.NewEventBus()

	err = h.store.CreateRoom(room)
	if err != nil {
		log.Println("Error: ", err)
		http.Error(w, "Error creating room", http.StatusInternalServerError)
		return
	}

	go room.Run()
	hub.HubInstance.AddRoom(room.ChannelID, room.ID, room)
	utils.SendJSONResponse(w, http.StatusCreated, room)
}

func (h *Handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID, err := strconv.Atoi(vars["roomID"])
	if err != nil {
		log.Println("Invalid room ID")
		http.Error(w, "Invalid room ID", http.StatusBadRequest)
		return
	}

	room, err := h.store.DeleteRoom(roomID)
	if err != nil {
		log.Println("Error: ", err)
		http.Error(w, "Error deleting room", http.StatusInternalServerError)
		return
	}

	hub.HubInstance.RemoveRoom(room.ChannelID, roomID)
	w.WriteHeader(http.StatusOK)
}
