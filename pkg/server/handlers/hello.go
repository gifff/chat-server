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
	upgrader           = websocket.Upgrader{}
	userConnectionPool = make(map[int]*ConnectionPool)
	mutex              sync.RWMutex
	id                 int64 = 0 // TODO: make this reset-able for testing purpose
)

// Connection wraps the websocket.Conn with mutex
type Connection struct {
	mu              sync.Mutex
	conn            *websocket.Conn
	messageQueue    chan interface{}
	workerIsRunning bool
}

// NewConnection returns Connection and do the necessary initializations
func NewConnection(conn *websocket.Conn) *Connection {
	c := &Connection{
		conn: conn,
	}
	c.initQueue()

	return c
}

// Enqueue pushes message into internal message queue to be picked by the worker
func (c *Connection) Enqueue(msg interface{}) {
	c.messageQueue <- msg
}

func (c *Connection) initQueue() {
	if c.messageQueue == nil {
		c.messageQueue = make(chan interface{})
	}
}

// StartWorker spawns a goroutine which job is to pick message from internal message queue
// and write the message to the opened connection.
// Only one worker at a time can be running.
func (c *Connection) StartWorker() {
	if c.workerIsRunning {
		return
	}

	c.workerIsRunning = true
	go func() {
		for c.workerIsRunning {
			msg := <-c.messageQueue
			c.mu.Lock()
			c.conn.SetWriteDeadline(time.Now().Add(time.Second))
			_ = c.conn.WriteJSON(&msg)
			c.mu.Unlock()
		}
	}()
}

// StopWorker notifies the running worker to cease from working by toggling off the running flag
func (c *Connection) StopWorker() {
	c.workerIsRunning = false
}

// ConnectionPool holds pool of Connection with concurrent-safe handling for fast storing and deletion
type ConnectionPool struct {
	lastID int
	mu     sync.Mutex
	pool   map[int]*Connection
}

// NewConnectionPool returns ConnectionPool with prepared internal map
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		pool: make(map[int]*Connection),
	}
}

// Store assigns the Connection into internal pool with generated ID and return the ID for deletion
func (cp *ConnectionPool) Store(conn *Connection) int {
	cp.mu.Lock()
	id := cp.lastID
	cp.pool[id] = conn
	cp.lastID++
	cp.mu.Unlock()
	return id
}

// Delete deletes Connection from pool based on given pool map ID
func (cp *ConnectionPool) Delete(i int) {
	cp.mu.Lock()
	delete(cp.pool, i)
	cp.mu.Unlock()
}

// Slice flattens the internal pool map into slice of connections
func (cp *ConnectionPool) Slice() []*Connection {
	cp.mu.Lock()
	conns := make([]*Connection, len(cp.pool))
	i := 0
	for k := range cp.pool {
		conns[i] = cp.pool[k]
		i++
	}
	cp.mu.Unlock()
	return conns
}

func broadcast(msg string, senderID int) {
	atomic.AddInt64(&id, 1)

	mutex.RLock()
	for userID, wsConns := range userConnectionPool {

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

		for connID, c := range wsConns.Slice() {
			// Problem: message order synchronization
			// just because a goroutine for sending message A prior to message B is spawned first
			// there is no guarantee that goroutine A will be executed prior to goroutine B
			log.Printf("Writing to [User ID: %d][Conn ID: %d] at %d", userID, connID, time.Now().UnixNano())
			c.Enqueue(message)
		}
	}
	mutex.RUnlock()
}

// Hello is a websocket handler
func Hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	conn := NewConnection(ws)

	userID, _ := c.Get("user_id").(int)
	mutex.Lock()
	_, ok := userConnectionPool[userID]
	if !ok {
		userConnectionPool[userID] = NewConnectionPool()
	}
	connPoolID := userConnectionPool[userID].Store(conn)
	conn.StartWorker()
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		conn.StopWorker()
		userConnectionPool[userID].Delete(connPoolID)
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
