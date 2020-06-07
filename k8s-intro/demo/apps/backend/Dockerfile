# builder
FROM golang:1.13-alpine as builder
WORKDIR /build

COPY ["./src", "src"]

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ./dist/todos ./src

# app
FROM scratch

COPY --from=builder ["/build/dist/todos", "/app/todos"]

ENTRYPOINT ["/app/todos"]
