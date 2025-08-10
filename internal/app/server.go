package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pradiptaagus/go-chat-app/internal/router"
	"github.com/pradiptaagus/go-chat-app/pkg/util"
)

func NewServer() {
	router, cancel := router.NewRouter()

	// Ensure the context is cancelled when the server stops
	defer cancel()

	server := http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			util.PanicIfError(err)
		}
	}()

	// Listen for OS signals to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Cancel the context to stop the hub
	cancel()

	// Perform a graceful shutdown of the HTTP server
	ctx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully shut down.")
}
