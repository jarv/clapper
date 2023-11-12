## wscnt

This is a simple example of implementing a websocket connection in Go, sending an update every one-tenth of a second to the client with a human-friendly display of `xxd:xxh:xxm:xx.xs`.

On receiving a `PUT` to the `/reset` endpoint, the counter will be reset to `0`.

Uses [gorilla/websocket](https://github.com/gorilla/websocket), see also [websocket examples](https://github.com/gorilla/websocket/tree/main/examples) for other very simple websocket implementations.

View the [demo](https://like.jarv.org)

## Local development

```
go run !(*_test).go
```

## Docker

```
docker-compose build
docker-compose up
```
