package router

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pradiptaagus/go-chat-app/internal/controller"
	"github.com/pradiptaagus/go-chat-app/internal/service"
)

func NewRouter() (*mux.Router, context.CancelFunc) {
	router := mux.NewRouter()
	subRouter := router.PathPrefix("/api/v1/").Subrouter()

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize services
	chatServerService := service.NewChatServerService()
	websocketService := service.NewWebSocketService()

	// Initialize controllers
	clientController := controller.NewClientController()
	websocketController := controller.NewWebsocketController()

	// UI routes
	router.HandleFunc("/", controller.ClientUI).Methods("GET")

	// WebSocket route
	router.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		websocketController.WsHandler(ctx, w, r, chatServerService, websocketService)
	})

	// Clienr API routes
	subRouter.HandleFunc("/clients", func(w http.ResponseWriter, r *http.Request) {
		clientController.Create(ctx, w, r)
	}).Methods("POST")
	subRouter.HandleFunc("/clients", func(w http.ResponseWriter, r *http.Request) {
		clientController.FindAll(ctx, w, r)
	}).Methods("GET")

	return router, cancel
}
