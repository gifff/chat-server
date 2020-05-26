package main

import (
	"fmt"

	"github.com/gifff/chat-server/pkg/server"

	"github.com/labstack/echo"
)

func main() {
	s := server.New(echo.New(), ":8080")
	s.Start()
	fmt.Println("Chat Server is started")
}
