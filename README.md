# Chat Server

This is a sample project for playing with websocket with integration test built in it.

## Flow

1. Client connect via websocket to `/messages/listen`
2. Publisher publishes Messages via HTTP endpoint `POST /messages`
3. Message will be stored in data store, and then pushed back to the connected clients

### Experimental / TODO

- Enable Publisher to publish message by replying to the websocket
