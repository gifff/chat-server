package service

import (
	"github.com/gifff/chat-server/contract"
	"github.com/gifff/chat-server/interactor"
)

// NewChat returns chatService instance which satisfies ChatService contract
func NewChat(messageInteractor interactor.MessageInteractor, realtimeMessagingInteractor interactor.RealtimeMessagingInteractor) contract.ChatService {
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
