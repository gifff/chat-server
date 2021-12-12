package gateway

import (
	"log"
	"sync"
	"time"

	"github.com/gifff/chat-server/contract"
	"github.com/gifff/chat-server/pkg/model"
	"github.com/gifff/chat-server/pkg/websocket"
)

// NewWebsocket returns Websocket object which satisfies the WebsocketGateway contract
func NewWebsocket() contract.WebsocketGateway {
	return &Websocket{
		userConnectionPoolMap: make(map[int]*websocket.ConnectionPool),
	}
}

// Websocket is WebsocketGateway implementation
type Websocket struct {
	mu                    sync.RWMutex
	userConnectionPoolMap map[int]*websocket.ConnectionPool
}

// EnqueueMessageBroadcast implementation
func (w *Websocket) EnqueueMessageBroadcast(messageID int, message string, fromUserID int) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for userID, userConnectionPool := range w.userConnectionPoolMap {
		userID := userID
		message := model.Message{
			ID:      messageID,
			Type:    model.TextMessage,
			Message: message,
			User: model.User{
				ID:   fromUserID,
				IsMe: userID == fromUserID,
			},
		}

		for connID, conn := range userConnectionPool.Slice() {
			log.Printf("[DEBUG] Writing to [User ID: %d][Conn ID: %d] at %d", userID, connID, time.Now().UnixNano())
			conn.Enqueue(message)
		}
	}
}

// RegisterConnection implementation
func (w *Websocket) RegisterConnection(userID int, connection contract.WebsocketConnection) (registrationID int) {
	w.mu.Lock()
	if _, ok := w.userConnectionPoolMap[userID]; !ok {
		w.userConnectionPoolMap[userID] = websocket.NewConnectionPool()
	}
	registrationID = w.userConnectionPoolMap[userID].Store(connection)
	connection.StartWorker()
	w.mu.Unlock()

	return
}

// UnregisterConnection implementation
func (w *Websocket) UnregisterConnection(userID int, registrationID int) {
	w.mu.Lock()
	defer w.mu.Unlock()

	userConnectionPool, ok := w.userConnectionPoolMap[userID]
	if !ok {
		return
	}

	connection := userConnectionPool.Get(registrationID)
	if connection == nil {
		return
	}

	connection.StopWorker()
	userConnectionPool.Delete(registrationID)
}
