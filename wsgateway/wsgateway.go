package wsgateway

import (
	"log"
	"sync"
	"time"

	"github.com/gifff/chat-server/model"
	"github.com/gifff/chat-server/websocket"
)

// New returns Websocket object which satisfies the WebsocketGateway contract
func New() WebsocketGateway {
	return &wsGateway{
		userConnectionPoolMap: make(map[int]*websocket.ConnectionPool),
	}
}

// wsGateway is WebsocketGateway implementation
type wsGateway struct {
	mu                    sync.RWMutex
	userConnectionPoolMap map[int]*websocket.ConnectionPool
}

// EnqueueMessageBroadcast implementation
func (w *wsGateway) EnqueueMessageBroadcast(messageID int, message string, fromUserID int) {
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
			conn.Dispatch(message)
		}
	}
}

// RegisterConnection implementation
func (w *wsGateway) RegisterConnection(userID int, connection websocket.ConnectionDispatcher) (registrationID int) {
	w.mu.Lock()
	if _, ok := w.userConnectionPoolMap[userID]; !ok {
		w.userConnectionPoolMap[userID] = websocket.NewConnectionPool()
	}
	registrationID = w.userConnectionPoolMap[userID].Store(connection)
	connection.StartDispatcher()
	w.mu.Unlock()

	return
}

// UnregisterConnection implementation
func (w *wsGateway) UnregisterConnection(userID int, registrationID int) {
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

	connection.StopDispatcher()
	userConnectionPool.Delete(registrationID)
}

func (w *wsGateway) TotalConnections() int {
	n := 0
	for _, connPool := range w.userConnectionPoolMap {
		n += connPool.Size()
	}

	return n
}
