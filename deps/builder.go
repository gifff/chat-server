package deps

import (
	"github.com/gifff/chat-server/contract"
	"github.com/gifff/chat-server/interactor"
	"github.com/gifff/chat-server/service"
	"github.com/gifff/chat-server/wsgateway"
)

func BuildDependencies() (websocketGateway wsgateway.WebsocketGateway, chatService contract.ChatService) {
	websocketGateway = wsgateway.New()

	messageInteractor := interactor.NewMessageInteractor()
	rtMessagingInteractor := interactor.NewRealtimeMessagingInteractor(websocketGateway)
	chatService = service.NewChat(
		messageInteractor,
		rtMessagingInteractor,
	)

	return
}
