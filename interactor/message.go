package interactor

import (
	"github.com/gifff/chat-server/domain"
)

type MessageInteractor interface {
	Create(message string, userID int) domain.Message
}

func NewMessageInteractor() MessageInteractor {
	return &messageInteractor{
		nextMessageID: 1,
	}
}

type messageInteractor struct {
	nextMessageID int
}

func (m *messageInteractor) Create(message string, userID int) domain.Message {
	msg := domain.Message{
		ID:      m.nextMessageID,
		Message: message,
		UserID:  userID,
	}

	m.nextMessageID++

	return msg
}
