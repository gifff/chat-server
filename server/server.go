package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gifff/chat-server/server/handlers"
	"github.com/gifff/chat-server/server/middlewares"

	"github.com/labstack/echo"
)

// New instantiates Server instance
func New(e *echo.Echo, port string, h handlers.Handlers) *Server {
	if port == "" {
		port = ":8080"
	}

	e.Use(middlewares.AuthenticationExtractor)
	e.GET("/messages/listen", h.MessageListener)
	e.POST("/messages", h.SendMessage)

	return &Server{
		e:    e,
		port: port,
	}
}

// Server wraps information about Echo instance and the used port for the webserver
type Server struct {
	e    *echo.Echo
	port string
}

// Start fires up the Echo server
func (s *Server) Start() <-chan struct{} {
	ch := make(chan struct{}, 1)
	go func() {
		if err := s.e.Start(s.port); err != nil && err != http.ErrServerClosed {
			log.Printf("[ERROR] got error when starting server: %s", err)
			ch <- struct{}{}
		}

		close(ch)
	}()

	return ch
}

// Stop signals the Server to stop
func (s *Server) Stop(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := s.e.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] error when shutting down: %s", err)
	}
}
