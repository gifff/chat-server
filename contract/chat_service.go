package contract

// ChatService contract
type ChatService interface {
	SendMessage(message string, fromUserID int)
}
