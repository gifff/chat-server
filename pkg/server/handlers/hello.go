package handlers

import (
	"log"

	gorillaWebsocket "github.com/gorilla/websocket"
	"github.com/labstack/echo"

	"github.com/gifff/chat-server/pkg/model"
	"github.com/gifff/chat-server/pkg/websocket"
)

var (
	upgrader = gorillaWebsocket.Upgrader{}
)

// Hello is a websocket handler
func Hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	userID, _ := c.Get("user_id").(int)
	conn := websocket.NewConnection(ws)
	registrationID := WebsocketGateway.RegisterConnection(userID, conn)

	defer func() {
		WebsocketGateway.UnregisterConnection(userID, registrationID)
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
			ChatService.SendMessage(msg.Message, userID)
		}
	}

	return nil
}
