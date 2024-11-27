FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY ./song-service/go.mod ./song-service/go.sum ./
COPY ./utils /utils
RUN go mod download

COPY ./song-service .
RUN go build -o song-service .

RUN go install github.com/air-verse/air@latest

CMD ["air"]