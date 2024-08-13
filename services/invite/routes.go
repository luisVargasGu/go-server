package invite

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"user/server/services/auth"
	"user/server/services/utils"
	"user/server/types"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type Handler struct {
	store           types.InviteStore
	userStore       types.UserStore
	permissionStore types.PermissionsStore
}

func NewHandler(store types.InviteStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/invite/create/{channelID}", auth.WithJWTAuth(
		utils.CorsHandler(h.CreateInviteHandler),
		h.userStore),
	).Methods("GET")
	r.HandleFunc("/invite/{inviteCode}", auth.WithJWTAuth(
		utils.CorsHandler(h.AcceptInviteHandler),
		h.userStore),
	).Methods("GET")
}

func (h *Handler) CreateInviteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	channelID, err := strconv.Atoi(vars["channelID"])
	if err != nil {
		http.Error(w, "Invalid channel", http.StatusBadRequest)
		log.Println("Invalid channel:", err)
		return
	}
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("Unauthorized:", err)
		return
	}

	if !h.permissionStore.UserHasPermission(user.ID, channelID) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		log.Println("Forbidden:", err)
		return
	}

	inviteCode := uuid.NewString()
	invite := types.Invite{
		ChanelID:   channelID,
		InviterID:  user.ID,
		InviteCode: inviteCode,
		Expiration: time.Now().Add(1 * time.Hour),
	}

	err = h.store.SaveInvite(invite)
	if err != nil {
		http.Error(w, "Failed to save Invite.", http.StatusInternalServerError)
		log.Println("Failed to save Invite.", err)
		return
	}

	// TODO: use a global constant for baseUR:
	inviteLink := "http://35.183.253.88:8080/invite/" + inviteCode
	response := types.InviteRespose{Link: inviteLink}
	utils.SendJSONResponse(w, http.StatusOK, response)
}

func (h *Handler) AcceptInviteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	inviteLink := vars["inviteCode"]

	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Println("Unauthorized")
		return
	}

	invite, err := h.store.AcceptInvite(inviteLink, user.ID)
	if err != nil {
		http.Error(w, "Failed to find Invite.", http.StatusInternalServerError)
		log.Println("Failed to find Invite.", err)
		return
	}

	if h.permissionStore.UserHasPermission(user.ID, invite.ChanelID) {
		http.Error(w, "User already in Channel.", http.StatusBadRequest)
		log.Println("User already in Channel.")
	}
}
