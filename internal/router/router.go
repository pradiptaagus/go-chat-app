package router

import (
	"context"
	"net/http"

	"github.com/pradiptaagus/go-chat-app/internal/handler"
	"github.com/pradiptaagus/go-chat-app/internal/service"
)

func NewRouter() (*http.ServeMux, context.CancelFunc) {
	router := http.NewServeMux()

	chatServer := service.NewChatServerImpl()

	ctx, cancel := context.WithCancel(context.Background())

	router.HandleFunc("/", handler.ClientUI)
	router.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		handler.WsHandler(ctx, w, r, chatServer)
	})

	return router, cancel
}
