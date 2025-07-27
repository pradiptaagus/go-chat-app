package router

import (
	"net/http"

	"github.com/pradiptaagus/go-chat-app/internal/handler"
	"github.com/pradiptaagus/go-chat-app/internal/service"
)

func NewRouter() *http.ServeMux {
	router := http.NewServeMux()

	// Create new Hub and invoke Run function immediately
	hub := service.NewHubImpl()
	go hub.Run()

	router.HandleFunc("/", handler.ClientUI)
	router.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		handler.WsHandler(w, r, hub)
	})

	return router
}
