package router

import (
	"context"
	"net/http"

	"github.com/pradiptaagus/go-chat-app/internal/handler"
	"github.com/pradiptaagus/go-chat-app/internal/service"
)

func NewRouter() (*http.ServeMux, context.CancelFunc) {
	router := http.NewServeMux()

	// Create new Hub and invoke Run function immediately
	hub := service.NewHubImpl()
	ctx, cancel := context.WithCancel(context.Background())
	go hub.Run(ctx)

	router.HandleFunc("/", handler.ClientUI)
	router.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		handler.WsHandler(ctx, w, r, hub)
	})

	return router, cancel
}
