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

func closeConnection(c *websocket.Conn) error {
	err := c.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		time.Now().Add(time.Second))
	if err != nil {
		return err
	}

	return c.Close()
}

func TestHelloHandler(t *testing.T) {
	e := echo.New()
	_ = server.New(e, "")

	senderOutgoingMessages := []model.Message{
		{
			Message: "hello",
			Type:    model.TextMessage,
		},
		{
			Message: "hello2",
			Type:    model.TextMessage,
		},
	}

	senderExpectedIncomingMessages := []model.Message{
		{
			ID:      1,
			Message: "hello",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: true,
			},
		},
		{
			ID:      2,
			Message: "hello2",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: true,
			},
		},
	}

	clientExpectedIncomingMessages := []model.Message{
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

			defer closeConnection(c)

			if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
				t.Errorf("resp.StatusCode = %q, want %q", got, want)
			}

			var msg model.Message

			for i, expectedMessage := range clientExpectedIncomingMessages {
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

		defer closeConnection(c)

		if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
			t.Errorf("resp.StatusCode = %q, want %q", got, want)
		}

		for i := range senderOutgoingMessages {
			outgoingMsg := senderOutgoingMessages[i]

			err = c.WriteJSON(&outgoingMsg)
			if err != nil {
				t.Fatal(err)
			}
		}

		for i := range senderExpectedIncomingMessages {
			expectedIncomingMsg := senderExpectedIncomingMessages[i]

			var incomingMsg model.Message
			err = c.ReadJSON(&incomingMsg)
			if err != nil {
				t.Fatal(err)
			}

			if incomingMsg != expectedIncomingMsg {
				t.Errorf("incoming message: got = %+v, want %+v", incomingMsg, expectedIncomingMsg)
			}
		}
	})
}
func TestHelloHandlerMultipleConnectionPerClient(t *testing.T) {
	e := echo.New()
	_ = server.New(e, "")

	senderOutgoingMessages := []model.Message{
		{
			Message: "hello",
			Type:    model.TextMessage,
		},
		{
			Message: "hello2",
			Type:    model.TextMessage,
		},
	}

	senderExpectedIncomingMessages := []model.Message{
		{
			ID:      3,
			Message: "hello",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: true,
			},
		},
		{
			ID:      4,
			Message: "hello2",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: true,
			},
		},
	}

	clientExpectedIncomingMessages := []model.Message{
		{
			ID:      3,
			Message: "hello",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: false,
			},
		},
		{
			ID:      4,
			Message: "hello2",
			Type:    model.TextMessage,
			User: model.User{
				ID:   1337,
				Name: "",
				IsMe: false,
			},
		},
	}

	numberOfClients := 4
	wg := sync.WaitGroup{}
	wg.Add(numberOfClients)
	for i := 1; i <= numberOfClients; i++ {

		userID := 100

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

			defer closeConnection(c)

			if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
				t.Errorf("resp.StatusCode = %q, want %q", got, want)
			}

			var msg model.Message

			for i, expectedMessage := range clientExpectedIncomingMessages {
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

		defer closeConnection(c)

		if got, want := resp.StatusCode, http.StatusSwitchingProtocols; got != want {
			t.Errorf("resp.StatusCode = %q, want %q", got, want)
		}

		for i := range senderOutgoingMessages {
			outgoingMsg := senderOutgoingMessages[i]

			err = c.WriteJSON(&outgoingMsg)
			if err != nil {
				t.Fatal(err)
			}
		}

		for i := range senderExpectedIncomingMessages {
			expectedIncomingMsg := senderExpectedIncomingMessages[i]

			var incomingMsg model.Message
			err = c.ReadJSON(&incomingMsg)
			if err != nil {
				t.Fatal(err)
			}

			if incomingMsg != expectedIncomingMsg {
				t.Errorf("incoming message: got = %+v, want %+v", incomingMsg, expectedIncomingMsg)
			}
		}
	})
}
