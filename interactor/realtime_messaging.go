package interactor

import (
	"github.com/gifff/chat-server/contract"
	"github.com/gifff/chat-server/domain"
)

type RealtimeMessagingInteractor interface {
	DeliverMessage(message domain.Message)
}

func NewRealtimeMessagingInteractor(websocketGateway contract.WebsocketGateway) RealtimeMessagingInteractor {
	return realtimeMessagingInteractor{websocketGateway}
}

type realtimeMessagingInteractor struct {
	websocketGateway contract.WebsocketGateway
}

func (r realtimeMessagingInteractor) DeliverMessage(message domain.Message) {
	r.websocketGateway.EnqueueMessageBroadcast(message.ID, message.Message, message.UserID)
}
