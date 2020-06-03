FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git gcc g++ make libc-dev pkgconfig zeromq-dev curl libunwind-dev

RUN adduser -D -u 1001 -g '' appuser
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .

RUN go env CGO_ENABLED
RUN go mod download
RUN go mod verify

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/indexerd ./cmd/indexerd

RUN chmod u+x /go/bin/*

FROM alpine:latest

RUN apk update && apk add --no-cache zeromq

WORKDIR /app

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/indexerd /app/indexerd

COPY .env.dist /app/.env

COPY ./config/mappings /app/mappings

ENTRYPOINT ["/app/indexerd"]