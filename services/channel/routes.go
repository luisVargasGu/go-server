package channel

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"user/server/services/auth"
	"user/server/services/hub"
	"user/server/services/image"
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

	err := r.ParseMultipartForm(image.MAX_UPLOAD_SIZE)
	if err != nil {
		log.Println("Error parsing multipart form: ", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Error retrieving file: ", err)
		http.Error(w, "Unable to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(fileType, "image/") {
		log.Println("Invalid file type: ", fileType)
		http.Error(w, "The uploaded file is not an image", http.StatusBadRequest)
		return
	}

	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Println("Error reading file data: ", err)
		http.Error(w, "Error reading file data", http.StatusInternalServerError)
		return
	}

	data, err := image.ResizeImage(fileData, image.MAX_WIDTH, image.MAX_HEIGHT)
	if err != nil {
		log.Println("Error getting file data: ", err)
		http.Error(w, "Error getting file data", http.StatusInternalServerError)
		return
	}
	channelName := r.FormValue("name")

	channel := &types.Channel{
		Name:  channelName,
		Rooms: make(map[int]*types.Room, 0),
		Avatar: data,
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
