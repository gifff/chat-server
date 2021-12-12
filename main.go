package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gifff/chat-server/pkg/di"
	"github.com/gifff/chat-server/pkg/server"

	"github.com/labstack/echo"
)

func main() {
	di.InjectDependencies()

	_, cancel := context.WithCancel(context.Background())

	e := echo.New()
	s := server.New(e, ":8080")
	serverCh := s.Start()
	log.Printf("[INFO] Chat Server is started")

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	select {
	case <-quit:
	case <-serverCh:
	}
	log.Printf("[INFO] Shutting down")
	cancel()

	shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCtxCancel()
	if err := e.Shutdown(shutdownCtx); err != nil {
		log.Printf("[ERROR] error when shutting down: %s", err)
	}
}
