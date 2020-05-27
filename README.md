# Chat Server

This is a sample project for playing with websocket with integration test built in it.

## How to test

Execute the following command:

```shell
$ go test -timeout 30s -parallel 10 ./...
```

> Note:
>
> The `-parallel 10` flag is necessary because the test uses parallel testing
> to simulate concurrent connections of 10 clients (9 consumers and 1 sender).

## Flow

1. Client connect via websocket to `/messages/listen` **(done)**
2. Publisher publishes Messages via HTTP endpoint `POST /messages`
3. Message will be stored in data store, and then pushed back to the connected clients

### Experimental / TODO

- [x] Enable Publisher to publish message by replying to the websocket
