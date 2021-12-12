package websocket

// ConnectionDispatcher adapter
type ConnectionDispatcher interface {
	Dispatch(msg interface{})
	StartDispatcher()
	StopDispatcher()
}
