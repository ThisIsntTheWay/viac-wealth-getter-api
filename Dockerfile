FROM golang:1.23.4-alpine

RUN apk add build-base

WORKDIR /go/src
COPY . .

RUN go mod download && \
    CGO_ENABLED=1 go build -o /go/bin/app

USER 1000

WORKDIR /app
CMD ["/go/bin/app"]
