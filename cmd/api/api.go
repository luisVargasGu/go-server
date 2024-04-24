package api

import (
	"database/sql"
	"log"
	"net/http"
	"user/server/services/channel"
	"user/server/services/hub"
	"user/server/services/image"
	"user/server/services/message"
	"user/server/services/room"
	"user/server/services/user"

	"github.com/gorilla/mux"
)

type APIServer struct {
	addr string
	db   *sql.DB
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db:   db,
	}
}

func (s *APIServer) Run() error {
	router := mux.NewRouter()
	subrouter := router.PathPrefix("/api/v1").Subrouter()

	userStore := user.NewStore(s.db)
	userHandler := user.NewHandler(userStore)
	userHandler.RegisterRoutes(subrouter)

	channelStore := channel.NewStore(s.db)
	channelHandler := channel.NewHandler(channelStore, userStore)
	channelHandler.RegisterRoutes(subrouter)

	roomStore := room.NewStore(s.db)
	roomHandler := room.NewHandler(roomStore, userStore)
	roomHandler.RegisterRoutes(subrouter)

	messageStore := message.NewStore(s.db)
	messageHandler := message.NewHandler(messageStore, userStore)
	messageHandler.RegisterRoutes(subrouter)

	hubStore := hub.NewStore(s.db)
	hubHandler := hub.NewHandler(hubStore, channelStore, roomStore, userStore)
	hubHandler.HubInitialize()

	subrouter.HandleFunc("/image", image.ImageHandler)

	log.Println("Starting server on", s.addr)
	return http.ListenAndServe(s.addr, router)
}
