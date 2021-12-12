package model

// User data model
type User struct {
	ID   int    `json:"id"`
	Name string `json:"string"`
	IsMe bool   `json:"is_me"`
}

// Message data model
type Message struct {
	ID      int         `json:"id"`
	Type    MessageType `json:"type"`
	Message string      `json:"message"`
	User    User        `json:"user"`
}

// MessageType enum type
type MessageType int

const (
	// UnknownMessage message type
	UnknownMessage MessageType = iota
	// TextMessage message type
	TextMessage
	// RetractMessage message type
	RetractMessage
)
