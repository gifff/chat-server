package server_test

import (
	"net/http"
	"testing"

	"github.com/labstack/echo"
	"github.com/posener/wstest"

	"github.com/gifff/chat-server/pkg/server"
)

type message struct {
	Message string `json:"message"`
	UserID  int    `json:"user_id"`
}

func TestHelloHandler(t *testing.T) {
	var err error

	e := echo.New()
	_ = server.New(e, "")

	requestHeader := http.Header{}
	requestHeader.Set("X-User-Id", "1337")

	d := wstest.NewDialer(e)
	c, resp, err := d.Dial("ws://whatever/messages/listen", requestHeader)
	if err != nil {
		t.Fatal(err)
	}

	defer c.Close()

	if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
		t.Errorf("resp.StatusCode = %q, want %q", got, want)
	}

	err = c.WriteJSON(map[string]string{"message": "Hello!"})
	if err != nil {
		t.Fatal(err)
	}

	var msg message
	err = c.ReadJSON(&msg)
	if err != nil {
		t.Fatal(err)
	}

	wantedMsg := message{Message: "reply", UserID: 1337}
	if msg != wantedMsg {
		t.Errorf("message: got = %+v, want %+v", msg, wantedMsg)
	}
}
