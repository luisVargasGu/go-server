package invite

import (
	"log"
	"net/http"
	"strconv"
	"time"
	"fmt"
	"user/server/services/auth"
	"user/server/services/utils"
	"user/server/types"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var BASE_URL string = "https://backendserver.me/api/v1"

type Handler struct {
	store           types.InviteStore
	userStore       types.UserStore
	permissionStore types.PermissionsStore
}

func NewHandler(
	store types.InviteStore,
	userStore types.UserStore,
	permissionStore types.PermissionsStore) *Handler {
	return &Handler{store: store, userStore: userStore, permissionStore: permissionStore}
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

	err = h.store.SaveInvite(&invite)
	if err != nil {
		http.Error(w, "Failed to save Invite.", http.StatusInternalServerError)
		log.Println("Failed to save Invite.", err)
		return
	}

	inviteLink := BASE_URL + "/invite/" + inviteCode
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

	invite, err := h.store.FindInvite(inviteLink)
	if err != nil {
		http.Error(w, "Failed to find Invite.", http.StatusNotFound)
		log.Println("Failed to find Invite.", err)
		return
	}

	if h.permissionStore.UserHasPermission(user.ID, invite.ChanelID) {
		http.Error(w, "User already in Channel.", http.StatusConflict)
		log.Println("User already in Channel.")
		return
	}

	if !isInviteValid(invite) {
		http.Error(w, "Invite is Invalid.", http.StatusGone)
		log.Println("Invite is Invalid.")
		return
	}

	err = h.store.AcceptInvite(invite.ID, user.ID)
	if err != nil {
		http.Error(w, "Failed to accept invite.", http.StatusInternalServerError)
		log.Println("Failed to accept invite.", err)
		return
	}

	response := types.InviteAcceptedResponse{Status: "success", Message: "Invite accepted successfully", Inv: *invite}
	utils.SendJSONResponse(w, http.StatusOK, response)
}

func isInviteValid(invite *types.Invite) bool {
	if invite.InviteeID != -1 {
		return false
	}

	if time.Now().After(invite.Expiration) {
		return false
	}

	return true
}
