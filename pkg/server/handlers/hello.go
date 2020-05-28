package handlers

import (
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"

	"github.com/gifff/chat-server/pkg/model"
)

var (
	upgrader       = websocket.Upgrader{}
	connectionPool = make(map[int]*Connection)
	mutex          sync.RWMutex
	id             int64 = 0
)

// Connection wraps the websocket.Conn with mutex
type Connection struct {
	mu   sync.Mutex
	conn *websocket.Conn
}

func broadcast(msg string, senderID int) {
	atomic.AddInt64(&id, 1)

	mutex.RLock()
	for userID, wsConn := range connectionPool {

		c := wsConn
		userID := userID

		message := model.Message{
			ID:      int(id),
			Type:    model.TextMessage,
			Message: msg,
			User: model.User{
				ID:   senderID,
				IsMe: userID == senderID,
			},
		}

		go func() {
			c.mu.Lock()
			log.Printf("Writing to [User ID: %d] at %d", userID, time.Now().UnixNano())
			c.conn.SetWriteDeadline(time.Now().Add(time.Second))
			_ = c.conn.WriteJSON(message)
			c.mu.Unlock()
		}()

	}
	mutex.RUnlock()
}

// Hello is a websocket handler
func Hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	conn := &Connection{
		conn: ws,
	}

	userID, _ := c.Get("user_id").(int)
	mutex.Lock()
	connectionPool[userID] = conn
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
