package deps

import (
	"github.com/gifff/chat-server/contract"
	"github.com/gifff/chat-server/gateway"
	"github.com/gifff/chat-server/interactor"
	"github.com/gifff/chat-server/service"
)

func BuildDependencies() (websocketGateway contract.WebsocketGateway, chatService contract.ChatService) {
	websocketGateway = gateway.NewWebsocket()

	messageInteractor := interactor.NewMessageInteractor()
	rtMessagingInteractor := interactor.NewRealtimeMessagingInteractor(websocketGateway)
	chatService = service.NewChat(
		messageInteractor,
		rtMessagingInteractor,
	)

	return
}
