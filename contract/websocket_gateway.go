package contract

import "github.com/gifff/chat-server/websocket"

// WebsocketGateway adapter
type WebsocketGateway interface {
	EnqueueMessageBroadcast(messageID int, message string, fromUserID int)
	RegisterConnection(userID int, connection websocket.ConnectionDispatcher) (registrationID int)
	UnregisterConnection(userID int, registrationID int)
}
