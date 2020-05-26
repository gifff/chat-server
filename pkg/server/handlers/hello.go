package handlers

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
)

var (
	upgrader       = websocket.Upgrader{}
	connectionPool map[int]*websocket.Conn
)

// Hello is a websocket handler
func Hello(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	defer func() {
		ws.Close()
	}()

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error while reading message : %v\n", err)
		}

		log.Printf("Message: %s\n", msg)

		userID, _ := c.Get("user_id").(int)
		err = ws.WriteJSON(map[string]interface{}{
			"message": "reply",
			"user_id": userID,
		})
		if err != nil {
			return err
		}
	}
}
