package deps

import (
	"github.com/gifff/chat-server/chatservice"
	"github.com/gifff/chat-server/interactor"
	"github.com/gifff/chat-server/wsgateway"
)

func BuildDependencies() (websocketGateway wsgateway.WebsocketGateway, chatService chatservice.ChatService) {
	websocketGateway = wsgateway.New()

	messageInteractor := interactor.NewMessageInteractor()
	rtMessagingInteractor := interactor.NewRealtimeMessagingInteractor(websocketGateway)
	chatService = chatservice.NewService(
		messageInteractor,
		rtMessagingInteractor,
	)

	return
}
