package contract

// WebsocketConnection adapter
type WebsocketConnection interface {
	Dispatch(msg interface{})
	StartDispatcher()
	StopDispatcher()
}
