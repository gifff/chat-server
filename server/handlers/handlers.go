package handlers

import (
	"github.com/gifff/chat-server/contract"

	gorillaWebsocket "github.com/gorilla/websocket"
)

type Handlers struct {
	WSUpgrader       gorillaWebsocket.Upgrader // default value is ok
	WebsocketGateway contract.WebsocketGateway
	ChatService      contract.ChatService
}
