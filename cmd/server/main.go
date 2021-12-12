package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gifff/chat-server/deps"
	"github.com/gifff/chat-server/logger"
	"github.com/gifff/chat-server/server"
	"github.com/gifff/chat-server/server/handlers"

	"github.com/labstack/echo"
)

var (
	logLevel     string
	port         int
	connReporter bool
)

func main() {
	flag.StringVar(&logLevel, "log-level", "INFO", "log level. Available options: DEBUG, INFO, WARN, DEBUG")
	flag.IntVar(&port, "port", 8080, "server port")
	flag.BoolVar(&connReporter, "reporter-enabled", false, "enable total connections reporter that ticks every second")
	flag.Parse()

	log.SetOutput(logger.NewLevelFilter(logLevel, os.Stdout))
	serverPort := fmt.Sprintf(":%d", port)

	wgw, chatSvc := deps.BuildDependencies()
	hs := handlers.Handlers{
		WebsocketGateway: wgw,
		ChatService:      chatSvc,
	}

	_, cancel := context.WithCancel(context.Background())

	e := echo.New()
	s := server.New(e, serverPort, hs)
	serverCh := s.Start()
	log.Printf("[INFO] Chat Server is started at port %d", port)

	if connReporter {
		go func() {
			for {
				<-time.After(1 * time.Second)
				log.Printf("[INFO] Number of connections: %d", wgw.TotalConnections())
			}
		}()
	}

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

	s.Stop(10 * time.Second)
}
