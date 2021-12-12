package interactor

import (
	"github.com/gifff/chat-server/domain"
	"github.com/gifff/chat-server/wsgateway"
)

type RealtimeMessagingInteractor interface {
	DeliverMessage(message domain.Message)
}

func NewRealtimeMessagingInteractor(websocketGateway wsgateway.WebsocketGateway) RealtimeMessagingInteractor {
	return realtimeMessagingInteractor{websocketGateway}
}

type realtimeMessagingInteractor struct {
	websocketGateway wsgateway.WebsocketGateway
}

func (r realtimeMessagingInteractor) DeliverMessage(message domain.Message) {
	r.websocketGateway.EnqueueMessageBroadcast(message.ID, message.Message, message.UserID)
}
