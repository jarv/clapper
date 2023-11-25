# syntax = docker/dockerfile:1-experimental

FROM golang:1.21 as wscnt-builder
WORKDIR /app
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
  CGO_ENABLED=0 go build -ldflags "-w" -o wscnt ./app.go ./store.go ./counter.go

FROM scratch
COPY --from=wscnt-builder /etc/passwd /etc/passwd
COPY --from=wscnt-builder /app/wscnt /
COPY --from=wscnt-builder /app/index.html /

ENTRYPOINT ["/wscnt", "-addr=:8710"]
