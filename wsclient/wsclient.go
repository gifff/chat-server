package wsclient

import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	gorillaWebsocket "github.com/gorilla/websocket"
)

var ErrAlreadyListening = errors.New("client is already listening")

type ObserverFunc func(messageType int, packet []byte)

type Client struct {
	uri    string
	header http.Header
	ctx    context.Context

	dialer        gorillaWebsocket.Dialer
	conn          *gorillaWebsocket.Conn
	connMutex     sync.Mutex
	observers     []ObserverFunc
	obsMutex      sync.Mutex
	listening     bool
	stopChan      chan struct{}
	listenErrChan chan error
}

func NewClient(uri string, header http.Header) *Client {
	return NewClientContext(context.Background(), uri, header)
}

func NewClientContext(ctx context.Context, uri string, header http.Header) *Client {
	return &Client{
		ctx:    ctx,
		uri:    uri,
		header: header,

		stopChan:      make(chan struct{}, 1),
		listenErrChan: make(chan error, 1),
	}
}

// AttachObserver attaches an observer that will be notified when a
// message is consumed.
func (c *Client) AttachObserver(o ObserverFunc) {
	c.obsMutex.Lock()
	defer c.obsMutex.Unlock()

	if c.observers == nil {
		c.observers = make([]ObserverFunc, 0, 4)
	}

	c.observers = append(c.observers, o)
}

// Listen starts the connection to host, subscribe for messages through
// established connection and notify the attached observers.
//
// Listen returns errors channel in which internal listener pipes to
// when it finds an error.
// The error on second returned argument will not be nil when it fails
// to establish connection or ErrAlreadyListening if Listen has been
// called before.
func (c *Client) Listen() (<-chan error, error) {
	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	if c.listening {
		return nil, ErrAlreadyListening
	}

	err := c.connect()
	if err != nil {
		return nil, err
	}
	c.listening = true

	go c.listenMessages()

	return c.listenErrChan, nil
}

func (c *Client) connect() error {
	conn, resp, err := c.dialer.DialContext(c.ctx, c.uri, c.header)
	if err != nil {
		return err
	}
	c.conn = conn

	log.Printf("[DEBUG] response: code=%d", resp.StatusCode)
	defer resp.Body.Close()

	return nil
}

func (c *Client) listenMessages() {
listenerLoop:
	for {
		select {
		case <-c.stopChan:
			break listenerLoop
		default:
			msgType, pkt, err := c.conn.ReadMessage()
			if err != nil {

				c.listenErrChan <- err
				break
			}

			c.notifyObservers(msgType, pkt)
		}
	}
}

func (c *Client) notifyObservers(messageType int, packet []byte) {
	for _, o := range c.observers {
		go o(messageType, packet)
	}
}

// Stop signals the Client to cease listening and close the underlying
// connection.
func (c *Client) Stop(timeout time.Duration) error {
	if !c.listening {
		return nil
	}

	c.connMutex.Lock()
	defer c.connMutex.Unlock()

	c.listening = false
	c.stopChan <- struct{}{}

	return c.close(timeout)
}

func (c *Client) close(timeout time.Duration) error {
	err := c.conn.WriteControl(
		gorillaWebsocket.CloseMessage,
		gorillaWebsocket.FormatCloseMessage(gorillaWebsocket.CloseNormalClosure, ""),
		time.Now().Add(timeout))
	if err != nil {
		return err
	}

	return c.conn.Close()
}
