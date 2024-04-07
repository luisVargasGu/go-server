package room

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"user/server/services/utils"
	"user/server/types"
)

type Handler struct {
	store     types.RoomStore
	userStore types.UserStore
}

func NewHandler(store types.RoomStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/channels/{channelID}/rooms", h.GetRoomsInChannel).Methods("GET")
	r.HandleFunc("/rooms", h.CreateRoom).Methods("POST")
	r.HandleFunc("/rooms/{roomID}", h.DeleteRoom).Methods("DELETE")
}

func (h *Handler) GetRoomsInChannel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channelID"])
	if err != nil {
		log.Println("Invalid channel ID")
		http.Error(w, "Invalid channel ID", http.StatusBadRequest)
		return
	}

	rooms, err := h.store.GetRoomsInChannel(channelID)
	if err != nil {
		log.Println("Error getting rooms")
		http.Error(w, "Error getting rooms", http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, rooms)
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	room := types.CreateRoomPayload{}
	err := utils.ParseJSON(r, room)
	if err != nil {
		log.Println("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	err = h.store.CreateRoom(room)
	if err != nil {
		log.Println("Error creating room")
		http.Error(w, "Error creating room", http.StatusInternalServerError)
		return
	}

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

	err = h.store.DeleteRoom(roomID)
	if err != nil {
		log.Println("Error deleting room")
		http.Error(w, "Error deleting room", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
