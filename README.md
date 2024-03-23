## 👏👏👏 Clapper

This is a simple implementation of using WebSocket connections in Go, sending an update every second to the client with a human-friendly display of `xxd:xxh:xxm:xx.xs`.

On receiving a `PUT` to the `/reset` endpoint, the clapper will be reset to `0`.

Uses [gorilla/websocket](https://github.com/gorilla/websocket)

View the [demo](https://clapper.jarv.org)

## Local development

By default it runs as a single binary

```
mise install
./bin/build-assets  # for htmx/tailwind
go run !(*_test).go
```

Optionally you can specify a file to persist the counter

```
go run !(*_test).go -fname /path/to/file.
```
