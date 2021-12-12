package handlers

import (
	"github.com/gifff/chat-server/chatservice"
	"github.com/gifff/chat-server/wsgateway"

	gorillaWebsocket "github.com/gorilla/websocket"
)

type Handlers struct {
	WSUpgrader       gorillaWebsocket.Upgrader // default value is ok
	WebsocketGateway wsgateway.WebsocketGateway
	ChatService      chatservice.ChatService
}
