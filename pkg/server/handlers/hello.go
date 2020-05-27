package handlers

import (
	"log"
	"sync"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"

	"github.com/gifff/chat-server/pkg/model"
)

var (
	upgrader       = websocket.Upgrader{}
	connectionPool = make(map[int]*websocket.Conn)
	mutex          sync.RWMutex
	id             int64 = 1
)

func broadcast(msg string, senderID int) {
	userData := map[string]interface{}{
		"id":    senderID,
		"is_me": false,
	}

	message := map[string]interface{}{
		"id":      id,
		"message": msg,
		"type":    model.TextMessage,
		"user":    userData,
	}
	atomic.AddInt64(&id, 1)

	mutex.RLock()
	for userID, wsConn := range connectionPool {

		if userID == senderID {
			userData["is_me"] = true
		} else {
			userData["is_me"] = false
		}

		_ = wsConn.WriteJSON(message)

	}
	mutex.RUnlock()
}

// Hello is a websocket handler
func Hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	userID, _ := c.Get("user_id").(int)
	mutex.Lock()
	connectionPool[userID] = ws
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		delete(connectionPool, userID)
		mutex.Unlock()
		ws.Close()
	}()

	for {
		var msg model.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error while reading message [userID: %d] : %v\n", userID, err)
			break
		}

		log.Printf("Message [userID: %d]: %+v\n", userID, msg)

		if msg.Type != model.UnknownMessage {
			broadcast(msg.Message, userID)
		}
	}

	return nil
}
