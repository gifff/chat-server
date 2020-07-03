package contract

// WebsocketGateway adapter
type WebsocketGateway interface {
	EnqueueMessageBroadcast(messageID int, message string, fromUserID int)
	RegisterConnection(userID int, connection WebsocketConnection) (registrationID int)
	UnregisterConnection(userID int, registrationID int)
}
