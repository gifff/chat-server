package websocket

import (
	"sync"
	"time"

	gorillaWebsocket "github.com/gorilla/websocket"

	"github.com/gifff/chat-server/contract"
)

// Connection wraps the websocket.Conn with mutex and satisfies the WebsocketConnection contract
type Connection struct {
	mu              sync.Mutex
	conn            *gorillaWebsocket.Conn
	messageQueue    chan interface{}
	stopSignal      chan struct{}
	workerIsRunning bool
}

// NewConnection returns Connection and do the necessary initializations
func NewConnection(conn *gorillaWebsocket.Conn) contract.WebsocketConnection {
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
	if c.stopSignal == nil {
		c.stopSignal = make(chan struct{})
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
			// use channel for stop signal instead of checking for c.workerIsRunning
			// it is because that if the toggle is off, the msg := <-c.messageQueue is still
			// waiting for incoming message and will be able to send one last message even
			// after StopWorker() is invoked.
			// simply put, StopWorker() does not immediately kill the worker goroutine
			select {
			case msg := <-c.messageQueue:
				c.mu.Lock()
				c.conn.SetWriteDeadline(time.Now().Add(time.Second))
				_ = c.conn.WriteJSON(&msg)
				c.mu.Unlock()
			case <-c.stopSignal:
				c.mu.Lock()
				c.workerIsRunning = false
				c.mu.Unlock()
				break
			}
		}
	}()
}

// StopWorker notifies the running worker to cease from working by toggling off the running flag
func (c *Connection) StopWorker() {
	if c.workerIsRunning {
		c.stopSignal <- struct{}{}
	}
}
