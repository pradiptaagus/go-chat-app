package app

import (
	"net/http"

	"github.com/pradiptaagus/go-chat-app/internal/router"
	"github.com/pradiptaagus/go-chat-app/pkg/util"
)

func NewServer() {
	router := router.NewRouter()

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	err := server.ListenAndServe()
	util.PanicIfError(err)
}
