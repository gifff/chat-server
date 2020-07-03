package main

import (
	"fmt"

	"github.com/gifff/chat-server/pkg/di"
	"github.com/gifff/chat-server/pkg/server"

	"github.com/labstack/echo"
)

func main() {
	di.InjectDependencies()

	s := server.New(echo.New(), ":8080")
	s.Start()
	fmt.Println("Chat Server is started")
}
