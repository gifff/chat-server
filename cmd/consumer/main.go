package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gifff/chat-server/logger"
	"github.com/gifff/chat-server/wsclient"
)

var (
	logLevel  string
	serverURI string
	userID    int
)

func main() {
	flag.StringVar(&logLevel, "log-level", "INFO", "log level. Available options: DEBUG, INFO, WARN, DEBUG")
	flag.StringVar(&serverURI, "server-uri", "", "websocket server URI. i.e: ws://localhost:8080/messages/listen")
	flag.IntVar(&userID, "user-id", 1, "user id to be embed in X-User-Id header")
	flag.Parse()

	log.SetOutput(logger.NewLevelFilter(logLevel, os.Stdout))

	requestHeader := http.Header{}
	requestHeader.Set("X-User-Id", strconv.Itoa(userID))

	wsClient := wsclient.NewClient(serverURI, requestHeader)
	wsClient.AttachObserver(consumer)
	errChan, err := wsClient.Listen()
	if err != nil {
		log.Printf("[ERROR] unable to listen: %s", err)
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
	case err := <-errChan:
		opErr, ok := err.(*net.OpError)
		if ok && opErr.Unwrap() == net.ErrClosed {
			log.Printf("[DEBUG] unable to read message due to connection closed")
		} else {
			log.Printf("[ERROR] error while reading message: %s", err)
		}
	}

	log.Printf("[INFO] shutting down consumer")

	err = wsClient.Stop(3 * time.Second)
	if err != nil {
		log.Printf("[ERROR] error when closing connection: %s", err)
	}
}

func consumer(msgType int, msg []byte) {
	log.Printf("[DEBUG] consumed message: %s", msg)
}
