package di

import (
	"github.com/gifff/chat-server/gateway"
	"github.com/gifff/chat-server/interactor"
	"github.com/gifff/chat-server/pkg/server/handlers"
	"github.com/gifff/chat-server/service"
)

func InjectDependencies() {
	websocketGateway := gateway.NewWebsocket()

	messageInteractor := interactor.NewMessageInteractor()
	rtMessagingInteractor := interactor.NewRealtimeMessagingInteractor(websocketGateway)
	chatService := service.NewChat(
		messageInteractor,
		rtMessagingInteractor,
	)

	handlers.ChatService = chatService
	handlers.WebsocketGateway = websocketGateway
}
