package server_test

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/posener/wstest"

	"github.com/gifff/chat-server/pkg/model"
	"github.com/gifff/chat-server/pkg/server"
)

func TestHelloHandler(t *testing.T) {
	e := echo.New()
	_ = server.New(e, "")

	expectedMessages := []model.Message{
		{
			ID:      1,
			Message: "hello",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: false,
			},
		},
		{
			ID:      2,
			Message: "hello2",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: false,
			},
		},
	}

	/*
		NOTE:
		Must run the test with flag -parallel <number of clients + 1>
		It is because the test is parallel with the number of clienst and the sender itself.
		Otherwise, the test will fail because some of the clients won't get the message from the sender
		and it will cause locking.
	*/
	numberOfClients := 9
	wg := sync.WaitGroup{}
	wg.Add(numberOfClients)
	for i := 1; i <= numberOfClients; i++ {

		userID := i * 100

		t.Run(fmt.Sprintf("client_%d", userID), func(t *testing.T) {
			t.Parallel()

			requestHeader := http.Header{}
			requestHeader.Set("X-User-Id", strconv.Itoa(userID))

			d := wstest.NewDialer(e)
			c, resp, err := d.Dial("ws://whatever/messages/listen", requestHeader)
			if err != nil {
				t.Fatal(err)
			}

			wg.Done()

			defer func() {
				c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
				c.Close()
			}()

			if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
				t.Errorf("resp.StatusCode = %q, want %q", got, want)
			}

			var msg model.Message

			for i, expectedMessage := range expectedMessages {
				err = c.ReadJSON(&msg)
				if err != nil {
					t.Fatal(err)
				}

				if msg != expectedMessage {
					t.Errorf("message [%d]: got = %+v, want %+v", i, msg, expectedMessage)
				}
			}
		})
	}

	t.Run("sender", func(t *testing.T) {
		t.Parallel()

		wg.Wait()

		requestHeader := http.Header{}
		requestHeader.Set("X-User-Id", "1337")

		d := wstest.NewDialer(e)
		c, resp, err := d.Dial("ws://whatever/messages/listen", requestHeader)
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			c.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second))
			c.Close()
		}()

		if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
			t.Errorf("resp.StatusCode = %q, want %q", got, want)
		}

		err = c.WriteJSON(map[string]interface{}{
			"message": "hello",
			"type":    model.TextMessage,
		})
		if err != nil {
			t.Fatal(err)
		}

		var msg model.Message
		err = c.ReadJSON(&msg)
		if err != nil {
			t.Fatal(err)
		}

		wantedMsg := model.Message{
			ID:      1,
			Message: "hello",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: true,
			},
		}
		if msg != wantedMsg {
			t.Errorf("message: got = %+v, want %+v", msg, wantedMsg)
		}

		err = c.WriteJSON(map[string]interface{}{
			"message": "hello2",
			"type":    model.TextMessage,
		})
		if err != nil {
			t.Fatal(err)
		}

		err = c.ReadJSON(&msg)
		if err != nil {
			t.Fatal(err)
		}

		wantedMsg = model.Message{
			ID:      2,
			Message: "hello2",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: true,
			},
		}
		if msg != wantedMsg {
			t.Errorf("message: got = %+v, want %+v", msg, wantedMsg)
		}
	})
}
