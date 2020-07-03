package handlers

import (
	"github.com/labstack/echo"

	"github.com/gifff/chat-server/pkg/model"
)

// SendMessage handler
func SendMessage(c echo.Context) error {
	var reqBody model.Message
	err := c.Bind(&reqBody)
	if err != nil {
		return echo.ErrBadRequest
	}

	// reqBody.Message
	return nil
}
