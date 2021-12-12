package handlers

import (
	"log"

	"github.com/labstack/echo"

	"github.com/gifff/chat-server/pkg/model"
	"github.com/gifff/chat-server/websocket"
)

// MessageListener is a websocket handler
func (h *Handlers) MessageListener(c echo.Context) error {
	ws, err := h.WSUpgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	userID, _ := c.Get("user_id").(int)
	conn := websocket.NewConnectionDispatcher(ws)
	registrationID := h.WebsocketGateway.RegisterConnection(userID, conn)

	defer func() {
		h.WebsocketGateway.UnregisterConnection(userID, registrationID)
		ws.Close()
	}()

	for {
		var msg model.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("[DEBUG] Error while reading message [userID: %d] : %v\n", userID, err)
			break
		}

		log.Printf("[DEBUG] Message [userID: %d]: %+v\n", userID, msg)

		if msg.Type != model.UnknownMessage {
			h.ChatService.SendMessage(msg.Message, userID)
		}
	}

	return nil
}
