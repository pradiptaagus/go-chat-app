package router

import (
	"net/http"

	"github.com/pradiptaagus/go-chat-app/internal/handler"
)

func NewRouter() *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("/", handler.ClientUI)
	router.HandleFunc("/chat", handler.WsHandler)

	return router
}
