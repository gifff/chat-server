package contract

// WebsocketConnection adapter
type WebsocketConnection interface {
	Enqueue(msg interface{})
	StartWorker()
	StopWorker()
}
