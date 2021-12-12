package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/gifff/chat-server/logger"
	"github.com/gifff/chat-server/wsclient"
)

var (
	logLevel      string
	serverURI     string
	userID        int
	numOfConsumer int
)

func main() {
	flag.StringVar(&logLevel, "log-level", "INFO", "log level. Available options: DEBUG, INFO, WARN, DEBUG")
	flag.StringVar(&serverURI, "server-uri", "", "websocket server URI. i.e: ws://localhost:8080/messages/listen")
	flag.IntVar(&userID, "user-id", 1, "user id to be embed in X-User-Id header")
	flag.IntVar(&numOfConsumer, "n", 1, "number of consumers to be spawned")
	flag.Parse()

	log.SetOutput(logger.NewLevelFilter(logLevel, os.Stdout))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	spawnConsumer(numOfConsumer, quit)
}

func spawnConsumer(n int, quit <-chan os.Signal) {
	if n < 1 {
		log.Printf("[ERROR] number of consumer must be positive. specified=%d", n)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(n)
	stopChan := make(chan struct{}, n)

	go func() {
		quitSig := <-quit
		log.Printf("[DEBUG] quit signal acknowledged: signal=%s", quitSig)
		log.Printf("[INFO] shutting down consumers")
		for i := 0; i < n; i++ {
			stopChan <- struct{}{}
		}
	}()

	requestHeader := http.Header{}
	requestHeader.Set("X-User-Id", strconv.Itoa(userID))

	for i := 0; i < n; i++ {
		wsClient := wsclient.NewClient(serverURI, requestHeader)
		wsClient.AttachObserver(createMessageConsumer(i))
		errChan, err := wsClient.Listen()
		if err != nil {
			log.Printf("[ERROR][Consumer: %d] unable to listen: %s", i, err)
			wg.Done()
			continue
		}

		go func(wsClient *wsclient.Client, wg *sync.WaitGroup, consumerNum int) {
			select {
			case <-stopChan:
			case err := <-errChan:
				if err != nil {
					opErr, ok := err.(*net.OpError)
					if ok && opErr.Unwrap() == net.ErrClosed {
						log.Printf("[DEBUG][Consumer: %d] unable to read message due to connection closed", consumerNum)
					} else {
						log.Printf("[ERROR][Consumer: %d] error while reading message: %s", consumerNum, err)
					}
				}
			}

			err := wsClient.Stop(3 * time.Second)
			if err != nil {
				log.Printf("[ERROR][Consumer: %d] error when closing connection: %s", consumerNum, err)
			}
			log.Printf("[DEBUG][Consumer: %d] consumer has been shut down", consumerNum)

			wg.Done()
		}(wsClient, &wg, i)

	}

	wg.Wait()
	log.Printf("[DEBUG] all consumers are shut down")
}

func createMessageConsumer(consumerNum int) wsclient.ObserverFunc {
	return func(msgType int, msg []byte) {
		log.Printf("[DEBUG][Consumer: %d] consumed message: %s", consumerNum, msg)
	}
}
