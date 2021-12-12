package main

import (
	"context"
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

	gorillaWebsocket "github.com/gorilla/websocket"
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

	d := gorillaWebsocket.Dialer{}
	c, resp, err := d.Dial(serverURI, requestHeader)
	if err != nil {
		log.Printf("[ERROR] unable to dial: %s", err)
		return
	}

	log.Printf("[DEBUG] response: code=%d", resp.StatusCode)
	defer resp.Body.Close()

	ctx, cancel := context.WithCancel(context.Background())
	breakingConsumerSignal := consumeAsync(ctx, c)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
	case <-breakingConsumerSignal:
	}

	cancel()
	log.Printf("[INFO] shutting down consumer")

	err = closeConnection(c, 3*time.Second)
	if err != nil {
		log.Printf("[ERROR] error when closing connection: %s", err)
	}
}

func closeConnection(c *gorillaWebsocket.Conn, timeout time.Duration) error {
	err := c.WriteControl(
		gorillaWebsocket.CloseMessage,
		gorillaWebsocket.FormatCloseMessage(gorillaWebsocket.CloseNormalClosure, ""),
		time.Now().Add(timeout))
	if err != nil {
		return err
	}

	return c.Close()
}

func consumeAsync(ctx context.Context, c *gorillaWebsocket.Conn) <-chan struct{} {
	quitCh := make(chan struct{}, 1)

	go func() {
	consumerLoop:
		for {
			select {
			case <-ctx.Done():
				log.Printf("[DEBUG] context is canceled when waiting for message")
				break consumerLoop
			default:
				_, msg, err := c.ReadMessage()
				if err != nil {
					opErr, ok := err.(*net.OpError)
					if ok && opErr.Unwrap() == net.ErrClosed {
						log.Printf("[DEBUG] unable to read message due to connection closed")
					} else {
						log.Printf("[ERROR] error while reading message: %s", err)
					}
					// send signal to shutdown consumer immediately
					quitCh <- struct{}{}
					break
				}

				log.Printf("[DEBUG] consumed message: %s", msg)
			}
		}
	}()

	return quitCh
}
