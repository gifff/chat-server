package chatservice

import (
	"github.com/gifff/chat-server/interactor"
)

// ChatService contract
type ChatService interface {
	SendMessage(message string, fromUserID int)
}

// NewService returns chatService instance which satisfies ChatService interface
func NewService(messageInteractor interactor.MessageInteractor, realtimeMessagingInteractor interactor.RealtimeMessagingInteractor) ChatService {
	return chatService{
		messageInteractor:           messageInteractor,
		realtimeMessagingInteractor: realtimeMessagingInteractor,
	}
}

type chatService struct {
	messageInteractor           interactor.MessageInteractor
	realtimeMessagingInteractor interactor.RealtimeMessagingInteractor
}

// SendMessage implementation
func (c chatService) SendMessage(message string, fromUserID int) {
	msg := c.messageInteractor.Create(message, fromUserID)
	c.realtimeMessagingInteractor.DeliverMessage(msg)
}
