package server

import (
	"github.com/gifff/chat-server/pkg/server/handlers"
	"github.com/gifff/chat-server/pkg/server/middlewares"

	"github.com/labstack/echo"
)

// New instantiates Server instance
func New(e *echo.Echo, port string) Server {
	if port == "" {
		port = ":8080"
	}

	e.Use(middlewares.AuthenticationExtractor)
	e.GET("/messages/listen", handlers.Hello)

	return Server{
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
func (s Server) Start() {
	s.e.Logger.Fatal(s.e.Start(s.port))
}
